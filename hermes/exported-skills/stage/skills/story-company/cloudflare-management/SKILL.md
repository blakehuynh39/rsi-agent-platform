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
- `templates/pr-body-ip-allowlist.md` — PR description template for IP allowlist additions
