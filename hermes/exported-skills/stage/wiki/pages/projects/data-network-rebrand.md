---
title: "Story to Data Network Rebranding"
type: "project"
slug: "projects/data-network-rebrand"
freshness: "2026-06-08T18:36:00Z"
tags:
  - "data-network"
  - "rebranding"
  - "story"
  - "token-migration"
owners: []
source_revision_ids:
  - "srcrev_b60d65170bb0deed4ac30871669621c1"
conflict_state: "none"
---

# Story to Data Network Rebranding

## Summary

In June 2026, Story rebranded to Data Network. The L1 chain, native token, and all associated domains and repositories were renamed. $IP becomes $DATA, a new wrapped ERC20 token $WDATA is deployed. Infrastructure transitions with a 1-month dual support period. No hardfork is performed; chain ID remains 1514.

## Claims

- Story Foundation is renamed to Data Foundation. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- Story L1 chain is renamed to Data Network. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- Native token ticker changes from $IP to $DATA. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- A new ERC20 wrapper token $WDATA is deployed at address 0xD18a56346227f25D1410F98f78234305660bB877. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- Domains containing 'story' (e.g., storyrpc.io) will transition to 'data' (datarpc.io), with both versions supported for 1 month. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- GitHub repos piplabs/story and piplabs/story-geth will be archived, replaced by forks piplabs/data-network and piplabs/data-network-geth. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- No hardfork is performed. Chain ID remains 1514. Consensus client references to 'story' in app ID 'story-1' and bech32 prefix 'story' are unchanged. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- $IP genesis JSON reference remains to preserve chain history; changing it would produce a different app hash. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- $WDATA is an independently deployed ERC20 contract; $WIP is immutable, so both tokens coexist without breaking changes. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- Centralized exchanges must update ticker to $DATA and change RPC endpoints to new domains (exact URLs unconfirmed), no chain ID change. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- Block explorer domain is planned to change to something like https://datanetwokscan (exact URL unconfirmed). `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_e5abd880870668279d118571a0dea4bb` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T18:36:00Z`
- A new LayerZero bridge will allow bridging $WDATA from Data Network to BSC via stargate.finance, with contract addresses to be determined. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_9f452afb7ad7ec5a1bdea80b230d03aa` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T18:36:00Z`
- Wallets and frontends must update token ticker to $DATA, network name to 'Data Network', and RPC domains (e.g., to https://mainnet.datanetworkrpc.io), but chain ID (1514) and native balances remain unchanged. `claim:claim_1_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_9f452afb7ad7ec5a1bdea80b230d03aa` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T18:36:00Z`
- Proof of Creativity protocol and IP Portal integrations are unaffected; they continue to support $WIP and require no SDK changes. `claim:claim_1_14` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b60d65170bb0deed4ac30871669621c1` `chunk_id=srcchunk_9f452afb7ad7ec5a1bdea80b230d03aa` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T18:36:00Z`

## Open Questions

- Block explorer domain is unconfirmed (placeholder 'datanetwokscan??').
- Bridge contract addresses for WDATA on BSC and OTFAdapter on Data Network are not provided (0x123... placeholders).
- Domain transition list incomplete (source says 'Full list: - …').
- Exact new RPC domain URLs (mainnet and aeneid) are unconfirmed, placeholders show '??' in source.
- GitHub organization migration details (where storyprotocol repos move) are not specified.

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_b60d65170bb0deed4ac30871669621c1`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
