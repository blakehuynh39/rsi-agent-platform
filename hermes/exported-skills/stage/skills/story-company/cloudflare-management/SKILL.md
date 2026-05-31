---
name: cloudflare-management
description: "Manage Cloudflare zones, WAF rules, IP lists, and DNS for Story Protocol domains via piplabs/cloudflare."
version: 1.0.0
metadata:
  hermes:
    tags: [cloudflare, waf, dns, pulumi, infrastructure, numo, depin]
    related_skills: [depin-prod-admin-read, numo-project-status]
---

# Cloudflare Management

Use this skill when diagnosing Cloudflare WAF blocks, modifying IP allowlists, changing WAF rules, or managing DNS records for Story Protocol domains (`numolabs.ai`, `storyprotocol.net`, `story.foundation`, etc.) via the `piplabs/cloudflare` repo.

## Source of Truth

- **Repo**: `piplabs/cloudflare` — Pulumi IaC for all Cloudflare zones
- **Zone WAF rules**: `src/zones/<zone>/waf.ts`
- **Zone rate limits**: `src/zones/<zone>/ratelimit.ts`
- **Zone DNS records**: `src/zones/<zone>/records.ts`
- **Zone hosts (domain sets)**: `src/zones/<zone>/hosts.ts`
- **IP lists (allow/block lists)**: `src/lists/numo_allowlists.ts` (List resources), `src/lists/numo_server_allowlist_items.ts` (ListItem entries)
- **Zone Access (Zero Trust)**: `src/zones/<zone>/access.ts`
- **GUIDANCE.md**: Conventions for importing existing CF resources into Pulumi state

## Domain Map

| Zone | API Hosts | WAF rules | Rate Limits |
|------|-----------|-----------|-------------|
| `numolabs.ai` | `api.numolabs.ai`, `staging-api.numolabs.ai` | `waf.ts` (origin-check, bearer-required, castle-required, bad-ua, threat-auth, etc.) | `ratelimit.ts` (4 tiers) |
| `storyprotocol.net` | Various | `waf.ts` | `ratelimit.ts` |
| `story.foundation` | Various | `waf.ts` | `ratelimit.ts` |
| `storyapis.com` | Various | `waf.ts` | `ratelimit.ts` |

## Diagnosing a Cloudflare Block

### 1. Read the Slack thread for the error

The Slack thread typically contains the curl command, HTTP response, and CF Ray ID. Key artifacts:

- **HTTP status**: 403 = WAF block, 429 = rate limit, 5xx = origin error
- **CF Ray ID**: The `-XXX` suffix indicates the CF data center (e.g., `SJC` = San Jose)
- **Block page type**: "Sorry, you have been blocked" = WAF custom rule; "Just a moment..." = managed challenge / Bot Fight Mode

### 2. Extract the blocked IP from the CF error page

**PITFALL:** Cloudflare block pages embed the visitor's IP in a hidden HTML element:
```html
<span class="hidden" id="cf-footer-ip"><IP_ADDRESS></span>
```
If the user pasted the raw HTML of the block page into Slack, the IP is visible in the source. This is standard CF behavior — every WAF block page includes the blocked IP behind the "Click to reveal" button.

### 3. Clone and inspect the WAF rules

```bash
cd /tmp && gh repo clone piplabs/cloudflare
cd cloudflare
```

Examine `src/zones/<zone>/waf.ts` for the WAF rules. Each rule has:
- `action`: `block`, `managed_challenge`, `skip`, or `log`
- `expression`: CF WAF expression language
- `ref`: unique rule identifier for CF dashboard correlation
- `enabled`: boolean

Cross-reference the host in the expression with `src/zones/<zone>/hosts.ts` to confirm which domain set is affected.

### 4. Identify the blocking rule

For each rule that could match, evaluate:

1. Does `http.host` match the request host?
2. Does `http.request.method` match?
3. Does the rule check `http.request.headers["origin"]` — and is it present/absent?
4. Is the IP in any bypass list (`$numo_server_allowlist`, `$numo_qa_allowlist`)?
5. Does a `not` clause flip the logic?

**Common pitfall:** The `numo-waf-origin-check` rule treats "no Origin header" the same as "bad Origin." Server-to-server calls (curl, workers, scripts) that don't send an `Origin` header get blocked unless the IP is in `$numo_server_allowlist`.

