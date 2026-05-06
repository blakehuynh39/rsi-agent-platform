# WAF Rule Reference (numolabs.ai)

## numo-waf-origin-check (line 73)

```
(http.host in {"api.numolabs.ai" "staging-api.numolabs.ai"})
and (http.request.method in {"POST" "PUT" "PATCH" "DELETE"})
and (not http.request.headers["origin"][0] in {"https://app.numolabs.ai" "https://admin.numolabs.ai" "https://staging.numolabs.ai" "https://staging-admin.numolabs.ai"})
and (not http.request.headers["origin"][0] matches "^(https?://(localhost|127\.0\.0\.1)(:[0-9]+)?|https://[a-zA-Z0-9-]+\.ngrok(-free)?\.(dev|app)|https://numo-[a-zA-Z0-9-]+\.vercel\.app)$")
and (not ip.src in $numo_server_allowlist)
```

Action: `block`
Description: "Block state-changing API with bad Origin"

**Pitfall:** When a request has no `Origin` header (server-to-server calls, curl, workers), both `http.request.headers["origin"][0]` checks fail. This means the rule blocks ALL state-changing requests without an Origin header unless the IP is in `$numo_server_allowlist`. This is intentional for browser-origin security but needs manual IP allowlisting for internal tools.

## numo-waf-bearer-required (line 84)

```
(http.host in {"api.numolabs.ai" "staging-api.numolabs.ai"})
and (http.request.method ne "OPTIONS")
and (http.request.uri.path matches "^/v1/(me|submissions|scripts|campaigns)")
and (len(http.request.headers["authorization"][0]) lt 20)
and (not ip.src in $numo_server_allowlist)
```

Action: `block`
Description: "Block authed endpoints without Authorization header"

## numo-waf-bad-ua (line 111)

```
(http.host in {"api.numolabs.ai" "staging-api.numolabs.ai"})
and ((http.user_agent eq "") or (lower(http.user_agent) contains "curl") or ... )
and (not ip.src in $numo_server_allowlist)
```

Action: `managed_challenge` (NOT block — issues JS challenge, not hard 403)
Blacklisted UAs: curl, wget, python-requests, go-http-client, httpx, okhttp, axios, postman, insomnia

## Sanctions rules (lines 34-49, evaluated first)

```
(http.host in API_HOSTS or http.host in APP_HOSTS) and (ip.src.country in {"CU" "IR" "KP" "SY"})
(http.host in API_HOSTS or http.host in APP_HOSTS) and (ip.src.country in {"RU" "BY"})
(http.host in API_HOSTS or http.host in APP_HOSTS) and (ip.src.country eq "UA") and (ip.src.subdivision_1_iso_code in {"43" "40" "14" "05"})
```

Action: `block`

## numo-waf-datacenter-asn (line 132)

Blocks hosting/proxy ASNs (DigitalOcean, Linode, OVH, Hetzner, Vultr, Choopa, M247, DataCamp, EC2, AWS, GCP, Cloudflare) on auth and upload-init paths. Exempted by `$numo_server_allowlist`.

## Rate Limit Tiers (ratelimit.ts)

| Tier | Paths | Limit | Exemptions |
|------|-------|-------|------------|
| T1 | auth + reward actions | 15/min per IP | server + qa allowlists |
| T2 | enumeration reads | 20/min per IP | server + qa allowlists |
| T3 | user reads + profile | 120/min per IP | server + qa allowlists |
| T4 | API global backstop | 200/min per IP | server + qa allowlists |
| T5 | App domain global | 600/min per IP | qa allowlist only |

# IP Allowlist Structure

Two lists managed in `src/lists/`:

| List | File | Description |
|------|------|-------------|
| `numo_server_allowlist` | `numo_allowlists.ts` (resource) + `numo_server_allowlist_items.ts` (items) | Internal server IPs exempt from ALL WAF + rate-limit rules |
| `numo_qa_allowlist` | `numo_allowlists.ts` (resource only) | QA tester IPs — items managed in CF dashboard |

The `numo_server_allowlist` bypass is used in: origin-check, bearer-required, castle-required, bad-ua, datacenter-asn, and all four API rate-limit tiers.

