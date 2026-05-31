## Summary

Adds <NAME>'s IP address (`<IP_ADDRESS>`) to the `numo_server_allowlist` Cloudflare IP list as a Pulumi-managed `ListItem`.

## Root Cause

<NAME>'s <SERVICE> hits `<HOST>` with <METHOD> requests that <TRIGGER_DESCRIPTION>. The `<WAF_RULE>` WAF rule blocks these requests because:

1. The host matches `<HOST>` ✓
2. The method is <METHOD> ✓
3. <CONDITION_1> ✓
4. <NAME>'s IP is NOT in `$numo_server_allowlist` ✓

→ Result: HTTP <CODE> WAF block (CF Ray `<RAY_ID>`)

## Fix

New entry in `src/lists/numo_server_allowlist_items.ts` creates a `ListItem` resource for <NAME>'s IP referencing the existing `numo_server_allowlist` list (defined in `numo_allowlists.ts`).

This exempts <NAME>'s IP from:
- `numo-waf-origin-check` (origin validation)
- `numo-waf-bearer-required` (auth header check)
- `numo-waf-castle-required` (Castle token check)
- `numo-waf-bad-ua` (managed challenge on scripted UAs)
- `numo-waf-datacenter-asn` (hosting ASN blocks)
- All four rate-limit tiers (T1–T4)

Existing dashboard-managed list entries are unaffected — Pulumi only manages the items defined in this file.

## Related

- Slack thread: <THREAD_LINK>
- CF Ray: `<RAY_ID>`
