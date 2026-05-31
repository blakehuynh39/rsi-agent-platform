---
title: "Infrastructure Tradeoffs"
type: "concept"
slug: "concepts/infrastructure-tradeoffs"
freshness: "2024-12-03T05:20:00Z"
tags:
  - "infrastructure"
  - "L1"
  - "L2"
  - "mainnet"
  - "protocol-design"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_bdb692eea0bee39414df2dcad969548c"
conflict_state: "none"
---

# Infrastructure Tradeoffs

## Summary

Discussion on testnet vs mainnet launch and L1 vs L2 deployment for the Story Protocol, including considerations for value accrual, data storage, and interoperability.

## Claims

- LC is more inclined to launch the v1 protocol on mainnet together with WF by the end of the year, citing flexibility from testnet but potential messaging failure if not on mainnet. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab) `source_document_id=srcdoc_564f2c346b555e518859caee696b1a1f` `source_revision_id=srcrev_bdb692eea0bee39414df2dcad969548c` `chunk_id=srcchunk_8d24d2fbda4c442521e4317a3d62906c` `native_locator=https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab` `source_timestamp=2024-12-03T05:20:00Z`
- A proposed architecture involves value-bearing ERC-721 assets on mainnet and data model contracts (Character metadata / Franchise contracts) on L2 to balance value accrual and data storage. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab) `source_document_id=srcdoc_564f2c346b555e518859caee696b1a1f` `source_revision_id=srcrev_bdb692eea0bee39414df2dcad969548c` `chunk_id=srcchunk_8d24d2fbda4c442521e4317a3d62906c` `native_locator=https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab` `source_timestamp=2024-12-03T05:20:00Z`
- LC expressed concerns about L2 interoperability with other protocols and the unsettled L2 landscape, stating Ethereum L1 is the safest bet for the protocol at the moment. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab) `source_document_id=srcdoc_564f2c346b555e518859caee696b1a1f` `source_revision_id=srcrev_bdb692eea0bee39414df2dcad969548c` `chunk_id=srcchunk_8d24d2fbda4c442521e4317a3d62906c` `native_locator=https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab` `source_timestamp=2024-12-03T05:20:00Z`
- RM suggested aiming for WF on Mainnet, then L2, with L2 controlling L1 collections as an option, and proposed Polygon zkEVM as a potential L2 partner. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab) `source_document_id=srcdoc_564f2c346b555e518859caee696b1a1f` `source_revision_id=srcrev_bdb692eea0bee39414df2dcad969548c` `chunk_id=srcchunk_8d24d2fbda4c442521e4317a3d62906c` `native_locator=https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab` `source_timestamp=2024-12-03T05:20:00Z`

## Open Questions

- Should the protocol deploy on L1, L2, or a hybrid architecture?
- Should the protocol launch on testnet or mainnet for v1?
- Which L2 solution (Optimistic, zkEVM, Polygon zkEVM) is most suitable?

## Related Pages

- `concepts/protocol-home`

## Sources

- `source_document_id`: `srcdoc_564f2c346b555e518859caee696b1a1f`
- `source_revision_id`: `srcrev_bdb692eea0bee39414df2dcad969548c`
- `source_url`: [Notion source](https://www.notion.so/Infrastructure-Tradeoffs-87f1d103c2c6489f9710b48df847c3ab)