# Cloudflare Block Page IP Extraction

Cloudflare WAF block pages embed the blocked visitor's IP in a hidden HTML element:

```html
<span id="cf-footer-item-ip" class="cf-footer-item hidden sm:block sm:mb-1">
  Your IP:
  <button type="button" id="cf-footer-ip-reveal" class="cf-footer-ip-reveal-btn">Click to reveal</button>
  <span class="hidden" id="cf-footer-ip">2601:640:8d80:2c70:c975:1831:b941:5ba5</span>
</span>
```

When users paste the raw HTML of the block page into Slack, the IP is visible in the source. This is often the fastest way to determine the blocked IP without asking the user to click "reveal" and report back.

# WAF Skip Rules (bypass patterns)

Skip rules are the preferred approach for internal service-to-service endpoints that have their own origin-level authentication. They're more general than per-IP allowlisting and don't require maintaining individual IP entries.

## numo-waf-slackbot-skip

```typescript
{
  action: "skip",
  actionParameters: {
    phases: ["http_request_firewall_managed", "http_request_sbfm"],
    ruleset: "current",
  },
  description: "numo-slackbot digest cron - bypass managed challenge + SBFM",
  enabled: true,
  expression: `(http.host in ${API_HOSTS}) and (http.request.method eq "GET") and (http.request.uri.path in {read-only admin paths}) and (http.user_agent contains "numo-slackbot/") and (http.request.headers["x-numo-source"][0] eq "slackbot")`,
  logging: { enabled: true },
  ref: "numo-waf-slackbot-skip",
}
```

Skips managed rules + SBFM only (NOT custom rules). Identity verified via User-Agent + custom header.

## numo-waf-skip-ndv-internal (PR #203)

```typescript
{
  action: "skip",
  actionParameters: {
    phases: ["http_request_firewall_custom", "http_request_firewall_managed", "http_request_sbfm"],
    ruleset: "current",
  },
  description: "numo-data-validation internal endpoints — bypass WAF",
  enabled: true,
  expression: `(http.host in ${API_HOSTS}) and (http.request.uri.path in {\"/v1/internal/numo-data-validation/submissions\" \"/v1/internal/numo-data-validation/validation-results\"})`,
  logging: { enabled: true },
  ref: "numo-waf-skip-ndv-internal",
}
```

Skips ALL three phases (custom, managed, SBFM). The endpoints have their own Bearer-token auth (`NumoValidationServiceSession`, SHA-256 hash match) so WAF origin-check is unnecessary. No User-Agent or header checks — path-only matching because the endpoint itself enforces auth.

### When to use each skip pattern

| Pattern | Use case | Phases skipped | Additional checks |
|---------|----------|----------------|-------------------|
| slackbot (managed+SBFM only) | Read-only cron jobs with identifiable UA | managed, sbfm | User-Agent + custom header |
| ndv-internal (all three) | Mutating internal endpoints with own auth | custom, managed, sbfm | Path only (origin auth suffices) |

**Rule:** When an endpoint already has its own origin-level authentication (Bearer token, API key, etc.), skip all three phases. When the endpoint relies on CF for some security (e.g., browser-facing), be more conservative and keep custom rules active.

# PR #201 (2026-05-05) — IP Allowlist pattern

Branch: `feat/add-seb-ip-allowlist` → `main`
Files changed:
- `src/lists/numo_server_allowlist_items.ts` (new) — ListItem resource for Seb's IPv6
- `src/lists/index.ts` — added barrel export

Key approach: Created a separate file for ListItems that references the existing List ID from `numo_allowlists.ts` rather than co-locating list definition and items. This is additive — it doesn't disrupt existing dashboard-managed entries.

# PR #203 (2026-05-06) — WAF Skip Rule pattern

Branch: `feat/add-ndv-waf-skip` → `main`
Files changed:
- `src/zones/numolabs.ai/waf.ts` — added skip rule before origin-check (+17 lines)

Key approach: Placed skip rule between existing slackbot skip and origin-check. Expression uses path-only matching (`http.request.uri.path in {…}`) because the endpoint has its own auth. Covered both GET and POST NDV paths proactively.
