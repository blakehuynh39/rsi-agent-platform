# Pitfall: WAF Skip Rule Targeting `http_request_firewall_custom` Phase

## The error

When adding a WAF skip rule on the `numolabs.ai` zone that includes `http_request_firewall_custom` in the `phases` array, the Pulumi `up` fails with:

```
cloudflare:index:Ruleset waf_numolabs updating (2s) [diff: ~rules];
  error: error updating ruleset with ID "85959585fae5454a9631c866964303ba":
  skip action parameter phase 'http_request_firewall_custom' is not authorized (20120)
```

This is a **Cloudflare plan limitation** — the `numolabs.ai` zone plan does not support skip rules targeting the custom WAF phase.

## Evidence

- **Failed run**: https://github.com/piplabs/cloudflare/actions/runs/25461941289/job/74706050278
- **PR that triggered it**: #203 — added `numo-waf-skip-ndv-internal` with phases `["http_request_firewall_custom", "http_request_firewall_managed", "http_request_sbfm"]`
- **Fix PR**: #204 — removed `http_request_firewall_custom` from phases, added path exclusion to origin-check

## How to verify the zone plan limitation

The existing `numo-waf-slackbot-skip` rule (`waf.ts:55-66`) **only** skips `http_request_firewall_managed` and `http_request_sbfm`. It never includes `http_request_firewall_custom`. This is the reference pattern — any new skip rule on `numolabs.ai` must match this.

## Correct fix pattern (two-part)

### Part 1: Skip rule (managed rules + SBFM only)

```typescript
{
  action: "skip",
  actionParameters: {
    phases: ["http_request_firewall_managed", "http_request_sbfm"],
    ruleset: "current",
  },
  description: "<service> internal endpoints — bypass managed rules + SBFM",
  enabled: true,
  expression: `(http.host in ${API_HOSTS}) and (http.request.uri.path in {\"/v1/internal/<path>\"})`,
  logging: { enabled: true },
  ref: "numo-waf-skip-<service>-internal",
},
```

### Part 2: Path exclusion in origin-check rule

Since the skip rule can't bypass the custom WAF phase, add a path exclusion directly to `numo-waf-origin-check`'s expression, **before** the Origin header checks:

```typescript
// Inside the origin-check expression, add:
(not http.request.uri.path in {\"/v1/internal/numo-data-validation/submissions\" \"/v1/internal/numo-data-validation/validation-results\"})
```

Full corrected expression (from PR #204):

```
(http.host in ${API_HOSTS})
and (http.request.method in {"POST" "PUT" "PATCH" "DELETE"})
and (not http.request.uri.path in {"/v1/internal/numo-data-validation/submissions" "/v1/internal/numo-data-validation/validation-results"})
and (not http.request.headers["origin"][0] in ${API_ALLOWED_ORIGINS})
and (not http.request.headers["origin"][0] matches ${DEV_PREVIEW_ORIGIN_REGEX})
and (not ip.src in $numo_server_allowlist)
```

## Why this is safe

The path exclusion in the origin-check expression is equivalent to a skip rule — it just achieves the same result through a different mechanism. The endpoints still have:
- Their own origin-level auth (Bearer token, SHA-256 hash, constant-time comparison)
- Rate limiting in separate `http_ratelimit` phase (unaffected)
- Managed rules + SBFM skipped by the companion skip rule

## Applicable zones

Confirmed for: `numolabs.ai`
Not confirmed (but likely safe): `storyprotocol.net`, `story.foundation`, `storyapis.com` — these zones have fewer WAF rules overall, but verify before adding skip rules.
