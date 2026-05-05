---
title: "Deployment Chain Decision"
type: "decision"
slug: "decisions/deployment-chain-decision"
freshness: "2026-05-05T06:28:33Z"
tags:
  - "deployment"
  - "Ethereum"
  - "L2"
  - "LayerZero"
  - "NFT"
owners: []
source_revision_ids:
  - "srcrev_9d33a2a39706496da724e1ec6941897f"
conflict_state: "none"
---

# Deployment Chain Decision

## Summary

Decision analysis for deploying the system on Layer 2 to mitigate high Ethereum gas costs, including options for mirroring L1 NFTs to L2.

## Claims

- Ethereum gas costs are prohibitively high: a simple USDC transfer costs $10 and swaps cost >$20 on average. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f) `source_document_id=srcdoc_731a2d8705bd9c1453eecbea1c8b67d1` `source_revision_id=srcrev_9d33a2a39706496da724e1ec6941897f` `chunk_id=srcchunk_85f49c59132fad364f6bd5078df4993b` `native_locator=https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f` `source_timestamp=2026-05-05T06:28:33Z`
- It is practically impossible to enable the creation of millions of IPAs on Ethereum without sacrifices like off-chain merkle trees for compressed NFTs. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f) `source_document_id=srcdoc_731a2d8705bd9c1453eecbea1c8b67d1` `source_revision_id=srcrev_9d33a2a39706496da724e1ec6941897f` `chunk_id=srcchunk_85f49c59132fad364f6bd5078df4993b` `native_locator=https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f` `source_timestamp=2026-05-05T06:28:33Z`
- The solution is to deploy on L2, especially since EIP-4844 will make L2s much cheaper. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f) `source_document_id=srcdoc_731a2d8705bd9c1453eecbea1c8b67d1` `source_revision_id=srcrev_9d33a2a39706496da724e1ec6941897f` `chunk_id=srcchunk_85f49c59132fad364f6bd5078df4993b` `native_locator=https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f` `source_timestamp=2026-05-05T06:28:33Z`
- L2 choices are: 1) Directly on Optimism, Arbitrum, etc. (maximal composability, zero layer-level flexibility); 2) Our own L2 via Optimism, etc. (maximal flexibility & revenue, more native token use cases, minimal composability). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f) `source_document_id=srcdoc_731a2d8705bd9c1453eecbea1c8b67d1` `source_revision_id=srcrev_9d33a2a39706496da724e1ec6941897f` `chunk_id=srcchunk_85f49c59132fad364f6bd5078df4993b` `native_locator=https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f` `source_timestamp=2026-05-05T06:28:33Z`
- One approach to capture Ethereum community on L2 is to mirror L1 NFTs to L2 via LayerZero's Generic Message Passing (GMP) without locking the NFTs in a contract. Pros: No need to lock NFTs, intuitive for holders. Cons: Sync time delay between L1 and L2 leaves a window for old owners to exploit the system. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f) `source_document_id=srcdoc_731a2d8705bd9c1453eecbea1c8b67d1` `source_revision_id=srcrev_9d33a2a39706496da724e1ec6941897f` `chunk_id=srcchunk_85f49c59132fad364f6bd5078df4993b` `native_locator=https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f` `source_timestamp=2026-05-05T06:28:33Z`
- Another approach is to mirror L1 NFTs to L2 via GMP, requiring locking/staking of the NFTs on Ethereum. Pros: Any action on the NFT requires unlocking, allowing state change reflection on L2 before L1. Cons: Need to lock NFTs in L1 contract, a huge concern for many owners. Incentivizing users to lock with SP tokens is suggested. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f) `source_document_id=srcdoc_731a2d8705bd9c1453eecbea1c8b67d1` `source_revision_id=srcrev_9d33a2a39706496da724e1ec6941897f` `chunk_id=srcchunk_85f49c59132fad364f6bd5078df4993b` `native_locator=https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f` `source_timestamp=2026-05-05T06:28:33Z`

## Open Questions

- How do we capture the Ethereum community on L2?

## Sources

- `source_document_id`: `srcdoc_731a2d8705bd9c1453eecbea1c8b67d1`
- `source_revision_id`: `srcrev_9d33a2a39706496da724e1ec6941897f`
- `source_url`: [Notion source](https://www.notion.so/WIP-Deployment-Chain-e7d74ea7af7f4f76be39b3493dcc2a4f)
