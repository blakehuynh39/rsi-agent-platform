---
title: "Story Rebrand to Data Network"
type: "decision"
slug: "decisions/story-rebrand-to-data-network"
freshness: "2026-06-04T23:00:00Z"
tags:
  - "data-network"
  - "rebranding"
  - "story-protocol"
  - "token-migration"
owners: []
source_revision_ids:
  - "srcrev_94444f7b9fd6dbfa0f1593632b2ba796"
conflict_state: "none"
---

# Story Rebrand to Data Network

## Summary

Story protocol rebrands to Data, renaming the foundation, L1 chain, and token, while maintaining continuity without a hardfork.

## Claims

- Story will rebrand to Data. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- Story Foundation becomes Data Foundation. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- Story, the L1 chain, becomes Data Network. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- $IP token renames to $DATA, the native token of Data Network. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- A new ERC20 wrapper of the native token, $WDATA, will be deployed. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- All domains with 'story' in them will transition to 'data' with 1 month supporting both versions. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- GitHub repositories under storyprotocol org will move; network client repos (piplabs/story and piplabs/story-geth) will be archived and work continued in forks piplabs/data-network and piplabs/data-network-geth. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- No hardfork is needed: the EVM execution environment doesn't reference the native token ticker or network name, and chain ID remains 1514. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- The consensus client (Cosmos based) has references to 'story' in App Id 'story-1' and validator address; these are not changed. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`
- The genesis JSON references $IP but must be kept unchanged to preserve the app hash and allow clients to sync from block 0. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_94444f7b9fd6dbfa0f1593632b2ba796` `chunk_id=srcchunk_cf96e0ae4b4c77e955af963e20b5fead` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-04T23:00:00Z`

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_94444f7b9fd6dbfa0f1593632b2ba796`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
