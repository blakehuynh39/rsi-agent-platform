---
title: "Hub API Staging Endpoint Update"
type: "runbook"
slug: "runbooks/hub-api-staging-endpoint-update"
freshness: "2026-01-24T02:01:26Z"
tags:
  - "configuration"
  - "cors"
  - "hub-api"
  - "portal-v2"
  - "staging"
owners:
  - "U05A515NBFC"
  - "U0772SH7BRA"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_1388b36620108bbe70735c9be8f043ae"
  - "srcrev_1f9e4436f0a576d0fdbb0a960ccc109d"
  - "srcrev_25abfb846367a14138dde157391c9c2d"
  - "srcrev_7bb6730f6316dfca30967dd5b0c7276a"
  - "srcrev_7efba0393ecb3aa73705d1be9f3f7c1a"
  - "srcrev_a6bba27298ceb138b672b3591c0138e6"
  - "srcrev_c05c503622c3adc4e4770f69951837b0"
conflict_state: "none"
---

# Hub API Staging Endpoint Update

## Summary

Resolved CORS and 401 issues on staging portal by updating the hub API endpoint from edge.stg.storyprotocol.net to staging-api.storyprotocol.net and adjusting dynamic environment variables.

## Claims

- A CORS issue was reported when accessing https://portal-v2-staging.vercel.app/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_7efba0393ecb3aa73705d1be9f3f7c1a` `chunk_id=srcchunk_cdb970a49b2667066e0dfd33a78750bd` `native_locator=slack:C0547N89JUB:1769130253.706209:1769130253.706209` `source_timestamp=2026-01-23T01:04:13Z`
- The specific endpoint being accessed was https://edge.stg.storyprotocol.net/hub/. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_25abfb846367a14138dde157391c9c2d` `chunk_id=srcchunk_70085ace57c6ed97159be19a518ef750` `native_locator=slack:C0547N89JUB:1769130253.706209:1769130377.892099` `source_timestamp=2026-01-23T01:06:17Z`
- The hub API endpoint edge.stg.storyprotocol.net was considered outdated by a team member. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_1388b36620108bbe70735c9be8f043ae` `chunk_id=srcchunk_62a4d88f8311fa0bf94787a561f49c8d` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218634.791769` `source_timestamp=2026-01-24T01:37:14Z`
- The staging API endpoint is https://staging-api.storyprotocol.net. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_1f9e4436f0a576d0fdbb0a960ccc109d` `chunk_id=srcchunk_bbcd100bc558f6ffbc032b07e23b7e63` `native_locator=slack:C0547N89JUB:1769130253.706209:1769195109.493409` `source_timestamp=2026-01-23T19:05:09Z`
- After updating the portal to use staging-api.storyprotocol.net, a 401 Unauthorized error occurred. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_c05c503622c3adc4e4770f69951837b0` `chunk_id=srcchunk_5de882f46985d4282c7723ee6d333c96` `native_locator=slack:C0547N89JUB:1769130253.706209:1769219327.667539` `source_timestamp=2026-01-24T01:48:47Z`
- The 401 error was resolved by updating dynamic environment variables. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_7bb6730f6316dfca30967dd5b0c7276a` `chunk_id=srcchunk_cd300688644a0e7a96f60518a1411885` `native_locator=slack:C0547N89JUB:1769130253.706209:1769219638.994309` `source_timestamp=2026-01-24T01:53:58Z`
- The hub API endpoint on staging is now working correctly. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_a6bba27298ceb138b672b3591c0138e6` `chunk_id=srcchunk_559451452aa64d3e20036709753c5d9f` `native_locator=slack:C0547N89JUB:1769130253.706209:1769220086.711289` `source_timestamp=2026-01-24T02:01:26Z`

## Sources

- `source_document_id`: `srcdoc_fc51226808b70b0ea393fdaceb83c658`
- `source_revision_id`: `srcrev_7efba0393ecb3aa73705d1be9f3f7c1a`
