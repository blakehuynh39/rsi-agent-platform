---
title: "Slack thread C0ASQ9K5V50 1777683214.676429"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_fdbe3fe8355612f210e6de138ceafcbd"
source_key: "slack:T045QQQQ7CZ:C0ASQ9K5V50:1777683214.676429"
source_session_key: "slack:T045QQQQ7CZ:C0ASQ9K5V50:1777683214.676429"
source_revision_ids:
  - "srcrev_0eecc72319946eb603ff50236cef0cf0"
  - "srcrev_16f04e9f1676bb31f506f703cfa0ee48"
  - "srcrev_3b5c4ba60e16980febcc5327918d820b"
  - "srcrev_58895f38878c581ae7c462672b0f6105"
  - "srcrev_6bd78d9a9db64d9d71263f227600f670"
  - "srcrev_b490e5209a9981747aea0d7e0673f426"
  - "srcrev_b4d89c9bbebbcdce5e3e7173ac774d48"
  - "srcrev_b57b7790700ccc86638f98dd047e4fb8"
  - "srcrev_cbc1a808b8c1a747731b7572512f99f2"
  - "srcrev_daa60c9c51149346cf8330be9b50c892"
  - "srcrev_e42fede5393c5348bdb2500f1cf39eb0"
  - "srcrev_eddc9be4b4ba0197ab3ec9c6ab412dcb"
conflicts: []
---

# Slack thread C0ASQ9K5V50 1777683214.676429

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_e42fede5393c5348bdb2500f1cf39eb0`
- `chunk_id`: `srcchunk_b239a81229c0c6a596f36eb2ec9dd8e8`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777683214.676429`

<@U0ASDQKU3UL> can you help with <https://storyprotocol.slack.com/archives/C0AKH5SNGKH/p1777650186068179>

### Citation 2

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_3b5c4ba60e16980febcc5327918d820b`
- `chunk_id`: `srcchunk_9b4d84677932156516a53bd6201c3bd7`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777683232.032629`

Tracking this RSI run: <https://staging-rsi-platform.storyprotocol.net/?conversation=conv-4d842cab857d43bdafcd63fa1080c2c9&amp;tab=conversations&amp;trace=trace-c0bfedefff2a43c2a10b824770f5269c|open trace>

### Citation 3

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_0eecc72319946eb603ff50236cef0cf0`
- `chunk_id`: `srcchunk_2d7860e617791a38a757728823f2dbfc`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777683623.316799`

I re-investigated the CORS thread at <https://storyprotocol.slack.com/archives/C0AKH5SNGKH/p1777650186068179>. The issue is still present — here's the updated status:

*Root cause*: Cloudflare WAF rule `numo-waf-origin-check` on the `<http://numolabs.ai|numolabs.ai>` zone blocks all POST/PUT/PATCH/DELETE requests unless the `Origin` header is in its allowlist. The allowlist currently only has:
• `<https://app.numolabs.ai>`
• `<https://admin.numolabs.ai>`
• `<https://staging.numolabs.ai>`
• `<https://staging-admin.numolabs.ai>`

It does *not* include `<http://localhost:*>`, `https://*.ngrok.dev`, or `<https://numo>-*.<http://vercel.app|vercel.app>` — even though the application-level CORS in depin-backend's staging.toml already allows them.

*Live test confirmation (just run)*:
• `POST <https://staging-api.numolabs.ai/v1/auth/exchange>` with `Origin: <http://localhost:3000>` → *HTTP 403* (Cloudflare WAF block)
• `POST <https://staging-depin.storyprotocol.net/v1/auth/exchange>` with same Origin → *HTTP 401* (reaches the app — expected auth failure, no WAF block)
• `OPTIONS <https://staging-depin.storyprotocol.net/v1/auth/exchange>` → *HTTP 200* with `Access-Control-Allow-Origin: <http://localhost:3000>` ✓

Both domains point to the same ALB (`<http://use1-stage-depin-backend-1940402382.us-east-1.elb.amazonaws.com|use1-stage-depin-backend-1940402382.us-east-1.elb.amazonaws.com>`) — the difference is purely the Cloudflare zone's WAF rules.

