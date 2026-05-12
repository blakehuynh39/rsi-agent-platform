---
title: "Story Protocol Core Architecture"
type: "system"
slug: "systems/story-protocol-core-architecture"
freshness: "2023-06-29T15:13:00Z"
tags:
  - "architecture"
  - "franchise"
  - "nft"
  - "protocol"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_4be2d88855699ba4d5b6cf1e544a5d4c"
conflict_state: "none"
---

# Story Protocol Core Architecture

## Summary

Core architectural components of the Story Protocol, including StoryBlocks, Franchises, Ownership vs Control, and data storage mechanisms.

## Claims

- StoryBlocks are SP-specific assets like Characters, Stories, Locations, and Groups, represented as ERC-721 tokens with a sequential ID within a range specific for each story block. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- Ownership is defined as ERC-721 compliant ownership, the most native on-chain ownership, where the token is held in the wallet. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- Control means having the capability to write the SP Asset and control over the IP it represents. If an asset has a different owner and controller, it cannot be transferred. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- External NFTs (PFPs like BAYC, Azuki) are any third-party ERC-721 not native to SP. They cannot have more than 1 SP Character representing them per franchise. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- The SPFranchiseRegistry is a contract that mints NFTs with sequential IDs representing control and ownership over a Franchise. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- A Franchise is a collection of works and story blocks represented by an entry in the SPFranchiseRegistry. White Fountain is SP's first franchise. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- Canon consists of Story Blocks under the control of the Franchise Owner. Non-Canon or Lore consists of Story Blocks of a Franchise not controlled by the Franchise Owner. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- On-chain IP relevant data for each Story Block is stored in its NFT storage. Off-chain storage uses Arweave. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- Data Access Modules (DAMs) are contracts with rules to read and write to storage, specific for each story block. Only DAMs can write to the StoryBlockStorage, and only owners of each asset can write. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- Data Access Modules are not extensible by developers or franchise owners at launch, as they are part of the StoryBlockRegistry. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`
- Soft staking may grant write rights to a role other than the ERC-721 Owner while restricting transfers for the ERC-721 Owner, possibly with License Module implications. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-2) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_607f4eda78ed302d749853ba8cf222b4` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-2` `source_timestamp=2023-06-29T15:13:00Z`
- ERC-6551 is being considered for token bonding to an ERC-721 Franchise or PFP. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-2) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_607f4eda78ed302d749853ba8cf222b4` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-2` `source_timestamp=2023-06-29T15:13:00Z`
- When an external PFP is transferred and changes ownership, the new owner can control the previously controlled assets. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-2) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_607f4eda78ed302d749853ba8cf222b4` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-2` `source_timestamp=2023-06-29T15:13:00Z`
- Contributors or credits will be tracked on the protocol via a separate field on StoryBlocks called contributors, which is a list of wallet addresses. Population methods are not assumed. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1) `source_document_id=srcdoc_0305d251cdf50e4f9a2f426185ce182a` `source_revision_id=srcrev_4be2d88855699ba4d5b6cf1e544a5d4c` `chunk_id=srcchunk_2c70c91e8a54ac39525a63f98384845c` `native_locator=https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7#chunk-1` `source_timestamp=2023-06-29T15:13:00Z`

## Open Questions

- How will bridgeable assets and franchises work across L1 and L2 deployments?
- Should canon assets be protected so people cannot send random assets to the Franchise Owner?
- What goes on L1 vs L2?
- What is the role of originalOwner and how does it relate to IP requirements?
- Will the protocol be deployed on every chain, and can PFPs on L1 control assets on L2?

## Sources

- `source_document_id`: `srcdoc_0305d251cdf50e4f9a2f426185ce182a`
- `source_revision_id`: `srcrev_4be2d88855699ba4d5b6cf1e544a5d4c`
- `source_url`: [Notion source](https://www.notion.so/Live-Protocol-Updates-3490b130294a481494992e59b84373a7)
