---
title: "STORY-API Fatal Stream Error"
type: "system"
slug: "systems/story-api-fatal-stream-error"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "error"
  - "incident"
  - "sentry"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
  - "romain.magne@piplabs.xyz"
source_revision_ids:
  - "srcrev_32ffbf9e3b7a4fba563a8af460f6adb4"
  - "srcrev_8b81295799f8764fa9363533cec55189"
  - "srcrev_bdb87871dd0d74fdb730e0a9d762eaff"
  - "srcrev_cabdfcb35e610fda866f0299389375ae"
  - "srcrev_f8550e9d00fc511801b90b1b88005fb5"
conflict_state: "none"
---

# STORY-API Fatal Stream Error

## Summary

The Story API repeatedly threw a fatal stream error, requiring multiple resolutions.

## Claims

- The STORY-API experienced a fatal stream error with message: '[story-api] shard task returned a fatal stream error'. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_791a98068e8f43b806246cc620f424ab` `source_revision_id=srcrev_8b81295799f8764fa9363533cec55189` `chunk_id=srcchunk_6a40620412ceee0b1959b4e34226dd89` `native_locator=slack:C07K3J4JTH6:1780882700.240239:1780882700.240239` `source_timestamp=2026-06-08T01:38:20Z`
- Romain Magne resolved the Sentry issue STORY-API-ES (issue ID 7535903686) at some point. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_791a98068e8f43b806246cc620f424ab` `source_revision_id=srcrev_cabdfcb35e610fda866f0299389375ae` `chunk_id=srcchunk_ed6eb048da3e6912db4ae57de672b0b6` `native_locator=slack:C07K3J4JTH6:1780882700.240239:1780942759.393779` `source_timestamp=2026-06-08T18:19:19Z`
- The error recurred after being resolved, as indicated by additional notifications. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_791a98068e8f43b806246cc620f424ab` `source_revision_id=srcrev_32ffbf9e3b7a4fba563a8af460f6adb4` `chunk_id=srcchunk_651b0e349e6cc60d6cbdef1a460c72ae` `native_locator=slack:C07K3J4JTH6:1780882700.240239:1781049858.555499` `source_timestamp=2026-06-10T00:04:18Z`
  - citation: `source_document_id=srcdoc_791a98068e8f43b806246cc620f424ab` `source_revision_id=srcrev_bdb87871dd0d74fdb730e0a9d762eaff` `chunk_id=srcchunk_6c78866aba450942f0dc7b01130d567f` `native_locator=slack:C07K3J4JTH6:1780882700.240239:1781411633.381049` `source_timestamp=2026-06-14T04:33:53Z`
- Blake Huynh later resolved the issue again. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_791a98068e8f43b806246cc620f424ab` `source_revision_id=srcrev_f8550e9d00fc511801b90b1b88005fb5` `chunk_id=srcchunk_c28eafe8637ac0e0259f0f6881df42cc` `native_locator=slack:C07K3J4JTH6:1780882700.240239:1781630303.466649` `source_timestamp=2026-06-16T17:18:23Z`

## Open Questions

- What is the root cause of the persistent fatal stream error in the STORY-API?

## Sources

- `source_document_id`: `srcdoc_791a98068e8f43b806246cc620f424ab`
- `source_revision_id`: `srcrev_32ffbf9e3b7a4fba563a8af460f6adb4`
