---
name: depin-prod-admin-read
description: "Live prod Numo/depin stats through admin REST reads and native approved DB reads."
version: 1.5.0
metadata:
  hermes:
    tags: [numo, depin, production, prod, admin, read-only, user-stats, users, submissions, api, db-read]
    related_skills: [github-auth, native-mcp]
---

# Depin Prod Admin Read

Use this skill when a Story request asks for live Numo/depin user stats, submission stats, admin user lookup, or production depin API route context.

## Runtime Contract

- Use `DEPIN_ADMIN_BASE_URL` as the base URL. Stage RSI intentionally points this at the production public endpoint, `https://depin.storyprotocol.net`.
- Use `DEPIN_ADMIN_READ_API_KEY_HEADER` for the header name and `DEPIN_ADMIN_READ_API_KEY` for the header value.
- Use `User-Agent: rsi-hermes-company-computer/1.0` and `X-RSI-Source: hermes` on direct prod depin REST reads so Cloudflare can distinguish the company-computer path from generic scripted traffic. These headers are not credentials; origin authorization still depends on the admin read key.
- Never print, summarize, export, or store credential values. Only report whether a credential is present.
- Treat the admin REST credential as read-only. Do not use it for write, mutation, delete, or admin management actions.

## Choose The Read Path

There are two separate read paths. Pick one deliberately and do not mix their results unless the user asks for comparison.

- **Admin REST path**: use this first for supported aggregate depin admin endpoints such as `/v1/admin/stats/*`, `/v1/admin/cohorts/languages`, and `/v1/admin/overview`. It uses `DEPIN_ADMIN_BASE_URL` plus the mounted admin read key. It is best for predefined product metrics and fast aggregate stats.
- **Native DB read path**: use Hermes native `db_read.*` tools only when REST endpoints cannot answer the exact question, such as table-level counts, arbitrary filters, joins, or validating data semantics directly in PostgreSQL. The query is an external permissioned tool call: Hermes pauses at `db_read.query`, RSI posts an approval card, an authorized admin approves or denies the exact SQL, RSI executes through the DB-read worker, and Hermes resumes with the sanitized result to produce the final answer.

For language-related counts, be precise about semantics. `/v1/admin/cohorts/languages` groups by `users.primary_language`; direct SQL against `scripts.language_code` counts transcript/script records. These are different questions and can return different numbers.

For per-transcript distribution queries (e.g., "how many unique users submitted to each Vietnamese script"), see `references/vi-transcript-queries.md` for the full join pattern, histogram query, and cross-validation technique. For per-campaign multi-language transcript distribution queries, the same file now includes Query 4 (per-campaign histogram) which generalizes the pattern to all active campaigns.

## Country Attribution

When the user asks for breakdowns by country (nationality, geographic distribution), there are two data sources with different trust models:

### Self-Reported (`users.nationality`)

The `users.nationality` column stores whatever the user declared at signup. It is available for all users (0 NULLs observed), but **can be tricked** — users may claim a different country via VPN, proxy, or outright false declaration. Use this for broad-stroke analysis where coverage matters more than precision.

### Castle IP Country (`castle_risk_events.ip_country_code`)

Castle's risk events capture the user's IP-based country at event time. This is harder to spoof and should be **preferred** when the user asks for country data that can't be gamed. However, Castle data only exists for users who triggered a risk event — coverage is ~9,400 users (vs 298K total submissions). For users without Castle events, fall back to self-reported nationality.

**Preferred query pattern** (Castle IP first, self-reported fallback):

```sql
WITH user_castle_country AS (
  SELECT DISTINCT ON (user_id) user_id, ip_country_code AS castle_country
  FROM castle_risk_events
  WHERE ip_country_code IS NOT NULL AND ip_country_code != ''
  ORDER BY user_id, created_at DESC
),
submission_country AS (
  SELECT s.id, s.campaign_id, s.user_id,
    COALESCE(uc.castle_country, u.nationality, 'UNKNOWN') AS effective_country,
    CASE WHEN uc.castle_country IS NOT NULL THEN 'castle'
         WHEN u.nationality IS NOT NULL THEN 'self_reported'
         ELSE 'unknown' END AS country_source
  FROM submissions s
  JOIN users u ON s.user_id = u.id
  LEFT JOIN user_castle_country uc ON u.id = uc.user_id
)
SELECT campaign_id, effective_country AS country,
       count(*) AS submission_count,
       count(DISTINCT user_id) AS unique_users
FROM submission_country
GROUP BY campaign_id, effective_country
ORDER BY campaign_id, submission_count DESC;
```

