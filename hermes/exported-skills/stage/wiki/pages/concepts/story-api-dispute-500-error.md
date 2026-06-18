---
title: "Story-API Dispute Endpoint 500 Error"
type: "concept"
slug: "concepts/story-api-dispute-500-error"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "500-error"
  - "disputes"
  - "story-api"
owners:
  - "Blake Huynh"
source_revision_ids:
  - "srcrev_8658552cba5079f2848dace5e1dd800e"
  - "srcrev_d2252e01d6a49d2fa30a836da394304e"
conflict_state: "none"
---

# Story-API Dispute Endpoint 500 Error

## Summary

On 2026-06-16, the story-api POST /api/v4/disputes endpoint returned a 500 Internal Server Error. The issue was resolved by Blake Huynh.

## Claims

- The story-api POST /api/v4/disputes endpoint returned a 500 Internal Server Error. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c6122f7d06f0601db14b70c562c40e14` `source_revision_id=srcrev_d2252e01d6a49d2fa30a836da394304e` `chunk_id=srcchunk_45f84e6b238d4d123ae261438e27df49` `native_locator=slack:C07K3J4JTH6:1781411697.987919:1781411697.987919` `source_timestamp=2026-06-14T04:34:57Z`
- Blake Huynh resolved the issue (STORY-API-F1). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c6122f7d06f0601db14b70c562c40e14` `source_revision_id=srcrev_8658552cba5079f2848dace5e1dd800e` `chunk_id=srcchunk_eba17477f29af56792beec303427a659` `native_locator=slack:C07K3J4JTH6:1781411697.987919:1781630303.109329` `source_timestamp=2026-06-16T17:18:23Z`

## Open Questions

- What was the root cause of the 500 error on the disputes endpoint?

## Sources

- `source_document_id`: `srcdoc_c6122f7d06f0601db14b70c562c40e14`
- `source_revision_id`: `srcrev_8658552cba5079f2848dace5e1dd800e`
