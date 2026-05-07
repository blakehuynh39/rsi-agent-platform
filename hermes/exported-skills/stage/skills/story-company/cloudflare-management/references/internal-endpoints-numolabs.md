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

### Auth model (verified from source)

**Extractor**: `NumoValidationServiceSession` (`extractors_numo_validation.rs:100-160`)

Auth flow (rejects before handler body runs):

1. **Extracts Bearer token** from `Authorization` header (line 222-228): `value.strip_prefix("Bearer ")`
2. **SHA-256 hash** of presented token: `hex::encode(Sha256::digest(token.as_bytes()))` (line 198)
3. **Constant-time comparison** against configured hash using `subtle::ConstantTimeEq` (line 215-219) — resistant to timing side-channel attacks
4. **Dual-hash rotation**: matches against both `service_token_sha256` AND `service_token_sha256_next` (lines 190-193), enabling zero-downtime credential rollover

**Fail-closed behavior** (lines 85-98):
| Condition | Error | HTTP |
|-----------|-------|------|
| `numo_validation.enabled = false` | `Disabled` | 403 |
| No configured hash | `Internal` | 500 |
| Missing `Authorization` header | `TokenMissing` | 401 |
| Token doesn't match any configured hash | `TokenInvalid` | 401 |

**What is NOT used for auth:**
- Source IP — extracted via `client_ip()` with `trusted_proxy_hops` but **observability only** (line 126-136). Confirmed by test at line 465-482: "valid token succeeds without resolved source IP"
- Origin header — never inspected
- Castle / Turnstile — not involved in NDV auth path

**Observability**: All auth outcomes logged with `source_ip`, `request_id`. Failures tracked per-endpoint via dedicated metrics (`track_numo_validation_export_request` / `track_numo_validation_ingest_request` — lines 162-176).

### Ingest-specific security layers (POST only)

**Idempotency** (`routes/numo_validation.rs:216-245`):
- `Idempotency-Key` header **required** — 400 if missing
- DB-backed claim with four outcomes:
  - `Claimed` — new request, proceeds to processing
  - `Replay` — same key + same body hash → returns cached response, no DB writes
  - `Conflict` — same key + different body → 409
  - `InProgress` — concurrent request → 409

**Body enforcement** (lines 185-203):
- Size clamped to `numo_validation.max_batch_bytes` → 413 if exceeded
- JSON deserialization validated (serde)

**Export validation** (lines 96-113):
- `updated_after` and `cursor` are mutually exclusive → 400
- Malformed cursor → 400
- Limit clamped to `numo_validation.max_export_limit`

### WAF rule cross-reference (why skip is safe)

Only `numo-waf-origin-check` actively blocks these endpoints. Other custom rules don't match:

| WAF rule | Path expression | Matches NDV? |
|----------|----------------|-------------|
| `numo-waf-origin-check` | All POST/PUT/PATCH/DELETE | **Yes** (the one we're skipping) |
| `numo-waf-bearer-required` | `^/v1/(me\|submissions\|scripts\|campaigns)` | No — doesn't match `/v1/internal/...` |
| `numo-waf-castle-required` | `^/v1/submissions/...` | No |
| `numo-waf-datacenter-asn` | `^/v1/(auth\|submissions/initiate-upload)` | No |
| `numo-waf-threat-auth` | `^/v1/auth` | No |

**Rate limiting**: In separate `http_ratelimit` phase — not skipped by the WAF skip rule. T1-T4 tiers exempt `$numo_server_allowlist`. T4 global backstop (200/min per IP) still applies to non-allowlisted callers.

### WAF implications

- Because these endpoints never carry an Origin header, `numo-waf-origin-check` blocks them unless the caller's IP is in `$numo_server_allowlist`
- The endpoints have their own auth → safe to add a WAF skip rule
- Skip rule should cover both GET and POST paths
- Pattern: `numo-waf-skip-ndv-internal` (can use slackbot-skip as template)

### Related PRs

- PR #201 (cloudflare): Added Seb's IPv6 to `$numo_server_allowlist` — workaround, now superseded
- PR #203 (cloudflare): WAF skip rule `numo-waf-skip-ndv-internal` for both NDV paths (created 2026-05-06)
- PR #204 (cloudflare): **FIX** for PR #203 — removed `http_request_firewall_custom` from skip phases, added NDV path exclusion to origin-check expression. Zone plan doesn't authorize custom-phase skip rules (error 20120). See `references/waf-skip-custom-phase-pitfall.md`.
- PR #449 (depin-backend): Added `enabled = true` explicitly to `production.toml`'s `[numo_validation]` section (2026-05-07). Production config had empty section header — explicit value removes ambiguity and ensures NDV endpoints are enabled.

## Slackbot Digest Cron

**Source**: `piplabs/numo-monorepo/pull/169`
**WAF rule**: `numo-waf-slackbot-skip` in `waf.ts:55-66`

Auth via `X-Numo-Source: slackbot` header + specific User-Agent (`numo-slackbot/`) — secret managed in Vault. Skip rule already in place for GET-only admin reads.

## Template for future internal endpoints

When adding a new internal service-to-server endpoint behind `numolabs.ai`:

1. Ensure the endpoint has its own auth (Bearer token, HMAC, mTLS, or Vault secret header)
2. Add a skip rule to `waf.ts` for the path(s) — skip managed rules + SBFM phases only (the `numolabs.ai` zone plan does NOT authorize `http_request_firewall_custom` phase for skip rules — see `references/waf-skip-custom-phase-pitfall.md`)
3. ALSO add a path exclusion to the `numo-waf-origin-check` rule's expression: `(not http.request.uri.path in {<paths>})`
4. Place the skip rule before `numo-waf-origin-check` in the rules array
5. Use `logging: { enabled: true }` for audit trail
