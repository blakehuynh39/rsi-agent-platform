---
title: "Unified IP Account"
type: "concept"
slug: "concepts/unified-ip-account"
freshness: "2026-01-30T20:26:00Z"
tags:
  - "ERC-6551"
  - "IP Account"
  - "IP NFT"
  - "Story Protocol"
owners: []
source_revision_ids:
  - "srcrev_8799f499196e5ea2c64ba7f5a81b0bd1"
conflict_state: "none"
---

# Unified IP Account

## Summary

Unified IP Accounts provide a contract-based representation of IP, enabling unified handling of relationships between IP NFTs and SP Assets for both new collections and existing external PFPs.

## Claims

- IP NFT stands for the identity of IP, analogous to a passport. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_944581bd7ed431e7b761f283c5ef899e` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1` `source_timestamp=2026-01-30T20:26:00Z`
- IP Account represents the IP as a real entity, analogous to a person. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_944581bd7ed431e7b761f283c5ef899e` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1` `source_timestamp=2026-01-30T20:26:00Z`
- Ownership of the IP NFT confers ownership of the IP Account. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_944581bd7ed431e7b761f283c5ef899e` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1` `source_timestamp=2026-01-30T20:26:00Z`
- IPAccount owns derived SP Assets. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_944581bd7ed431e7b761f283c5ef899e` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1` `source_timestamp=2026-01-30T20:26:00Z`
- Unified IP Accounts support transferring ownership of SP Assets along with transferring ownership of an external PFP NFT. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_944581bd7ed431e7b761f283c5ef899e` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1` `source_timestamp=2026-01-30T20:26:00Z`
- IP Account is a first-class citizen contract on Ethereum that can hold ERC-20/ETH tokens as royalties, integrate with other protocols, and build an IP graph. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_944581bd7ed431e7b761f283c5ef899e` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1` `source_timestamp=2026-01-30T20:26:00Z`
- IPAccount can be lazily created on demand because its address can be computed without deployment. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_944581bd7ed431e7b761f283c5ef899e` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-1` `source_timestamp=2026-01-30T20:26:00Z`
- IPAccount can be deployed by any third-party ERC-6551 Registry as long as it uses the designated account implementation address. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-2) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_6fd871bdfc7eaba2d9ffb01d1e2fc513` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-2` `source_timestamp=2026-01-30T20:26:00Z`
- IPAccount address can be derived from any third-party ERC-6551 Registry using the designated account implementation address. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-2) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_6fd871bdfc7eaba2d9ffb01d1e2fc513` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-2` `source_timestamp=2026-01-30T20:26:00Z`
- Future gas optimization may embed the Beacon/Hub address into IPAccount to eliminate a second delegateCall. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-2) `source_document_id=srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5` `source_revision_id=srcrev_8799f499196e5ea2c64ba7f5a81b0bd1` `chunk_id=srcchunk_6fd871bdfc7eaba2d9ffb01d1e2fc513` `native_locator=https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7#chunk-2` `source_timestamp=2026-01-30T20:26:00Z`

## Sources

- `source_document_id`: `srcdoc_ee5b92f3cba78c6a41db359bacf1d1b5`
- `source_revision_id`: `srcrev_8799f499196e5ea2c64ba7f5a81b0bd1`
- `source_url`: [Notion source](https://www.notion.so/Unified-IP-Account-1d4101e5d54b441ca218b5302e786ed7)
