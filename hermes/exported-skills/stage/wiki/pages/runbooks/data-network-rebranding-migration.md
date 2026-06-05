---
title: "Data Network Rebranding and Migration"
type: "runbook"
slug: "runbooks/data-network-rebranding-migration"
freshness: "2026-06-05T22:35:00Z"
tags:
  - "data"
  - "data-token"
  - "ip"
  - "migration"
  - "rebranding"
  - "story"
  - "token"
owners: []
source_revision_ids:
  - "srcrev_61954d360be86f07c31a7e0ec8639ab4"
conflict_state: "none"
---

# Data Network Rebranding and Migration

## Summary

The Story ecosystem is rebranding to Data, with the L1 chain becoming Data Network, token $IP becoming $DATA, and a new ERC20 wrapper $WDATA deployed. RPC and other infrastructure domains are transitioning. This page outlines the required technical actions for exchanges, wallets, and integrators.

## Claims

- Story ecosystem rebrands to Data: Story Foundation becomes Data Foundation, Story (L1) becomes Data Network, native token $IP becomes $DATA, and a new ERC20 wrapper $WDATA is deployed. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_56c0e28b5f20c5e7250fa8071ac28967` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:35:00Z`
- WDATA contract address is 0xD18a56346227f25D1410F98f78234305660bB877. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_56c0e28b5f20c5e7250fa8071ac28967` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:35:00Z`
- RPC domains transition from storyrpc.io to datanetworkrpc.io with one month supporting both versions. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_56c0e28b5f20c5e7250fa8071ac28967` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:35:00Z`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_be62a9756773eda38e38899369a0c910` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-05T22:35:00Z`
- GitHub repositories piplabs/story and piplabs/story-geth are archived; development continues in piplabs/data-network and piplabs/data-network-geth respectively. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_56c0e28b5f20c5e7250fa8071ac28967` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:35:00Z`
- Chain ID remains 1514; no hard fork executed; changes to consensus client values (app id 'story-1', validator prefix 'story') are deprioritized due to low visibility and complexity; may be included in a future hard fork with community notice. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_56c0e28b5f20c5e7250fa8071ac28967` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:35:00Z`
- WDATA is an independently deployed ERC20 contract; WIP is immutable; both tokens coexist as native token wrappers without breaking changes. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_56c0e28b5f20c5e7250fa8071ac28967` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:35:00Z`
- WDATA can be bridged to BSC via a new Layer Zero bridge; contract addresses are not yet provided. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_be62a9756773eda38e38899369a0c910` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-05T22:35:00Z`
- Balances of native tokens are not affected by the rebranding. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_be62a9756773eda38e38899369a0c910` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-05T22:35:00Z`
- Wallets and frontends must update token ticker to $DATA, network name to Data Network, graphics, RPC domains, and block explorer domain. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_be62a9756773eda38e38899369a0c910` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-05T22:35:00Z`
- Integrations of Proof of Creativity protocol and IP Portal require no changes; they continue to support WIP. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_61954d360be86f07c31a7e0ec8639ab4` `chunk_id=srcchunk_be62a9756773eda38e38899369a0c910` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-05T22:35:00Z`

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_61954d360be86f07c31a7e0ec8639ab4`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
