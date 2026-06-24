---
title: "CORS Issue Troubleshooting for portal-v2-staging"
type: "runbook"
slug: "runbooks/portal-v2-staging-cors-issue"
freshness: "2026-01-24T02:01:26Z"
tags:
  - "api"
  - "cors"
  - "portal-v2"
  - "staging"
  - "storyprotocol"
  - "vercel"
owners:
  - "U05A515NBFC"
  - "U0772SH7BRA"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_1388b36620108bbe70735c9be8f043ae"
  - "srcrev_1f9e4436f0a576d0fdbb0a960ccc109d"
  - "srcrev_25abfb846367a14138dde157391c9c2d"
  - "srcrev_3d67e52c423e8c409854ed6f282e8b59"
  - "srcrev_55556eafadbb5670e3c654683b45a6c2"
  - "srcrev_641c23d2e98acb1aeaa1e81c0a35eeb4"
  - "srcrev_7bb6730f6316dfca30967dd5b0c7276a"
  - "srcrev_7efba0393ecb3aa73705d1be9f3f7c1a"
  - "srcrev_a6bba27298ceb138b672b3591c0138e6"
  - "srcrev_c05c503622c3adc4e4770f69951837b0"
  - "srcrev_f3808a45e912e86847fc3387ac0a4472"
conflict_state: "none"
---

# CORS Issue Troubleshooting for portal-v2-staging

## Summary

Troubleshooting guide for CORS issues encountered when accessing https://portal-v2-staging.vercel.app/. The root cause was the use of the outdated endpoint https://edge.stg.storyprotocol.net/hub/, which should be replaced with https://staging-api.storyprotocol.net. After the change, a 401 error occurred until dynamic environment variables were updated. The staging portal is password-protected (password: `programmableOG`).

## Claims

- The staging portal at https://portal-v2-staging.vercel.app/ is protected by the password 'programmableOG'. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_3d67e52c423e8c409854ed6f282e8b59` `chunk_id=srcchunk_b878407ae207e2d1b8c91c51ecc4c408` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218333.198919` `source_timestamp=2026-01-24T01:36:19Z`
- Accessing https://edge.stg.storyprotocol.net/hub/ from the staging portal causes a CORS error. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_7efba0393ecb3aa73705d1be9f3f7c1a` `chunk_id=srcchunk_cdb970a49b2667066e0dfd33a78750bd` `native_locator=slack:C0547N89JUB:1769130253.706209:1769130253.706209` `source_timestamp=2026-01-23T01:04:13Z`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_25abfb846367a14138dde157391c9c2d` `chunk_id=srcchunk_70085ace57c6ed97159be19a518ef750` `native_locator=slack:C0547N89JUB:1769130253.706209:1769130377.892099` `source_timestamp=2026-01-23T01:06:17Z`
- The endpoint edge.stg.storyprotocol.net is considered outdated. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_1388b36620108bbe70735c9be8f043ae` `chunk_id=srcchunk_62a4d88f8311fa0bf94787a561f49c8d` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218634.791769` `source_timestamp=2026-01-24T01:37:14Z`
- The recommended staging API endpoint is https://staging-api.storyprotocol.net. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_1f9e4436f0a576d0fdbb0a960ccc109d` `chunk_id=srcchunk_bbcd100bc558f6ffbc032b07e23b7e63` `native_locator=slack:C0547N89JUB:1769130253.706209:1769195109.493409` `source_timestamp=2026-01-23T19:05:09Z`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_641c23d2e98acb1aeaa1e81c0a35eeb4` `chunk_id=srcchunk_df6b07fde6af74574097f2a3b0d004d9` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218625.699329` `source_timestamp=2026-01-24T01:37:05Z`
- Switching to https://staging-api.storyprotocol.net initially resulted in a 401 Unauthorized error. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_c05c503622c3adc4e4770f69951837b0` `chunk_id=srcchunk_5de882f46985d4282c7723ee6d333c96` `native_locator=slack:C0547N89JUB:1769130253.706209:1769219327.667539` `source_timestamp=2026-01-24T01:48:47Z`
- Updating the dynamic environment variables resolved the 401 error, and the endpoint worked afterwards. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_7bb6730f6316dfca30967dd5b0c7276a` `chunk_id=srcchunk_cd300688644a0e7a96f60518a1411885` `native_locator=slack:C0547N89JUB:1769130253.706209:1769219638.994309` `source_timestamp=2026-01-24T01:53:58Z`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_a6bba27298ceb138b672b3591c0138e6` `chunk_id=srcchunk_559451452aa64d3e20036709753c5d9f` `native_locator=slack:C0547N89JUB:1769130253.706209:1769220086.711289` `source_timestamp=2026-01-24T02:01:26Z`
- The user was attempting to test the production (mainnet) environment, but the staging (aeneid) environment was already working. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_3d67e52c423e8c409854ed6f282e8b59` `chunk_id=srcchunk_b878407ae207e2d1b8c91c51ecc4c408` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218333.198919` `source_timestamp=2026-01-24T01:36:19Z`
- The current production backend API is at https://api.storyapis.com/api/v4/assets. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_f3808a45e912e86847fc3387ac0a4472` `chunk_id=srcchunk_e0277dc93cf0dac5434dfabbabc98f2e` `native_locator=slack:C0547N89JUB:1769130253.706209:1769195072.462509` `source_timestamp=2026-01-23T19:04:32Z`
- The hub API endpoint is used for user profile data, e.g., /hub/users/v1/profile. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_55556eafadbb5670e3c654683b45a6c2` `chunk_id=srcchunk_09f3c23c8094362be338da3bb54f9502` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218381.841619` `source_timestamp=2026-01-24T01:33:01Z`

## Open Questions

- What is the current CORS configuration for the nginx server mentioned?
- What specific dynamic environment variable was updated to fix the 401 error?

## Sources

- `source_document_id`: `srcdoc_fc51226808b70b0ea393fdaceb83c658`
- `source_revision_id`: `srcrev_1f9e4436f0a576d0fdbb0a960ccc109d`
