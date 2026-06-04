---
title: "Story to Data Rebranding"
type: "decision"
slug: "decisions/story-to-data-rebranding"
freshness: "2026-06-04T23:05:00Z"
tags:
  - "data-network"
  - "rebranding"
  - "story"
  - "token"
owners: []
source_revision_ids:
  - "srcrev_67fba4ca3443e802fbe8d4d6be0d42de"
conflict_state: "none"
---

# Story to Data Rebranding

## Summary

Rebranding of Story Protocol to Data, including foundation, chain, token, domains, and GitHub organization.

## Claims

- Story Foundation becomes Data Foundation. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_67fba4ca3443e802fbe8d4d6be0d42de` `chunk_id=srcchunk_c8cb96a5b068ebf4ccdbd55f62b6632e` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:05:00Z`
- Story L1 chain becomes Data Network. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_67fba4ca3443e802fbe8d4d6be0d42de` `chunk_id=srcchunk_c8cb96a5b068ebf4ccdbd55f62b6632e` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:05:00Z`
- Story native network token $IP renames to $DATA, the native token of Data Network. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_67fba4ca3443e802fbe8d4d6be0d42de` `chunk_id=srcchunk_c8cb96a5b068ebf4ccdbd55f62b6632e` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:05:00Z`
- A new ERC20 wrapper of the native token will be deployed, $WDATA. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_67fba4ca3443e802fbe8d4d6be0d42de` `chunk_id=srcchunk_c8cb96a5b068ebf4ccdbd55f62b6632e` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:05:00Z`
- All domains with 'story' in them will transition to 'data' with 1 month supporting both versions (e.g., storyrpc.io → datarpc.io). `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_67fba4ca3443e802fbe8d4d6be0d42de` `chunk_id=srcchunk_c8cb96a5b068ebf4ccdbd55f62b6632e` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:05:00Z`
- GitHub repositories story and story-geth under piplabs will be archived, and work will continue in forks data-network and data-network-geth. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_67fba4ca3443e802fbe8d4d6be0d42de` `chunk_id=srcchunk_c8cb96a5b068ebf4ccdbd55f62b6632e` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:05:00Z`
- No hardfork is planned because the EVM execution environment does not reference the native token ticker or network name, the chain ID remains 1514, and the low visibility consensus client references were deprioritized due to complexity. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_67fba4ca3443e802fbe8d4d6be0d42de` `chunk_id=srcchunk_c8cb96a5b068ebf4ccdbd55f62b6632e` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:05:00Z`
- WDATA is an independently deployed ERC20 contract; WIP is immutable, so both tokens will coexist as identical native token wrappers without breaking changes or hardfork. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_67fba4ca3443e802fbe8d4d6be0d42de` `chunk_id=srcchunk_c8cb96a5b068ebf4ccdbd55f62b6632e` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:05:00Z`

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_67fba4ca3443e802fbe8d4d6be0d42de`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
