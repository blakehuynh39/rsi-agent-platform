---
title: "Protocol v1.0.0-rc.1 Subgraph Indexing"
type: "system"
slug: "systems/protocol-v1-0-0-rc-1-subgraph-indexing"
freshness: "2024-04-22T03:58:00Z"
tags:
  - "entities"
  - "handlers"
  - "indexing"
  - "protocol-upgrade"
  - "subgraph"
owners: []
source_revision_ids:
  - "srcrev_4c1bd0a5b58fa83e1f83a05a96683680"
conflict_state: "none"
---

# Protocol v1.0.0-rc.1 Subgraph Indexing

## Summary

Details the contract addresses, entity changes, and handler modifications for the protocol-beta-v0 subgraph used in Protocol v1.0.0-rc.1 indexing.

## Claims

- The Protocol v1.0.0-rc.1 subgraph removes the Tag and IPRoyalty entities. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493) `source_document_id=srcdoc_64907f33f611d4da4c7c0d05bad71fea` `source_revision_id=srcrev_4c1bd0a5b58fa83e1f83a05a96683680` `chunk_id=srcchunk_5e53a8e6678ad490ca513e1a08661b10` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493` `source_timestamp=2024-04-22T03:58:00Z`
- The Protocol v1.0.0-rc.1 subgraph removes the handleTagSet, handleTagRemoved, and handleRoyaltyPolicySet handlers. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493) `source_document_id=srcdoc_64907f33f611d4da4c7c0d05bad71fea` `source_revision_id=srcrev_4c1bd0a5b58fa83e1f83a05a96683680` `chunk_id=srcchunk_5e53a8e6678ad490ca513e1a08661b10` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493` `source_timestamp=2024-04-22T03:58:00Z`
- The RoyaltyPolicy entity schema is updated to include ipRoyaltyVault, royaltyStack, targetAncestors, and targetRoyaltyAmount fields. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493) `source_document_id=srcdoc_64907f33f611d4da4c7c0d05bad71fea` `source_revision_id=srcrev_4c1bd0a5b58fa83e1f83a05a96683680` `chunk_id=srcchunk_5e53a8e6678ad490ca513e1a08661b10` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493` `source_timestamp=2024-04-22T03:58:00Z`
- The IPAsset entity schema adds metadata, childIpIds, parentIpIds, and rootIpIds fields. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493) `source_document_id=srcdoc_64907f33f611d4da4c7c0d05bad71fea` `source_revision_id=srcrev_4c1bd0a5b58fa83e1f83a05a96683680` `chunk_id=srcchunk_5e53a8e6678ad490ca513e1a08661b10` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493` `source_timestamp=2024-04-22T03:58:00Z`
- The handleRoyaltyPolicyInitialized handler removes logic related to ancestorsVault and splitClone, and adds assignment for ipRoyaltyVault. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493) `source_document_id=srcdoc_64907f33f611d4da4c7c0d05bad71fea` `source_revision_id=srcrev_4c1bd0a5b58fa83e1f83a05a96683680` `chunk_id=srcchunk_5e53a8e6678ad490ca513e1a08661b10` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493` `source_timestamp=2024-04-22T03:58:00Z`
- The deployed contract addresses include AccessController at 0x6fB5BA9A8747E897109044a1cd1192898AA384a9 and IPAssetRegistry at 0x30C89bCB41277f09b18DF0375b9438909e193bf0. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493) `source_document_id=srcdoc_64907f33f611d4da4c7c0d05bad71fea` `source_revision_id=srcrev_4c1bd0a5b58fa83e1f83a05a96683680` `chunk_id=srcchunk_5e53a8e6678ad490ca513e1a08661b10` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493` `source_timestamp=2024-04-22T03:58:00Z`

## Sources

- `source_document_id`: `srcdoc_64907f33f611d4da4c7c0d05bad71fea`
- `source_revision_id`: `srcrev_4c1bd0a5b58fa83e1f83a05a96683680`
- `source_url`: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-1-indexing-6f3ccd4756584bd59bdfed0cb513d493)
