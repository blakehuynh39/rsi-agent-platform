---
title: "story-api POST /api/v4/collections 500 Error"
type: "system"
slug: "systems/story-api-collections-post-500-error"
freshness: "2026-03-03T01:51:56Z"
tags:
  - "api"
  - "collections"
  - "error"
  - "incident"
  - "story-api"
owners: []
source_revision_ids:
  - "srcrev_c042cace736e1d5f420c9f5d0c387c17"
  - "srcrev_d0986934ab24544a25b5ad7737b1de48"
conflict_state: "none"
---

# story-api POST /api/v4/collections 500 Error

## Summary

The story-api endpoint POST /api/v4/collections returned a 500 Internal Server Error on at least two occasions: around 2026-02-18 and 2026-02-24.

## Claims

- POST /api/v4/collections failed with HTTP 500 Internal Server Error. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0ff7bbfedaad80575ff1ad83e22df869` `source_revision_id=srcrev_d0986934ab24544a25b5ad7737b1de48` `chunk_id=srcchunk_d85bea75b0277745bf0bc1246c3bc241` `native_locator=slack:C07K3J4JTH6:1771978942.017629:1771978942.017629` `source_timestamp=2026-02-25T00:22:22Z`
  - citation: `source_document_id=srcdoc_0ff7bbfedaad80575ff1ad83e22df869` `source_revision_id=srcrev_c042cace736e1d5f420c9f5d0c387c17` `chunk_id=srcchunk_cb4c54d7b5486b63fa6e36f540ab545c` `native_locator=slack:C07K3J4JTH6:1771978942.017629:1772502716.846319` `source_timestamp=2026-03-03T01:51:56Z`

## Open Questions

- Has the issue been resolved?
- What caused the 500 error?

## Sources

- `source_document_id`: `srcdoc_0ff7bbfedaad80575ff1ad83e22df869`
- `source_revision_id`: `srcrev_d0986934ab24544a25b5ad7737b1de48`
