---
name: depin-prod-admin-read
description: "Live prod Numo/depin stats through admin REST reads and native approved DB reads."
version: 1.3.0
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

## Native DB Read Rules

- Use `db_read.sources` to list available DB targets.
- Use `db_read.schema` to inspect allowlisted tables and columns for a target.
- Use `db_read.validate` when you need repair feedback before submitting an approval request.
- Use `db_read.query` exactly once for the SQL you intend to run. The tool must show the exact SQL to an authorized admin before execution.
- After `db_read.query` pauses, do not use terminal, Kubernetes, or hand-built network calls to bypass the native permission flow.
- Do not self-approve. The approval is for an authorized DB-read admin, not necessarily the requester.
- The approval card is audit UI. The resumed Hermes run owns the final user-facing answer in the original Slack thread.
- If the DB result is denied, expired, or fails, use the resumed tool result to explain the blocker and propose a safe next query.

## Pitfalls

- **`execute_code` Python env is blind to `DEPIN_*` vars.** The Python subprocess spawned by `execute_code` runs in a limited environment that does not inherit the admin read key env vars. Use terminal only to confirm credential presence without printing values. If `execute_code` reports all vars as missing but terminal shows them present, trust the shell output.
- **Admin read key is stats-scoped only.** `DEPIN_ADMIN_READ_API_KEY` authorizes endpoints gated by `AdminReadOnlyAccess` (including `/v1/admin/stats/*`, `/v1/admin/cohorts/languages`, and `/v1/admin/overview`) but does not work for endpoints gated by `AdminAccess` such as `/v1/admin/users/**` and `/v1/admin/cohorts/demographics`. Those endpoints require JWT. A `401` from an `AdminAccess` endpoint is expected and should be reported as an auth scope mismatch, not a credential failure.
- **`/v1/admin/stats/submissions` is dimension-blind.** It returns only `[{date, count}]` with no language, nationality, campaign, or state filter. For "Vietnamese submissions", always try `/v1/admin/cohorts/languages?range=all` first. Only use DB read if the user asks for script/transcript-level semantics or campaign-language semantics that REST cannot answer.
- **`/v1/admin/submissions` is cursor-paginated with no total.** The response contains only `{items, next_cursor}`. Prefer aggregate endpoints or DB read for exact counts when pagination would be slow or ambiguous.
- **`cohorts/languages` uses `users.primary_language`, not campaign language.** For campaign-scoped language counts, find the campaign by `supported_languages` in `/v1/admin/campaigns`, then paginate `/v1/admin/submissions?campaign_id=X` or use native DB read when exact SQL is appropriate. Full enumeration technique and script: `references/campaign-language-filtering.md`.

## Source Of Truth

- Public API discovery: the deployed production OpenAPI document is intentionally client-facing and must not be treated as authoritative for internal admin stats routes.
- Internal admin stats contract: `piplabs/depin-backend`, especially `apps/api/src/http/routes/admin.rs`, `apps/api/src/http/extractors.rs`, `apps/api/src/services/admin_dashboard.rs`, and `docs/api-workflows.md`.
- Public DNS/WAF routing: `piplabs/cloudflare`, especially `src/zones/storyprotocol.net/records.ts` and `src/zones/storyprotocol.net/waf.ts`.
- Deployment and Vault wiring: `story-deployments`, `rsi-platform/rsi-agent-platform/use1-stage.yaml`, and `story/depin-backend/use1-prod.yaml`.
- Runtime readiness: the `story-deployments` depin admin read validation hook curls `/v1/admin/stats/user-growth` from the stage RSI cluster with the mounted read key and must fail loudly on non-200 responses.

## Query Pattern

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
12. When REST endpoints are insufficient, use the native DB read path: inspect schema if needed, write one exact read-only SQL query, call `db_read.query`, wait for approval/resume, then answer from the sanitized tool result.

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