### 5. Search Honcho for prior investigations

```bash
mcp conversations_search(query="Cloudflare WAF <zone>")
mcp documents_search(query="Cloudflare WAF <zone>")
```

The Honcho corpus may contain prior diagnoses in private channels (`mirror_denied: true`) that are still useful context.

## Modifying IP Allowlists

### Adding an IP to `$numo_server_allowlist`

The list is defined in two parts:

1. **List resource** (`src/lists/numo_allowlists.ts`): Creates the CF List via Pulumi. Do NOT modify unless creating a new list.
2. **List items** (`src/lists/numo_server_allowlist_items.ts`): Manages individual ListItem resources.

To add an IP, add it to the `ipValues` array in `numo_server_allowlist_items.ts`:

```typescript
const ipValues = [
  // Existing entry example...
  "2601:640:8d80:2c70:c975:1831:b941:5ba5",  // Seb — numo-data-validation-worker
];
```

Each entry creates a `ListItem` resource. ListItem resources are additive — existing dashboard-managed entries in the list are unaffected. Pulumi only manages what's defined in this file.

**PR workflow:**
```bash
cd /tmp/cloudflare
git checkout -b feat/add-<name>-ip-allowlist
# Edit src/lists/numo_server_allowlist_items.ts
# Ensure src/lists/index.ts exports the file
git add -A && git commit -m "feat(numolabs.ai): add <name> IP to numo_server_allowlist"
git push -u origin feat/add-<name>-ip-allowlist
gh pr create --title "feat(numolabs.ai): add <name> IP to numo_server_allowlist" \
  --body "## Summary\n\nAdds <name>'s IP to \$numo_server_allowlist.\n\n## Root Cause\n\n..." \
  --base main
```

### Existing convention for IP list items

The repo uses two patterns:

1. **Greenfield pattern** (`uptime_robot.ts`): List AND items defined in same file. Use for new lists where Pulumi is the sole owner.
2. **Additive pattern** (`numo_server_allowlist_items.ts`): List defined separately, items added incrementally. Use for existing lists where the dashboard may have pre-existing entries.

## Skip Rule vs. IP Allowlisting — Decision Guide

When a WAF rule blocks legitimate traffic, you have two fix options:

| Approach | Best for | Trade-off |
|----------|----------|-----------|
| **Add IP to `$numo_server_allowlist`** | One-off personal IPs, fixed-infra servers | Doesn't scale — each new worker/operator needs another PR |
| **Add WAF skip rule** | Entire endpoint classes, internal service-to-server paths, or when the origin has its own auth | Surgical and scalable, but must verify the endpoint has adequate origin-level auth |

**Rule of thumb:** If the blocked endpoint is an **internal service-to-server path** with its own origin-level auth (Bearer token verification, HMAC, etc.), prefer a **skip rule**. It's the same pattern as the existing `numo-waf-slackbot-skip` rule — skip managed rules + SBFM for the specific paths, plus a path exclusion in the origin-check rule to bypass the custom WAF phase.

**PITFALL — don't create skip rules blindly.** When asked "is it safe to relax WAF for this endpoint?", verify the origin auth model first:
1. Clone `piplabs/depin-backend` (or the relevant origin repo) and find the route handler + auth extractor
2. Confirm auth runs *before* the handler body (extractor, middleware, or guard) — not an in-band check
3. Confirm it fails closed: missing/invalid credentials → non-2xx response
4. Cross-reference other WAF rules to confirm only the expected rule(s) match this path
5. Rate limiting is in a separate CF phase (`http_ratelimit`) — skip rules in `http_request_firewall_custom` don't affect rate limits, which is a good safety net
6. Document findings in `references/internal-endpoints-numolabs.md` for future sessions

See `references/internal-endpoints-numolabs.md` for the detailed NDV auth model verification performed in the 2026-05-06 session — use it as a template for future internal endpoint audits.

If the traffic is from a **person's dev machine** that can't send an Origin header (CLI, curl, scripts), IP allowlisting is usually simpler.

**CRITICAL PITFALL — `http_request_firewall_custom` phase is not authorized for skip rules on numolabs.ai zone.**

