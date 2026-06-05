---
title: "Story Rebrands to Data (2026 Decision)"
type: "decision"
slug: "decisions/story-rebrand-to-data-decision"
freshness: "2026-06-05T20:29:00Z"
tags:
  - "chain"
  - "migration"
  - "rebranding"
  - "token"
owners:
  - "Data Foundation"
source_revision_ids:
  - "srcrev_e44fda018dcb769685df1dc84fbd0b61"
conflict_state: "none"
---

# Story Rebrands to Data (2026 Decision)

## Summary

Story Protocol has decided to rebrand to 'Data' across the ecosystem: renaming the L1 chain to Data Network, the native token from $IP to $DATA, and introducing an ERC20 wrapper $WDATA. Technical details and required exchange actions are provided.

## Claims

- Story will rebrand to Data, including renaming Story Foundation to Data Foundation, Story L1 chain to Data Network, and native token $IP to $DATA. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_e44fda018dcb769685df1dc84fbd0b61` `chunk_id=srcchunk_9307beb56eb6a6b62fa773e9e38d1733` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T20:29:00Z`
- A new ERC20 wrapper of the native token will be deployed, $WDATA, with address 0xD18a56346227f25D1410F98f78234305660bB877. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_e44fda018dcb769685df1dc84fbd0b61` `chunk_id=srcchunk_9307beb56eb6a6b62fa773e9e38d1733` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T20:29:00Z`
- Domains containing 'story' (like storyrpc.io) will transition to 'data' (like datarpc.io) with one month of dual support both versions. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_e44fda018dcb769685df1dc84fbd0b61` `chunk_id=srcchunk_9307beb56eb6a6b62fa773e9e38d1733` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T20:29:00Z`
- GitHub repositories under storyprotocol org will move; network client repos story and story-geth will be archived, continuing in forks data-network and data-network-geth. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_e44fda018dcb769685df1dc84fbd0b61` `chunk_id=srcchunk_9307beb56eb6a6b62fa773e9e38d1733` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T20:29:00Z`
- Chain ID remains 1514 and no hardfork is required because the EVM execution environment doesn't reference the token ticker or network name, and low visibility consensus client references (App Id story-1, validator prefix story) are not changed. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_e44fda018dcb769685df1dc84fbd0b61` `chunk_id=srcchunk_9307beb56eb6a6b62fa773e9e38d1733` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T20:29:00Z`
- The legacy WIP token is immutable, so WDATA and WIP will coexist as functionally identical native token wrappers without breaking changes. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_e44fda018dcb769685df1dc84fbd0b61` `chunk_id=srcchunk_9307beb56eb6a6b62fa773e9e38d1733` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T20:29:00Z`
- Centralized exchanges need to update RPC endpoint from mainnet.storyrpc.io to mainnet.datanetworkrpc.io, update ticker to $DATA, and network name to Data Network, while chain ID remains 1514. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_e44fda018dcb769685df1dc84fbd0b61` `chunk_id=srcchunk_9307beb56eb6a6b62fa773e9e38d1733` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T20:29:00Z`

## Open Questions

- Exact new RPC domains for mainnet and testnet are not yet finalized (placeholders shown).
- New GitHub org for repositories under `storyprotocol` is undecided.

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_e44fda018dcb769685df1dc84fbd0b61`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
