---
title: "Explorer Use Case"
type: "concept"
slug: "concepts/explorer-use-case"
freshness: "2024-04-12T18:30:00Z"
tags:
  - "api"
  - "explorer"
  - "mainnet"
  - "use-case"
owners: []
source_revision_ids:
  - "srcrev_3d26cefb46d38234ae94bff498d4666e"
conflict_state: "none"
---

# Explorer Use Case

## Summary

Describes how the Explorer application leverages the Mainnet API, filtering by IPA id and providing search, transaction, IPAsset, lineage, collection, derivative, policy, license, royalty, dispute, and permission views.

## Claims

- Everything is filtered by IPA id. `claim:claim_3_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- Search Bar: Search IPA id, transaction id and collection id using get functions of these 3 resources. `claim:claim_3_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- Transaction: List Transaction on Transaction page; Get transaction for transaction details. `claim:claim_3_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- IPAsset: List IPAsset for the IPAsset page; Get IPAsset for IPAsset details. `claim:claim_3_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- IPAsset lineage: childIP and parent IP from List IPAsset; isRoot? Check if rootIPAsset is empty. `claim:claim_3_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- Collection: List collection for the collection page; List IPAsset filtering by collection. `claim:claim_3_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- Derivative IPAs: Get children IP from List IPAsset. `claim:claim_3_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- Policy: Get all policies and details for an IPA. `claim:claim_3_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- License: List license based on the IPA. `claim:claim_3_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- Royalty: Royalty Token Holder for the IPA; Royalty Payment; Royalty Claiming. `claim:claim_3_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- Dispute: List dispute based on the IPA. `claim:claim_3_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`
- Permission: List Permission based on the IPA. `claim:claim_3_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80) `source_document_id=srcdoc_019e39185e81d2cd2a27cd1123dbce59` `source_revision_id=srcrev_3d26cefb46d38234ae94bff498d4666e` `chunk_id=srcchunk_448d43a7beeac5522b01cabedf12f5ed` `native_locator=https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80` `source_timestamp=2024-04-12T18:30:00Z`

## Related Pages

- `mainnet-api`

## Sources

- `source_document_id`: `srcdoc_019e39185e81d2cd2a27cd1123dbce59`
- `source_revision_id`: `srcrev_3d26cefb46d38234ae94bff498d4666e`
- `source_url`: [Notion source](https://www.notion.so/Mainnet-API-17edd4d6506141c4b5245d5c09995e80)
