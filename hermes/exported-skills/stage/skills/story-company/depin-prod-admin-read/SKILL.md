---
name: depin-prod-admin-read
description: "Live prod Numo/depin user/submission stats admin reads."
version: 1.1.0
metadata:
  hermes:
    tags: [numo, depin, production, prod, admin, read-only, user-stats, users, submissions, api, vault]
    related_skills: [github-auth, native-mcp]
---

# Depin Prod Admin Read

Use this skill when a Story request asks for live Numo/depin user stats, submission stats, admin user lookup, or production depin API route context.

## Runtime Contract

- Use `DEPIN_ADMIN_BASE_URL` as the base URL. Stage RSI intentionally points this at the production public endpoint, `https://depin.storyprotocol.net`.
- Use `DEPIN_ADMIN_READ_API_KEY_HEADER` for the header name and `DEPIN_ADMIN_READ_API_KEY` for the header value.
- Use `User-Agent: rsi-hermes-company-computer/1.0` and `X-RSI-Source: hermes` on direct prod depin reads so Cloudflare can distinguish the company-computer path from generic scripted traffic. These headers are not credentials; origin authorization still depends on the admin read key.
- Never print, summarize, export, or store the credential value. Only report whether it is present.
- Treat this credential as read-only. Do not use it for write, mutation, delete, or admin management actions.

## Pitfalls

- **`execute_code` Python env is blind to `DEPIN_*` vars.** The Python subprocess spawned by `execute_code` runs in a limited environment that does not inherit the admin read key env vars. Use `terminal` (shell `env | grep DEPIN`) to confirm credential presence. If `execute_code` reports all vars as MISSING but `terminal` shows them present, trust the shell output.
- **Admin read key is stats-scoped only.** `DEPIN_ADMIN_READ_API_KEY` authorizes endpoints gated by `AdminReadOnlyAccess` (which includes `/v1/admin/stats/*`, `/v1/admin/cohorts/languages`, and `/v1/admin/overview`) but does **not** work for endpoints gated by `AdminAccess` such as `/v1/admin/users/**` and `/v1/admin/cohorts/demographics`. Those endpoints require JWT. `/v1/admin/cohorts/demographics` returning `401` with the stats key is expected — it needs `Authorization: Bearer <jwt>`, not the read key header. Do not report "credential rejected" when a `401` comes from an `AdminAccess`-gated endpoint.
- **`/v1/admin/stats/submissions` is dimension-blind.** It returns only `[{date, count}]` with no way to filter by language, nationality, campaign, or state. When the user asks for submissions by a specific language or country (e.g., "Vietnamese submissions"), this endpoint cannot answer the question. Reach for `/v1/admin/cohorts/languages` (by `users.primary_language`) or `/v1/admin/cohorts/demographics` (by `users.nationality`, JWT required) instead.

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
5. For aggregate submission stats, call `/v1/admin/stats/submissions` directly with the configured read-key header. Note: this returns only `[{date, count}]` — no dimension filtering.
6. For stats broken down by **language** (e.g., "how many Vietnamese submissions"), call `/v1/admin/cohorts/languages?range=all` with the read key. This endpoint groups submissions by `users.primary_language` and returns `{range, items: [{language_code, user_count, submission_count, avg_submissions_per_user}]}`. Supported range values: `all`, `30d`, `90d`, `7d`, `1d`. (Source: `admin_dashboard.rs` `fetch_languages_cohort`.)
7. For stats broken down by **nationality/country** (e.g., "how many submissions from Vietnam/VN"), call `/v1/admin/cohorts/demographics?dimension=country&range=all`. **Note:** this endpoint is gated by `AdminAccess` (JWT), NOT `AdminReadOnlyAccess` — the admin read key will return `401`. If the user needs nationality data, report the auth limitation. (Source: `admin.rs` line 704–711, `admin_dashboard.rs` `fetch_demographics`.)
8. For a specific user lookup, note that the admin read key is **aggregate-stats-only** and will return `401` on `/v1/admin/users/**`. Per-user routes require a different authorization mechanism. Report this limitation to the user rather than treating it as a credential failure.
9. If the public endpoint returns a Cloudflare block before reaching depin, report it as a Cloudflare/WAF routing issue and check the Cloudflare SoT before guessing.
10. If depin returns `401` or `403`, distinguish these cases without exposing the credential:
   - credential env var absent (check with `terminal` shell, not `execute_code` Python)
   - credential env var mounted but rejected by prod
   - request blocked before reaching depin
   - **stats key hitting a non-stats endpoint** (e.g., `/v1/admin/users`) — the key is working; the endpoint expects different auth. Report this as an auth scope mismatch, not a key rejection.
11. Prefer `https://depin.storyprotocol.net` for production Numo/depin stats. Do not switch to staging APIs unless the user explicitly asks for staging data.

## Response Standard

- State whether the answer came from production live API data or from repository/OpenAPI context.
- Include endpoint paths and status codes when debugging.
- For stats, include the query time and any filters/parameters used.
- If live data is unavailable, explain the exact blocker and the next required infrastructure or API fix.
- Do not ask the user for a credential when the credential is already mounted but rejected; report that the Vault/config value needs to be fixed.

