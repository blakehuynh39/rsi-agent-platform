---
title: "PFP Engineering Specs"
type: "project"
slug: "projects/pfp-engineering-specs"
freshness: "2026-05-05T06:27:23Z"
tags:
  - "ip-asset"
  - "licensing"
  - "nft"
  - "pfp"
  - "protocol-integration"
owners: []
source_revision_ids:
  - "srcrev_c0c7faa312cf2545a0ee63cdc4db60c8"
conflict_state: "none"
---

# PFP Engineering Specs

## Summary

Specification for the PFP project, a mechanism to bring contributors into the Emergence ecosystem via a legal contract and protocol integrations like licensing, IP assets, collect, royalties, and relationships.

## Claims

- The primary purpose of the PFPs is to have a scalable mechanism of bringing contributors into the Emergence ecosystem / our protocol. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649) `source_document_id=srcdoc_1abee670b77650df72ec3f1df937ca83` `source_revision_id=srcrev_c0c7faa312cf2545a0ee63cdc4db60c8` `chunk_id=srcchunk_caaeea4610a8fde58b95cb31bf7f3216` `native_locator=https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649` `source_timestamp=2026-05-05T06:27:23Z`
- The PFP holds a legal contract that gives the holder certain rights to contribute to the franchise and to benefit from those contributions. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649) `source_document_id=srcdoc_1abee670b77650df72ec3f1df937ca83` `source_revision_id=srcrev_c0c7faa312cf2545a0ee63cdc4db60c8` `chunk_id=srcchunk_caaeea4610a8fde58b95cb31bf7f3216` `native_locator=https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649` `source_timestamp=2026-05-05T06:27:23Z`
- Goals include using the licensing module to grant prophets their rights, providing a web3-native welcome, demonstrating protocol capabilities (licensing, IP Asset, relationships, royalties, collect), creating supplementary artwork, and being a reference for future projects. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649) `source_document_id=srcdoc_1abee670b77650df72ec3f1df937ca83` `source_revision_id=srcrev_c0c7faa312cf2545a0ee63cdc4db60c8` `chunk_id=srcchunk_caaeea4610a8fde58b95cb31bf7f3216` `native_locator=https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649` `source_timestamp=2026-05-05T06:27:23Z`
- Smart contract requirements: PFPs should not be tradable (or conditionally tradable), ERC721 on Ethereum Mainnet, registered as an IP Asset with a Prophet program license, collectible with tipping option, airdroppable, and metadata should display protocol stats (feasibility TBD). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649) `source_document_id=srcdoc_1abee670b77650df72ec3f1df937ca83` `source_revision_id=srcrev_c0c7faa312cf2545a0ee63cdc4db60c8` `chunk_id=srcchunk_caaeea4610a8fde58b95cb31bf7f3216` `native_locator=https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649` `source_timestamp=2026-05-05T06:27:23Z`
- Protocol integrations sections (Licensing, IP Asset, Collect, Royalties, Relationships) are detailed in the document but no specific implementation steps are listed. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649) `source_document_id=srcdoc_1abee670b77650df72ec3f1df937ca83` `source_revision_id=srcrev_c0c7faa312cf2545a0ee63cdc4db60c8` `chunk_id=srcchunk_caaeea4610a8fde58b95cb31bf7f3216` `native_locator=https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649` `source_timestamp=2026-05-05T06:27:23Z`
- Tasks identified: Create and register PFPs as IPAs, implement/integrate licenses, royalties, collect parameters, relationships, create NFT metadata, upload images, end-to-end airdrop/minting testing. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649) `source_document_id=srcdoc_1abee670b77650df72ec3f1df937ca83` `source_revision_id=srcrev_c0c7faa312cf2545a0ee63cdc4db60c8` `chunk_id=srcchunk_caaeea4610a8fde58b95cb31bf7f3216` `native_locator=https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649` `source_timestamp=2026-05-05T06:27:23Z`

## Open Questions

- Are there features we want/need that won’t be ready from the protocol team?
- Minting methods: 1. External contract and use relationships to tie the collection to the protocol? 2. Create a contract that calls the protocol’s create IP function? 3. Create the IPAs by calling the contract directly?
- What hooks do we need? 1. Token-gated hook 2. Payment hook

## Sources

- `source_document_id`: `srcdoc_1abee670b77650df72ec3f1df937ca83`
- `source_revision_id`: `srcrev_c0c7faa312cf2545a0ee63cdc4db60c8`
- `source_url`: [Notion source](https://www.notion.so/WIP-PFP-Engineering-Specs-20afc72c60c242e49a4b8eb6c6bec649)
