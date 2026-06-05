---
title: "Data Rebrand and Token Migration"
type: "project"
slug: "projects/data-rebrand-and-migration"
freshness: "2026-06-05T21:14:00Z"
tags:
  - "data-network"
  - "migration"
  - "rebrand"
  - "token"
owners: []
source_revision_ids:
  - "srcrev_fec79d31f5015813a9a4d787dafa65ce"
conflict_state: "none"
---

# Data Rebrand and Token Migration

## Summary

Story is rebranding to Data, including renaming the L1 chain to Data Network, token $IP to $DATA, deploying a new ERC20 wrapper $WDATA, transitioning domains, and archiving GitHub repos. No hardfork is required; chain ID remains 1514.

## Claims

- Story Foundation becomes Data Foundation. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`
- Story, the L1 chain, becomes Data Network. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`
- Story native network token $IP renames to $DATA, the native token of Data Network. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`
- A new ERC20 wrapper of the native token will be deployed, $WDATA, at address 0xD18a56346227f25D1410F98f78234305660bB877. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`
- All domains with 'story' in them will transition to 'data' (e.g., storyrpc.io to datarpc.io) with 1 month supporting both versions. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`
- GitHub repositories piplabs/story and piplabs/story-geth will be archived; work continues in forks piplabs/data-network and piplabs/data-network-geth. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`
- No hardfork is required because the EVM execution environment (data-network-geth) does not reference the native token ticker or network name; chain ID remains 1514. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`
- Consensus client has low-visibility references to 'story' in App ID 'story-1' and Bech32 validator prefix 'story'; changes were deprioritized due to complexity and minimal usability impact. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`
- Centralized exchanges: no chain ID change (remains 1514), RPC domains change to mainnet.datanetworkrpc.io and aeneid.datanetworkrpc.io, ticker updated to $DATA. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_fec79d31f5015813a9a4d787dafa65ce` `chunk_id=srcchunk_906024a24e755913ffc2503b41755c06` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:14:00Z`

## Open Questions

- What are the exact final RPC domain names for mainnet and testnet? (source had placeholders with '??')
- Will the GitHub organization 'storyprotocol' be renamed or will repos move to a new org? (source left as '…?')

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_fec79d31f5015813a9a4d787dafa65ce`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
