---
title: "Portal V2 Staging Configuration"
type: "system"
slug: "systems/portal-v2-staging-config"
freshness: "2026-01-24T02:01:26Z"
tags:
  - "api-endpoint"
  - "CORS"
  - "portal"
  - "staging"
owners: []
source_revision_ids:
  - "srcrev_1388b36620108bbe70735c9be8f043ae"
  - "srcrev_1f9e4436f0a576d0fdbb0a960ccc109d"
  - "srcrev_25abfb846367a14138dde157391c9c2d"
  - "srcrev_3d67e52c423e8c409854ed6f282e8b59"
  - "srcrev_a6bba27298ceb138b672b3591c0138e6"
conflict_state: "none"
---

# Portal V2 Staging Configuration

## Summary

Configuration details for the staging portal-v2 at https://portal-v2-staging.vercel.app, including the correct Staging API endpoint and password.

## Claims

- The staging portal URL is https://portal-v2-staging.vercel.app. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_3d67e52c423e8c409854ed6f282e8b59` `chunk_id=srcchunk_b878407ae207e2d1b8c91c51ecc4c408` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218333.198919` `source_timestamp=2026-01-24T01:36:19Z`
- The staging portal password is `programmableOG`. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_3d67e52c423e8c409854ed6f282e8b59` `chunk_id=srcchunk_b878407ae207e2d1b8c91c51ecc4c408` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218333.198919` `source_timestamp=2026-01-24T01:36:19Z`
- The correct staging API endpoint for the hub is https://staging-api.storyprotocol.net. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_1f9e4436f0a576d0fdbb0a960ccc109d` `chunk_id=srcchunk_bbcd100bc558f6ffbc032b07e23b7e63` `native_locator=slack:C0547N89JUB:1769130253.706209:1769195109.493409` `source_timestamp=2026-01-23T19:05:09Z`
- The endpoint https://edge.stg.storyprotocol.net/hub/ is outdated and was causing CORS issues. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_25abfb846367a14138dde157391c9c2d` `chunk_id=srcchunk_70085ace57c6ed97159be19a518ef750` `native_locator=slack:C0547N89JUB:1769130253.706209:1769130377.892099` `source_timestamp=2026-01-23T01:06:17Z`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_1388b36620108bbe70735c9be8f043ae` `chunk_id=srcchunk_62a4d88f8311fa0bf94787a561f49c8d` `native_locator=slack:C0547N89JUB:1769130253.706209:1769218634.791769` `source_timestamp=2026-01-24T01:37:14Z`
- Updating the portal to use the staging-api.storyprotocol.net endpoint resolved the CORS issue. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc51226808b70b0ea393fdaceb83c658` `source_revision_id=srcrev_a6bba27298ceb138b672b3591c0138e6` `chunk_id=srcchunk_559451452aa64d3e20036709753c5d9f` `native_locator=slack:C0547N89JUB:1769130253.706209:1769220086.711289` `source_timestamp=2026-01-24T02:01:26Z`

## Open Questions

- What is the correct production (mainnet) hub API endpoint for portal v2?

## Sources

- `source_document_id`: `srcdoc_fc51226808b70b0ea393fdaceb83c658`
- `source_revision_id`: `srcrev_3d67e52c423e8c409854ed6f282e8b59`
