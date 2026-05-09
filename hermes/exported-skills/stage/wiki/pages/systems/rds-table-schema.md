---
title: "RDS Table Schema (March 2023)"
type: "system"
slug: "systems/rds-table-schema"
freshness: "2023-04-03T05:25:00Z"
tags:
  - "database"
  - "nft"
  - "rds"
  - "schema"
  - "story"
  - "wallet"
owners: []
source_revision_ids:
  - "srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd"
conflict_state: "none"
---

# RDS Table Schema (March 2023)

## Summary

Overview of the RDS table schema as of March 2023, covering story-related, NFT-related, and wallet-related tables, along with relationship tables.

## Claims

- The RDS schema is organized into three main groups: Story related, NFT related, and Wallet related tables. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1) `source_document_id=srcdoc_3414a33a226ac7a9512a93b1dfa6756e` `source_revision_id=srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd` `chunk_id=srcchunk_fac389c5908dc0aef7ab09a2d11cfed2` `native_locator=https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1` `source_timestamp=2023-04-03T05:25:00Z`
- Relationship tables exist to capture connections between entities, such as franchise_collection. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1) `source_document_id=srcdoc_3414a33a226ac7a9512a93b1dfa6756e` `source_revision_id=srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd` `chunk_id=srcchunk_fac389c5908dc0aef7ab09a2d11cfed2` `native_locator=https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1` `source_timestamp=2023-04-03T05:25:00Z`
- Story entities consist of franchise, story, and chapter, with a hierarchy: a franchise can include multiple stories, and a story can include multiple chapters. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1) `source_document_id=srcdoc_3414a33a226ac7a9512a93b1dfa6756e` `source_revision_id=srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd` `chunk_id=srcchunk_fac389c5908dc0aef7ab09a2d11cfed2` `native_locator=https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1` `source_timestamp=2023-04-03T05:25:00Z`
- Story content and images are stored in S3 and preloaded in cache, not in the database. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1) `source_document_id=srcdoc_3414a33a226ac7a9512a93b1dfa6756e` `source_revision_id=srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd` `chunk_id=srcchunk_fac389c5908dc0aef7ab09a2d11cfed2` `native_locator=https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1` `source_timestamp=2023-04-03T05:25:00Z`
- The story_franchise table stores franchise-level information. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1) `source_document_id=srcdoc_3414a33a226ac7a9512a93b1dfa6756e` `source_revision_id=srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd` `chunk_id=srcchunk_fac389c5908dc0aef7ab09a2d11cfed2` `native_locator=https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-1` `source_timestamp=2023-04-03T05:25:00Z`
- NFT-related tables include nft_collection (collection-level info), nft_token (token-level info), and nft_allowlist (tracks allowlists for collections). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-2) `source_document_id=srcdoc_3414a33a226ac7a9512a93b1dfa6756e` `source_revision_id=srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd` `chunk_id=srcchunk_747e44d0a5b47bf0d9c9dad139551bc5` `native_locator=https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-2` `source_timestamp=2023-04-03T05:25:00Z`
- The wallet_merkle_proof table stores wallet proofs for allowlists; future plans may include a wallet profile table for identities like emails and ENS. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-3) `source_document_id=srcdoc_3414a33a226ac7a9512a93b1dfa6756e` `source_revision_id=srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd` `chunk_id=srcchunk_74f7b012f6536ca850d12ff91644c117` `native_locator=https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7#chunk-3` `source_timestamp=2023-04-03T05:25:00Z`

## Sources

- `source_document_id`: `srcdoc_3414a33a226ac7a9512a93b1dfa6756e`
- `source_revision_id`: `srcrev_5b9f5ab07d96aeec4acc83f9cfbcd7dd`
- `source_url`: [Notion source](https://www.notion.so/RDS-table-Schema-March-2023-e546c719255948eb8ab0748c895d93c7)
