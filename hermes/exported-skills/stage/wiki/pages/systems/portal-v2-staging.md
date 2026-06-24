---
title: "Portal V2 Staging Environment"
type: "system"
slug: "systems/portal-v2-staging"
freshness: "2026-01-24T02:01:26Z"
tags:
  - "cors"
  - "hub-api"
  - "portal"
  - "staging"
owners:
  - "U04L0DD6B6F"
  - "U0772SH7BRA"
  - "U08332YRB7W"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_1388b36620108bbe70735c9be8f043ae"
  - "srcrev_3d67e52c423e8c409854ed6f282e8b59"
  - "srcrev_55556eafadbb5670e3c654683b45a6c2"
  - "srcrev_641c23d2e98acb1aeaa1e81c0a35eeb4"
  - "srcrev_7bb6730f6316dfca30967dd5b0c7276a"
  - "srcrev_7efba0393ecb3aa73705d1be9f3f7c1a"
  - "srcrev_a6bba27298ceb138b672b3591c0138e6"
  - "srcrev_c05c503622c3adc4e4770f69951837b0"
conflict_state: "none"
---

# Portal V2 Staging Environment

## Summary

Details of the Portal V2 staging deployment, its authentication, and API endpoint configuration.

## Claims

- The Portal V2 staging app URL is https://portal-v2-staging.vercel.app/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_7efba0393ecb3aa73705d1be9f3f7c1a` `chunk_id=srcchunk_cdb970a49b2667066e0dfd33a78750bd` `native_locator=slack:C0547N89JUB:1769130253.706209:1769130253.706209` `source_timestamp=2026-01-23T01:04:13Z`
- The staging portal is password-protected; the password is 'programmableOG'. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_3d67e52c423e8c409854ed6f282e8b59` `chunk_id=srcchunk_b878407ae207e2d1b8c91c51ecc4c408` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218333.198919` `source_timestamp=2026-01-24T01:36:19Z`
- The staging portal accesses the Hub API for user profile data, originally using the endpoint https://edge.stg.storyprotocol.net/hub/users/v1/profile. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_55556eafadbb5670e3c654683b45a6c2` `chunk_id=srcchunk_09f3c23c8094362be338da3bb54f9502` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218381.841619` `source_timestamp=2026-01-24T01:33:01Z`
- The endpoint edge.stg.storyprotocol.net is outdated; the correct staging API endpoint is https://staging-api.storyprotocol.net. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_1388b36620108bbe70735c9be8f043ae` `chunk_id=srcchunk_62a4d88f8311fa0bf94787a561f49c8d` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218634.791769` `source_timestamp=2026-01-24T01:37:14Z`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_641c23d2e98acb1aeaa1e81c0a35eeb4` `chunk_id=srcchunk_df6b07fde6af74574097f2a3b0d004d9` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218625.699329` `source_timestamp=2026-01-24T01:37:05Z`
- After updating the endpoint to staging-api.storyprotocol.net, the portal returned 401 Unauthorized, which was resolved by updating dynamic environment variables. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_c05c503622c3adc4e4770f69951837b0` `chunk_id=srcchunk_5de882f46985d4282c7723ee6d333c96` `native_locator=slack:C0547N89JUB:1769130253.706209:1769219327.667539` `source_timestamp=2026-01-24T01:48:47Z`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_7bb6730f6316dfca30967dd5b0c7276a` `chunk_id=srcchunk_cd300688644a0e7a96f60518a1411885` `native_locator=slack:C0547N89JUB:1769130253.706209:1769219638.994309` `source_timestamp=2026-01-24T01:53:58Z`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_a6bba27298ceb138b672b3591c0138e6` `chunk_id=srcchunk_559451452aa64d3e20036709753c5d9f` `native_locator=slack:C0547N89JUB:1769130253.706209:1769220086.711289` `source_timestamp=2026-01-24T02:01:26Z`
- The production (mainnet) hub API endpoint is separate from staging; the user was testing the production environment while staging worked fine. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_3d67e52c423e8c409854ed6f282e8b59` `chunk_id=srcchunk_b878407ae207e2d1b8c91c51ecc4c408` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218333.198919` `source_timestamp=2026-01-24T01:36:19Z`
- The initial CORS issue was thought to involve an nginx server, but the root cause was an outdated API endpoint. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_7efba0393ecb3aa73705d1be9f3f7c1a` `chunk_id=srcchunk_cdb970a49b2667066e0dfd33a78750bd` `native_locator=slack:C0547N89JUB:1769130253.706209:1769130253.706209` `source_timestamp=2026-01-23T01:04:13Z`

## Open Questions

- What are the exact dynamic environment variables required for the portal?
- What is the correct production (mainnet) hub API endpoint?

## Sources

- `source_document_id`: `srcdoc_fc51226808b70b0ea393fdaceb83c658`
- `source_revision_id`: `srcrev_55556eafadbb5670e3c654683b45a6c2`