## DB Read Tool (Direct SQL)

When the admin REST API endpoints don't support the needed query (e.g., table-level counts, arbitrary filters, joins not exposed via REST), use the **db-read tool** — a control-plane service that runs read-only SQL directly against prod PostgreSQL via an AWS Lambda proxy.

### Architecture

- **Control plane API**: `http://<control-plane-pod-ip>:8080/internal/db-read/*`
- **Credentials**: `RSI_DB_READ_CLIENT_TOKEN` env var (bearer token in `Authorization: Bearer <token>` header)
- **Prod path**: control-plane → db-read-worker → AWS Lambda `use1-prod-rsi-db-read-depin-prod` → prod PostgreSQL
- **Stage path**: control-plane → db-read-worker → direct PostgreSQL connection (Vault-injected DSN)
- **Worker pod**: `use1-stage-rsi-agent-platform-control-plane-db-read-worker-*` in `rsi-platform` namespace

### Discovery

1. **Find control plane pod IP**: `kubectl get pod -n rsi-platform -l app.kubernetes.io/component=control-plane -o jsonpath='{.items[0].status.podIP}'`
2. **Confirm token**: `env | grep RSI_DB_READ_CLIENT_TOKEN`
3. **List available targets**: `GET /internal/db-read/sources` with `Authorization: Bearer $RSI_DB_READ_CLIENT_TOKEN`

### Prod Caps and Constraints

| Cap | Value |
|-----|-------|
| max_rows | 100 |
| max_bytes | 65,536 |
| timeout_seconds | 20 |
| lock_timeout_ms | 250 |
| approval_ttl | 1 hour |

**Redacted columns** (automatically stripped from results): `access_token`, `api_key`, `authorization`, `email`, `password`, `phone`, `private_key`, `refresh_token`, `secret`, `token`.

### Query Flow

1. `POST /internal/db-read/validate` with `{target, sql}` to pre-validate (optional but recommended)
2. `POST /internal/db-read/query` with `{target, sql, purpose, requester, channel_id, thread_ts, ...}` to submit
3. The control plane validates SQL, stores the request, and for prod targets posts a Slack approval card in the specified channel/thread
4. The approving user clicks "Approve once" — the query goes to the Lambda and results are posted back to the thread
5. Results include row count, truncation status, and a sample

### Approval Flow

Prod queries require explicit human approval via Slack interactive buttons ("Approve once" / "Deny"). The approval card is posted to the thread specified in `channel_id`/`thread_ts`. Stage queries may auto-approve depending on configuration.

**Important**: Do NOT try to self-approve programmatically. Wait for the user to click "Approve once" in Slack.

### Query Examples

```
# Count active scripts (transcripts)
SELECT COUNT(*) AS available_transcripts FROM scripts WHERE is_active = TRUE

# Active scripts by language in active campaigns
SELECT s.language_code, COUNT(*) AS transcript_count
FROM scripts s JOIN campaigns c ON c.id = s.campaign_id
WHERE s.is_active = TRUE AND c.is_active = TRUE
GROUP BY s.language_code ORDER BY transcript_count DESC LIMIT 20
```

### Pitfalls

- **kubectl port-forward is RBAC-blocked.** The hermes executor SA lacks `pods/portforward` permission in `rsi-platform`. Use direct pod IP instead.
- **Control plane pod IP changes on restart.** Re-run the discovery step each session.
- **Prod timeout is 20s.** Keep queries simple. Complex joins on large tables may timeout.
- **Results sample may be truncated.** The `truncated` field in the result indicates if the full result exceeded the byte cap.
- **SQL is validated offline first**, then on-target. Make sure column names exist in the schema before submitting. Use `/internal/db-read/schema?target=depin-prod` to inspect available tables and columns.
- Full API details, endpoint schemas, and request/response formats are documented in `references/db-read-api.md`.

## DB Observability Gap

The depin-backend API does **not** instrument database queries in its Prometheus metrics layer. `observability/metrics.rs` tracks HTTP requests, external service calls (Dynamic, World), and NDV endpoints — but has zero DB query metrics (no histograms, no connection pool gauges, no query error counters).

The `db_query_duration_seconds_*` metrics that exist in Thanos belong to `story-orchestration-service`, not depin-backend. When asked about DB health for depin-backend, you must:

1. **Acknowledge the gap** — explain that there are no direct DB metrics
2. **Infer DB health from indirect signals**: HTTP 5xx rate (zero means DB isn't failing), pod uptime/restarts, scrape duration, and API endpoint responsiveness
3. **Check the admin stats endpoints** — if `/v1/admin/stats/user-growth` returns data, the DB is alive
4. **Note connection pool pressure** — the API uses `max_connections=10` per pod (4 API pods = 40 potential connections), and the IP registration worker uses another `max_connections=10` per pod (8 worker pods = 80 more). Total: up to 120 connections to the same PostgreSQL 16 instance.
5. **Flag slow admin endpoints** — `/v1/admin/users` had p95 latency of 4.8s as of May 2026, suggesting a query plan or indexing issue, not just a lack of metrics
