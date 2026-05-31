---
title: "Web3 Cross-Chain Solution"
type: "project"
slug: "projects/web3-cross-chain-solution"
freshness: "2024-11-21T14:30:00Z"
tags:
  - "AA-wallet"
  - "cross-chain"
  - "NFT"
  - "story-protocol"
  - "web3"
owners: []
source_revision_ids:
  - "srcrev_cf079f6965434a6e7a9a8bc442ab5133"
conflict_state: "none"
---

# Web3 Cross-Chain Solution

## Summary

Strategy and design for enabling Story Protocol interactions from other networks, including EVM chains and Solana, using AA wallets and a centralized cross-chain API service.

## Claims

- The goal is to provide tech solutions to support using Story Protocol from other networks, potentially combining with the web2 API strategy. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_c2f1de82a51e902c97e21475779c4870` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1` `source_timestamp=2024-11-21T14:30:00Z`
- Interacting with Story Protocol on another network will be similar to using Story Network if the user has a signer that controls an AA wallet on Story Network. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_c2f1de82a51e902c97e21475779c4870` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1` `source_timestamp=2024-11-21T14:30:00Z`
- If users have an existing NFT on another network and want to register it on Story Network, a centralized service is required to validate NFT ownership. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_c2f1de82a51e902c97e21475779c4870` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1` `source_timestamp=2024-11-21T14:30:00Z`
- Supported user actions include creating IPA with/without NFT, creating policy, adding policy to IPA, minting license, creating derivative IP with license, paying/claiming royalty, and raising dispute. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_c2f1de82a51e902c97e21475779c4870` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1` `source_timestamp=2024-11-21T14:30:00Z`
- Target chains include EVM chains (shared addresses) and Solana; target app types include NFT projects, Web3 Social (Lens), and DeFi. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_c2f1de82a51e902c97e21475779c4870` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1` `source_timestamp=2024-11-21T14:30:00Z`
- There are three user flow types: web2 user (no wallet, no NFT), cross-chain web3 user registering IPA with no NFT, and cross-chain web3 user registering IPA with existing NFT. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_c2f1de82a51e902c97e21475779c4870` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1` `source_timestamp=2024-11-21T14:30:00Z`
- For web2 users, the flow involves creating a signer on Story Network, creating an AA wallet controlled by that signer, registering IPA (NFT minted to AA wallet), and managing license/royalty tokens via the AA wallet. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_c2f1de82a51e902c97e21475779c4870` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-1` `source_timestamp=2024-11-21T14:30:00Z`
- A cross-chain API can orchestrate royalty payments to a user's native chain (e.g., Solana) by verifying ownership, claiming USDC on Story L1, bridging funds, and sending to the user's wallet. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-2) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_62cef42949b7f8cbb47059471c768d47` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-2` `source_timestamp=2024-11-21T14:30:00Z`
- A risk exists where a user could call the API to transfer all tokens from an IP Account to their wallet; mitigation includes adding a delay at the API layer to wait for next block finalization. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-2) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_62cef42949b7f8cbb47059471c768d47` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-2` `source_timestamp=2024-11-21T14:30:00Z`
- Legal risk is probably acceptable based on conversation with Ben, but may require adding language to the user agreement. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-2) `source_document_id=srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54` `source_revision_id=srcrev_cf079f6965434a6e7a9a8bc442ab5133` `chunk_id=srcchunk_62cef42949b7f8cbb47059471c768d47` `native_locator=https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae#chunk-2` `source_timestamp=2024-11-21T14:30:00Z`

## Open Questions

- How will the centralized NFT ownership validation service be implemented securely?
- What are the specific smart contract optimizations for combining cross-chain steps?
- Which additional chains beyond EVM and Solana should be supported?

## Related Pages

- `web2-api-strategy`
- `xchain-api-v0-design`

## Sources

- `source_document_id`: `srcdoc_0a84b4fa3bbd45cb2c277cd681e87a54`
- `source_revision_id`: `srcrev_cf079f6965434a6e7a9a8bc442ab5133`
- `source_url`: [Notion source](https://www.notion.so/Web3-cross-chain-Solution-a428cedca1b142f393201038307c11ae)
