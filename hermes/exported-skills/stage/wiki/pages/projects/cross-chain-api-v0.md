---
title: "Cross-chain API v0 Design"
type: "project"
slug: "projects/cross-chain-api-v0"
freshness: "2024-11-21T14:30:00Z"
tags:
  - "api"
  - "cross-chain"
  - "nft"
  - "story-l1"
owners: []
source_revision_ids:
  - "srcrev_ab684e0da414bd7a3f3d9e7414fd19bc"
conflict_state: "none"
---

# Cross-chain API v0 Design

## Summary

Design of v0 API for cross-chain IP registration, royalty claiming, and derivative IP registration, enabling NFT owners on other chains to interact with Story L1.

## Claims

- Initial proposal was 'Web3 cross-chain Solution'. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- With tech pivot to focus on L1 development, the team decided to support cross-chain use cases where NFT owners on another chain A can register their NFTs as IPA on Story L1 with their chain A wallet, and claim royalty payment in chain A’s native tokens. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- IPA, License Token and Royalty Token will remain on Story L1. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- An API solution for this cross-chain use case is the best option at the moment, providing better user experience and requiring less engineering resource, despite sacrificing some decentralization and having security risks. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- v0 API covers: cross-chain IP Registration with license term configurations (with metadata), cross-chain royalty claiming, cross-chain derivative IP registration (with no minting fee), register license term and attach to IP. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- v0 does not cover: cross-chain mint license token, royalty token transfer, cross-chain pay royalty, as their use cases are not clearly defined in cross-chain context and require additional engineering effort. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- The protocol will whitelist the API’s wallet to allow API to do any protocol operations without owning the actual NFT. The protocol admin will set the whitelists. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- User wallet authentication: user signs the request message using its wallet’s private key; API verifies the signature. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- App authentication: apps must register and get an API key; API verifies the key. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- NFT owner verification: API calls the NFT contract directly (fastest, most reliable) and uses NFT indexers (SimpleHash, reservoir) as backup. SimpleHash worst case freshness is 15s, not ideal. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`
- API wallet authentication: API needs its own identity so the protocol recognizes and whitelists it for managing cross-chain IP Account. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49) `source_document_id=srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff` `source_revision_id=srcrev_ab684e0da414bd7a3f3d9e7414fd19bc` `chunk_id=srcchunk_cf3a6878455853b8949e2104bff43448` `native_locator=https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49` `source_timestamp=2024-11-21T14:30:00Z`

## Open Questions

- solana-nft-mapping-to-evm-address

## Related Pages

- `api-based-cross-chain-solution-decision`

## Sources

- `source_document_id`: `srcdoc_bbf39ccc2bbb8e9983b30aa4a180a3ff`
- `source_revision_id`: `srcrev_ab684e0da414bd7a3f3d9e7414fd19bc`
- `source_url`: [Notion source](https://www.notion.so/xchain-API-v0-design-d8ee95612756478883902d58646eca49)
