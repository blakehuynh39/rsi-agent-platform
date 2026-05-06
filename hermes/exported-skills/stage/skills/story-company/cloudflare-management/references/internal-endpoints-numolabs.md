# Internal Endpoints Reference (depin-backend)

Internal endpoints hosted behind `api.numolabs.ai` / `staging-api.numolabs.ai` that are server-to-server (no browser Origin header) with their own origin-level auth. These are candidates for WAF skip rules when blocked by `numo-waf-origin-check`.

## Numo Data Validation (NDV)

**Source repo**: `piplabs/depin-backend`
**Route file**: `apps/api/src/http/routes/numo_validation.rs`
**Extractor**: `apps/api/src/http/extractors_numo_validation.rs`
**Architecture docs**: `docs/architecture/api.md:173-250`

### Endpoints

| Method | Path | Auth | Notes |
|--------|------|------|-------|
| GET | `/v1/internal/numo-data-validation/submissions` | Bearer token (SHA-256) | Export pending_review audio submissions; cursor-based pagination |
| POST | `/v1/internal/numo-data-validation/validation-results` | Bearer token (SHA-256) | Batch ingest of per-submission validation decisions; Idempotency-Key required |

### Auth model

- `Authorization: Bearer <token>` — SHA-256 hash compared against `NUMO_VALIDATION_SERVICE_TOKEN_SHA256` or `NUMO_VALIDATION_SERVICE_TOKEN_SHA256_NEXT` (rotation overlap)
- No Origin header dependency — these are service-to-server calls
- Source IP is observability context only (logged, not authorization gate)
- Auth failures counted in NDV metrics

### WAF implications

- Because these endpoints never carry an Origin header, `numo-waf-origin-check` blocks them unless the caller's IP is in `$numo_server_allowlist`
- The endpoints have their own auth → safe to add a WAF skip rule
- Skip rule should cover both GET and POST paths
- Pattern: `numo-waf-skip-ndv-internal` (can use slackbot-skip as template)

### Related PRs

- PR #201 (cloudflare): Added Seb's IPv6 to `$numo_server_allowlist` — workaround for now
- Pending: skip rule proposal for both NDV paths (2026-05-06 thread)

## Slackbot Digest Cron

**Source**: `piplabs/numo-monorepo/pull/169`
**WAF rule**: `numo-waf-slackbot-skip` in `waf.ts:55-66`

Auth via `X-Numo-Source: slackbot` header + specific User-Agent (`numo-slackbot/`) — secret managed in Vault. Skip rule already in place for GET-only admin reads.

## Template for future internal endpoints

When adding a new internal service-to-server endpoint behind `numolabs.ai`:

1. Ensure the endpoint has its own auth (Bearer token, HMAC, mTLS, or Vault secret header)
2. Add a skip rule to `waf.ts` for the path(s) — skip custom WAF + managed + SBFM phases
3. Place the skip rule before `numo-waf-origin-check` in the rules array
4. Use `logging: { enabled: true }` for audit trail
