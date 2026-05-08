---
title: "License Registry and License NFT Contract Relationship"
type: "decision"
slug: "decisions/license-registry-nft-contract-relationship"
freshness: "2024-03-22T08:21:00Z"
tags:
  - "architecture"
  - "design"
  - "license-registry"
  - "nft"
owners: []
source_revision_ids:
  - "srcrev_6bd94b3a6e414b491bd59d810e614171"
conflict_state: "none"
---

# License Registry and License NFT Contract Relationship

## Summary

Explores two design directions for the relationship between the License Registry and License NFT contracts: D.1 (tight coupling, all licenses on one address) and D.2 (decoupling, each policy framework has its own NFT contract).

## Claims

- Currently, the License Registry is ERC-1155 and represents ALL License Tokens from ALL policy frameworks. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99) `source_document_id=srcdoc_4e6c8fe14ff2281fb1db95efc482e91f` `source_revision_id=srcrev_6bd94b3a6e414b491bd59d810e614171` `chunk_id=srcchunk_55fc7667d5c5c807f7307899d4b4894c` `native_locator=https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99` `source_timestamp=2024-03-22T08:21:00Z`
- Leo has suggested that each license type should belong in its own NFT contract. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99) `source_document_id=srcdoc_4e6c8fe14ff2281fb1db95efc482e91f` `source_revision_id=srcrev_6bd94b3a6e414b491bd59d810e614171` `chunk_id=srcchunk_55fc7667d5c5c807f7307899d4b4894c` `native_locator=https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99` `source_timestamp=2024-03-22T08:21:00Z`
- Design direction D.1 maintains tight coupling: all License NFTs are minted on a single address (0x123), regardless of policy framework. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99) `source_document_id=srcdoc_4e6c8fe14ff2281fb1db95efc482e91f` `source_revision_id=srcrev_6bd94b3a6e414b491bd59d810e614171` `chunk_id=srcchunk_55fc7667d5c5c807f7307899d4b4894c` `native_locator=https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99` `source_timestamp=2024-03-22T08:21:00Z`
- Design direction D.2 decouples the License Registry from the License NFT contract, allowing each policy framework to mint License NFTs on its own address (e.g., Story’s PIL on 0x123, Lens on 0x456, Arweave on 0x789). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99) `source_document_id=srcdoc_4e6c8fe14ff2281fb1db95efc482e91f` `source_revision_id=srcrev_6bd94b3a6e414b491bd59d810e614171` `chunk_id=srcchunk_55fc7667d5c5c807f7307899d4b4894c` `native_locator=https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99` `source_timestamp=2024-03-22T08:21:00Z`
- A benefit of D.1 is a single source of truth (one address) for all types of License NFTs, simplifying license management for developers. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99) `source_document_id=srcdoc_4e6c8fe14ff2281fb1db95efc482e91f` `source_revision_id=srcrev_6bd94b3a6e414b491bd59d810e614171` `chunk_id=srcchunk_55fc7667d5c5c807f7307899d4b4894c` `native_locator=https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99` `source_timestamp=2024-03-22T08:21:00Z`
- A downside of D.1 is that all license types are under one contract address, which might be bad for business branding. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99) `source_document_id=srcdoc_4e6c8fe14ff2281fb1db95efc482e91f` `source_revision_id=srcrev_6bd94b3a6e414b491bd59d810e614171` `chunk_id=srcchunk_55fc7667d5c5c807f7307899d4b4894c` `native_locator=https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99` `source_timestamp=2024-03-22T08:21:00Z`
- A benefit of D.2 is that it allows companies to create their own licensing terms and own it as a brand with a unique address. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99) `source_document_id=srcdoc_4e6c8fe14ff2281fb1db95efc482e91f` `source_revision_id=srcrev_6bd94b3a6e414b491bd59d810e614171` `chunk_id=srcchunk_55fc7667d5c5c807f7307899d4b4894c` `native_locator=https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99` `source_timestamp=2024-03-22T08:21:00Z`
- A downside of D.2 is that developers have to manage ERC-1155 themselves, and there is no global license ID unless using a composite key (NFT contract address + license ID). `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99) `source_document_id=srcdoc_4e6c8fe14ff2281fb1db95efc482e91f` `source_revision_id=srcrev_6bd94b3a6e414b491bd59d810e614171` `chunk_id=srcchunk_55fc7667d5c5c807f7307899d4b4894c` `native_locator=https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99` `source_timestamp=2024-03-22T08:21:00Z`

## Open Questions

- If we remove ERC-1155 from the License Registry, can the License Registry and Module be combined?

## Sources

- `source_document_id`: `srcdoc_4e6c8fe14ff2281fb1db95efc482e91f`
- `source_revision_id`: `srcrev_6bd94b3a6e414b491bd59d810e614171`
- `source_url`: [Notion source](https://www.notion.so/8-Relationship-between-License-Registry-and-License-NFT-contracts-65d32e7545d34c5a9f7294fe8ef96a99)