*Two fix options*:
1. *Immediate workaround*: Use `<https://staging-depin.storyprotocol.net>` instead of `<https://staging-api.numolabs.ai>` for local/dev work (the <http://storyprotocol.net|storyprotocol.net> zone has no origin-filtering WAF rule).
2. *Permanent fix*: Update `API_ALLOWED_ORIGINS` in `cloudflare/src/zones/numolabs.ai/waf.ts` (line 13) to include `<http://localhost:*>`, `https://*.ngrok.dev`, `https://*.ngrok-free.dev`, and `<https://numo>-*.<http://vercel.app|vercel.app>`. The application CORS config already supports glob patterns — the WAF needs to match.

*Relevant files*:
• WAF rule: `cloudflare/src/zones/numolabs.ai/waf.ts:12-13` (`API_ALLOWED_ORIGINS` constant, line 66 for the `numo-waf-origin-check` expression)
• App CORS: `depin-backend/apps/api/config/staging.toml:11` (`cors_allowed_origins`)
• No changes to the WAF file since May 1 — this still needs a PR

### Citation 4

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_b57b7790700ccc86638f98dd047e4fb8`
- `chunk_id`: `srcchunk_a17acdc80c0d21ed140230984ba46012`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777683651.317469`

<@U0ASDQKU3UL> can you make a PR for it to cloudflare repo

### Citation 5

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_6bd78d9a9db64d9d71263f227600f670`
- `chunk_id`: `srcchunk_7134dfc805d5c84fb547e940a68e72c5`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777683673.169459`

Tracking this RSI run: <https://staging-rsi-platform.storyprotocol.net/?conversation=conv-4d842cab857d43bdafcd63fa1080c2c9&amp;tab=conversations&amp;trace=trace-2979c2884914425c88a0a010a62d37a6|open trace>

### Citation 6

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_58895f38878c581ae7c462672b0f6105`
- `chunk_id`: `srcchunk_36d2cb0bfc9fd45b6646d0e455b6eebd`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777699634.606059`

<@U0ASDQKU3UL> can you make a PR for it to cloudflare repo

### Citation 7

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_eddc9be4b4ba0197ab3ec9c6ab412dcb`
- `chunk_id`: `srcchunk_2be9b10b669ee8d279afbd39e6ab1b6e`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777699638.847519`

Tracking this RSI run: <https://staging-rsi-platform.storyprotocol.net/?conversation=conv-4d842cab857d43bdafcd63fa1080c2c9&amp;tab=conversations&amp;trace=trace-e69b3e7dd87c46439253227e40cd9a7e|open trace>

### Citation 8

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_daa60c9c51149346cf8330be9b50c892`
- `chunk_id`: `srcchunk_798efba38c92c1ddeb9a970cb1dd185c`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777705414.572059`

<@U0ASDQKU3UL> can you make a PR for it to cloudflare repo

### Citation 9

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_b4d89c9bbebbcdce5e3e7173ac774d48`
- `chunk_id`: `srcchunk_7662d5a2877a0f4361969e4436d87246`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777705416.410399`

Tracking this RSI run: <https://staging-rsi-platform.storyprotocol.net/?conversation=conv-4d842cab857d43bdafcd63fa1080c2c9&amp;tab=conversations&amp;trace=trace-acb1260d036443a3ae8543ded407506c|open trace>

### Citation 10

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_b490e5209a9981747aea0d7e0673f426`
- `chunk_id`: `srcchunk_0304cd03c3d31cdc7ad76e17e9ae0c25`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777709498.960249`

<@U0ASDQKU3UL> can you make a PR for it to cloudflare repo

### Citation 11

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_16f04e9f1676bb31f506f703cfa0ee48`
- `chunk_id`: `srcchunk_6a12798800d63348c27575089bab3bff`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777709501.265839`

Tracking this RSI run: <https://staging-rsi-platform.storyprotocol.net/?conversation=conv-4d842cab857d43bdafcd63fa1080c2c9&amp;tab=conversations&amp;trace=trace-4942da2166d241c4b5c2867e7a2e639f|open trace>

### Citation 12

- `source_document_id`: `srcdoc_fdbe3fe8355612f210e6de138ceafcbd`
- `source_revision_id`: `srcrev_cbc1a808b8c1a747731b7572512f99f2`
- `chunk_id`: `srcchunk_c44555ef41c38e4e90bbad1388f4b182`
- `native_locator`: `slack:C0ASQ9K5V50:1777683214.676429:1777709720.081119`

PR created: <https://github.com/piplabs/cloudflare/pull/200>

*fix(<http://numolabs.ai|numolabs.ai>): allow localhost, ngrok, and Vercel preview origins through WAF origin check*

The WAF `numo-waf-origin-check` rule now has a `DEV_PREVIEW_ORIGIN_REGEX` that allows:
• `http(s)://localhost(:port)?`
• `https://*.ngrok(-free).(dev|app)`
• `<https://numo>-*.<http://vercel.app|vercel.app>`

This runs alongside the existing exact-match set so production origins are unaffected. Once this is merged and applied, `<http://staging-api.numolabs.ai|staging-api.numolabs.ai>` will accept CORS requests from localhost, ngrok tunnels, and Vercel previews — matching what the app-level CORS already allows.

Workaround until merged: use `<https://staging-depin.storyprotocol.net>` (same ALB, no origin WAF rule).