The `numolabs.ai` zone's Cloudflare plan does **not** support a WAF skip rule targeting the custom WAF phase. Attempting to include `http_request_firewall_custom` in the `phases` array results in:

```
error updating ruleset: skip action parameter phase 'http_request_firewall_custom' is not authorized (20120)
```

The existing `numo-waf-slackbot-skip` rule skips only `http_request_firewall_managed` + `http_request_sbfm` — never `http_request_firewall_custom`. This is intentional and documented in the failed Pulumi run for PR #203 (see `references/waf-skip-custom-phase-pitfall.md`).

**The correct skip rule template** (for `numolabs.ai` zone, matching slackbot pattern):

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

**BUT** since the skip rule can't bypass `numo-waf-origin-check` (which lives in the custom WAF phase), you must ALSO add a path exclusion directly to the origin-check rule's expression. Example from PR #204:

```typescript
// Inside numo-waf-origin-check expression, add BEFORE the Origin header checks:
(not http.request.uri.path in {\"/v1/internal/numo-data-validation/submissions\" \"/v1/internal/numo-data-validation/validation-results\"})
```

This two-part pattern (skip rule for managed/SBFM + path exclusion in origin-check) is the correct approach for any internal service-to-server endpoint on `numolabs.ai`.

PR workflow for skip rules: same as WAF rule modifications:
```
fix(numolabs.ai): add WAF skip for <service> internal endpoints
```

## Modifying WAF Rules

WAF rules are in `src/zones/<zone>/waf.ts`. Each rule is an object in the `rules` array.

**Key patterns:**
- `API_ALLOWED_ORIGINS`: Exact-match set of allowed Origin header values
- `DEV_PREVIEW_ORIGIN_REGEX`: Regex for dynamic origins (localhost, ngrok, Vercel)
- `$numo_server_allowlist`: IP bypass list for internal servers
- `$numo_qa_allowlist`: IP bypass list for QA testers

**PR workflow:** Same as IP allowlist, but commit message prefix should match the zone:
```
fix(numolabs.ai): allow server-to-server calls through origin-check
feat(storyprotocol.net): add WAF skip for new service
```

## Immediate Workarounds

When `numo-waf-origin-check` blocks a request and a permanent fix is pending:

- **Use the alternate domain**: `staging-depin.storyprotocol.net` points to the same ALB as `staging-api.numolabs.ai` but the `storyprotocol.net` zone has no origin-filtering WAF rule. Just swap the hostname in the URL; the path, method, and headers remain the same.

This pattern applies to other zones — the `storyprotocol.net` zone has fewer WAF restrictions than `numolabs.ai`.

## Pre-commit Checklist

Before pushing a Cloudflare PR:
- [ ] New file exported from parent `index.ts` (barrel export)
- [ ] No destructive changes to existing CF resources without import blocks
- [ ] CF expression syntax is valid (no mismatched braces, proper quoting)
- [ ] PR description links to the Slack thread and CF Ray ID for traceability

## Support Files

- `references/waf-rules-numolabs.md` — Full WAF rule expressions, rate-limit tiers, IP allowlist structure, and block page IP extraction pattern
- `references/internal-endpoints-numolabs.md` — Internal service-to-server endpoints behind numolabs.ai with their own origin-level auth (NDV, slackbot); guidance for WAF skip-rule decisions and auth model audit template
- `references/waf-skip-custom-phase-pitfall.md` — **CRITICAL**: The `numolabs.ai` zone plan does NOT authorize WAF skip rules targeting `http_request_firewall_custom` phase (error 20120). Correct two-part fix: skip rule for managed+SBFM only + path exclusion in origin-check expression. Evidence from failed Pulumi run #25461941289.
- `references/diagnosing-403-origin-vs-cf.md` — How to distinguish origin-level 403 (empty body, e.g. `NUMO_VALIDATION_ENABLED=false`) from Cloudflare WAF block (HTML block page). Includes error code mapping table and env var checklist.
- `references/pulumi-gha-log-fetching.md` — Reliable technique for downloading Pulumi GHA logs when `gh run view --log` returns empty. Zip download via `gh api` + Python extraction.
- `templates/pr-body-ip-allowlist.md` — PR description template for IP allowlist additions