**Key findings (2026-05-08)**:
- Castle coverage: 9,424 users; 90.2% match self-reported, 5.8% mismatch, 3.9% had NULL self-reported
- Top mismatch patterns: VN IPs claiming UK (50 users), US IPs claiming VN (47), NL IPs claiming NG (14)
- Biggest impact: Tamil campaign — US goes from 1.8% (self-reported) to 11.2% (Castle IP), revealing ~3,600 submissions from US IPs that claimed other nationalities
- Nigeria: 819 unique Castle IP users vs only 219 self-reported — ~600 users with Nigerian IPs hiding their origin
- New countries emerge through Castle: Japan, Saudi Arabia, Kyrgyzstan, Lithuania, Serbia

**`submission_quality_country_daily.castle_ip_country_code` is UNUSABLE.** All values in production are `'UNKNOWN'` — this field is not populated. The only viable Castle IP source is `castle_risk_events`.

### Two Dimensions: Submission Count ≠ Unique Users

When the user asks about "submissions from Country X," always consider both dimensions:
- **Submission count**: raw volume (sensitive to farming — one user submitting 100+ times)
- **Unique user count**: distinct participants (better signal for genuine adoption)

These can diverge dramatically. Nigeria has only 755 submissions (rank ~#16) but 219 unique self-reported users (rank #4 globally). Conversely, Poland has 2,589 submissions (rank #6) but only 25 unique users (103.6 subs/user — farming pattern). Present both when they diverge significantly.

Full reference queries and case studies: `references/country-breakdown-queries.md`.

## Native DB Read Rules

- Use `db_read.sources` to list available DB targets.
- Use `db_read.schema` to inspect allowlisted tables and columns for a target.
- Use `db_read.validate` when you need repair feedback before submitting an approval request.
- Use `db_read.query` exactly once for the SQL you intend to run. The tool must show the exact SQL to an authorized admin before execution.
- **CRITICAL**: `db_read.query` may return `"db-read query did not create an external tool pause"` — this error is often a false negative. The RSI platform posts the Slack approval card **asynchronously** while Hermes receives the error synchronously. Each apparent failure call actually creates a new approval card in the thread. **Never loop-retry db_read.query** — check the Slack thread for approval cards instead.
- After `db_read.query` is called, **check the Slack thread immediately** before pursuing fallbacks. If approval cards appeared, the queries are live — use `db_read.status` with each request ID to retrieve the full result data (including the `result_sample` rows that Slack cards summarize).
- **`db_read.status` frequently returns 404** even for valid, succeeded queries. Do not loop-retry `db_read.status` — it is not a reliable diagnostic. When repeated 404s occur, switch to the Slack mirror (`mcp_rsi_task_trace_*_conversation_get` on the thread) to read approval card statuses (approved/executing/succeeded/failed). The approval cards are the ground truth; `db_read.status` is a best-effort convenience.
- **When `db_read.query` returns `"[Result unavailable — see context summary above]"`**: the query was approved and executed, but the sanitized row data was not propagated to the Hermes context. Check the Slack mirror for the approval card status (`succeeded; rows=N truncated=false`). If you need the actual row values, look for `*Result:*` or `*Sample:*` blocks in the approval cards on Slack; not all approval card templates include full result samples.
- Use `session_search` before diving deep: if a prior session tackled the same task, its summary reveals what queries worked, what data was found, and what pitfalls were hit. This avoids re-inventing the schema exploration path.
- When `db_read.query` results fail to propagate, see `references/rsidb-propagation-architecture.md` for the full RSI platform internal mechanism — the 3-layer async pipeline, specific file/line references in `piplabs/rsi-agent-platform`, and known failure modes.
- When targeting `depin-prod`, note that prod queries route through AWS Lambda and require Slack approval. Stage queries (`depin-stage`) route through the stage worker pod directly and have a 5s timeout. For distribution queries that fit within caps, both work; stage may return slightly stale data but avoids the Lambda relay hop.
- Do not use terminal, Kubernetes, or hand-built network calls to bypass the native permission flow.
- Do not self-approve. The approval is for an authorized DB-read admin, not necessarily the requester.
- The approval card is audit UI. The resumed Hermes run owns the final user-facing answer in the original Slack thread.
- If the DB result is denied, expired, or fails, use the resumed tool result to explain the blocker and propose a safe next query.

## Pitfalls

- **`execute_code` Python env is blind to `DEPIN_*` vars.** The Python subprocess spawned by `execute_code` runs in a limited environment that does not inherit the admin read key env vars. Use terminal only to confirm credential presence without printing values. If `execute_code` reports all vars as missing but terminal shows them present, trust the shell output.
- **Admin read key is stats-scoped only.** `DEPIN_ADMIN_READ_API_KEY` authorizes endpoints gated by `AdminReadOnlyAccess` (including `/v1/admin/stats/*`, `/v1/admin/cohorts/languages`, and `/v1/admin/overview`) but does not work for endpoints gated by `AdminAccess` such as `/v1/admin/users/**` and `/v1/admin/cohorts/demographics`. Those endpoints require JWT. A `401` from an `AdminAccess` endpoint is expected and should be reported as an auth scope mismatch, not a credential failure.
- **`db_read.status` is hit-or-miss.** It frequently returns `404 "db read request not found"` even for requests that are confirmed "succeeded" in the Slack approval cards. Do not loop on it. After one 404, fall back to the Slack mirror. The approval cards in the thread are the authoritative source for request state, not `db_read.status`.
- **When `db_read.query` result is invisible**, the sanitized row data may not propagate back to the Hermes context after resume. The Slack approval card shows the status line (`succeeded; rows=N truncated=false`) but may or may not include a `*Result:*` sample block. For distribution queries, the histogram query (Query 3 in `references/vi-transcript-queries.md`) returns compact results (one row per bucket) that fit in a Slack card.
- **`session_search` finds prior work.** When a depin data task looks familiar, search for it. Prior sessions may have already discovered the schema, the right query, and the data — saving multiple approval cycles.
- **send_message is trace-scoped idempotent: only one message lands per trace.** The RSI gateway uses a fixed idempotency key `{channel}:{thread_ts}:trace-{trace_id}` for all `send_message` calls in one trace. The second call deduplicates to the first silently. **Never send a placeholder or "stand by" message first.** Collect all data, build the complete answer, then send exactly one message with all content at the end. If a placeholder was already sent, accept that no further messages will land and provide the full answer in the Hermes conversation instead.
- **`/v1/admin/stats/submissions` is dimension-blind.** It returns only `[{date, count}]` with no language, nationality, campaign, or state filter. For "Vietnamese submissions", always try `/v1/admin/cohorts/languages?range=all` first. Only use DB read if the user asks for script/transcript-level semantics or campaign-language semantics that REST cannot answer.
- **`/v1/admin/submissions` is cursor-paginated with no total.** The response contains only `{items, next_cursor}`. Prefer aggregate endpoints or DB read for exact counts when pagination would be slow or ambiguous.
- **`cohorts/languages` uses `users.primary_language`, not campaign language.** For campaign-scoped language counts, find the campaign by `supported_languages` in `/v1/admin/campaigns`, then paginate `/v1/admin/submissions?campaign_id=X` or use native DB read when exact SQL is appropriate. Full enumeration technique and script: `references/campaign-language-filtering.md`.
- **`range=1d` on cohorts/languages may return empty.** The `1d` range has been observed returning `items: []` (zero items for all languages) even when `7d`, `30d`, and `all` return populated results. This may be a UTC day boundary issue — the window may require a full calendar day to have elapsed. When `1d` returns empty, fall back to `7d` or `all` and note the gap in the response.
- **Submission counts can fluctuate rapidly during active ingestion.** Per-language submission counts on cohorts/languages may shift significantly within hours — observed jumps of 3× to 10× for a single language in one session. When answering the same question across multiple traces, always re-query the live API rather than relying on the prior trace's numbers. Note any growth since prior runs so the user understands the metric is volatile.
- **`campaigns` table uses `campaign_name` and `campaign_type`, not `name` and `type`.** The documentation in `docs/architecture/database.md` uses `name` in the table description, but the live production column is `campaign_name`. Always verify column names with `information_schema.columns` before joining campaigns to other tables — a `SELECT column_name FROM information_schema.columns WHERE table_name = 'campaigns'` is cheap and avoids a wasted approval cycle.
- **Campaign scripts are bulk-loaded — expect large counts.** Active campaigns may have 350K+ scripts each. The histogram query (group by unique_user count after the DISTINCT subquery) compresses this into a handful of rows and fits easily within the 100-row cap. Direct per-script listing will hit the cap.

## Source Of Truth

- Public API discovery: the deployed production OpenAPI document is intentionally client-facing and must not be treated as authoritative for internal admin stats routes.
- Internal admin stats contract: `piplabs/depin-backend`, especially `apps/api/src/http/routes/admin.rs`, `apps/api/src/http/extractors.rs`, `apps/api/src/services/admin_dashboard.rs`, and `docs/api-workflows.md`.
- Public DNS/WAF routing: `piplabs/cloudflare`, especially `src/zones/storyprotocol.net/records.ts` and `src/zones/storyprotocol.net/waf.ts`.
- Deployment and Vault wiring: `story-deployments`, `rsi-platform/rsi-agent-platform/use1-stage.yaml`, and `story/depin-backend/use1-prod.yaml`.
- Runtime readiness: the `story-deployments` depin admin read validation hook curls `/v1/admin/stats/user-growth` from the stage RSI cluster with the mounted read key and must fail loudly on non-200 responses.

## Query Pattern

0. **Search for prior work.** Run `session_search` with a query describing the task before diving into schema exploration or writing SQL. Prior sessions may have already mapped the schema, written the query, and found the answer.
1. Confirm `DEPIN_ADMIN_READ_API_KEY` is present without printing its value.
2. For public/client-facing route context, prefer the checked-in `piplabs/depin-backend` OpenAPI/source generation over the production OpenAPI document. Do not expect production OpenAPI to advertise internal admin stats.
3. For internal admin stats route shape, inspect the deployed `piplabs/depin-backend` source code and `story-deployments` image pin before answering from memory.
4. For aggregate user stats, call `/v1/admin/stats/user-growth` directly with the configured read-key header.
5. For aggregate submission stats, call `/v1/admin/stats/submissions` directly with the configured read-key header. Note: this returns only `[{date, count}]` with no dimension filtering.
6. For stats broken down by language (for example, "how many Vietnamese submissions"), call `/v1/admin/cohorts/languages?range=all` with the read key. This endpoint groups submissions by `users.primary_language` and returns `{range, items: [{language_code, user_count, submission_count, avg_submissions_per_user}]}`. Supported range values: `all`, `30d`, `90d`, `7d`, `1d`.
7. For stats broken down by nationality/country, call `/v1/admin/cohorts/demographics?dimension=country&range=all` only if a JWT-backed AdminAccess path is available. The admin read key alone will return `401`.
8. For a specific user lookup, note that the admin read key is aggregate-stats-only and will return `401` on `/v1/admin/users/**`. Report this limitation unless the task provides a separate authorized path.
9. If the public endpoint returns a Cloudflare block before reaching depin, report it as a Cloudflare/WAF routing issue and check the Cloudflare SoT before guessing.
10. If depin returns `401` or `403`, distinguish these cases without exposing credentials:
    - credential env var absent
    - credential env var mounted but rejected by prod
    - request blocked before reaching depin
    - stats key hitting a non-stats endpoint
11. Prefer `https://depin.storyprotocol.net` for production Numo/depin stats. Do not switch to staging APIs unless the user explicitly asks for staging data.
12. When REST endpoints are insufficient, use the native DB read path:
    a. **Verify column names first.** If your query joins to `campaigns`, query `SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'campaigns'` — the documentation may say `name` but the live column is `campaign_name`. This 1-row schema query avoids wasting an approval cycle on a typo.
    b. Inspect schema if needed, write one exact read-only SQL query, call `db_read.query`, wait for approval/resume, then answer from the sanitized tool result.

## Response Standard

- State whether the answer came from production live API data, native DB read, or repository/OpenAPI context.
- Include endpoint paths and status codes when debugging REST reads.
- For DB reads, include target, SQL semantics, row count, and truncation status when useful. Do not expose secrets or unredacted artifacts.
- For stats, include the query time and any filters/parameters used.
- If live data is unavailable, explain the exact blocker and the next required infrastructure or API fix.
- Do not ask the user for a credential when the credential is already mounted but rejected; report that the Vault/config value needs to be fixed.
- For simple numeric queries, lead with the number, then provide supporting breakdown in a compact table or bullet list.

## DB Observability Gap

The depin-backend API does not instrument database queries in its Prometheus metrics layer. `observability/metrics.rs` tracks HTTP requests, external service calls, and NDV endpoints, but has no DB query duration histograms, connection pool gauges, or query error counters.

The `db_query_duration_seconds_*` metrics that exist in Thanos belong to `story-orchestration-service`, not depin-backend. When asked about DB health for depin-backend:

1. Acknowledge the gap.
2. Infer DB health from indirect signals: HTTP 5xx rate, pod uptime/restarts, scrape duration, and API endpoint responsiveness.
3. Check admin stats endpoints. If `/v1/admin/stats/user-growth` returns data, the DB is reachable.
4. Note connection pool pressure: API pods and workers share the PostgreSQL instance.
5. Flag slow admin endpoints as likely query plan or indexing issues, not just missing metrics.
