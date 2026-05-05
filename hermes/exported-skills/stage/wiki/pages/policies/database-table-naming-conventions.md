---
title: "Database Table Naming Conventions"
type: "policy"
slug: "policies/database-table-naming-conventions"
freshness: "2026-05-05T05:40:56Z"
tags:
  - "database"
  - "naming-conventions"
  - "prefixes"
owners: []
source_revision_ids:
  - "srcrev_b15d3133f2efee003ce55dff172e8e31"
conflict_state: "none"
---

# Database Table Naming Conventions

## Summary

Policy for database table naming using prefixes to indicate responsibility: wallet_, nft_, story_.

## Claims

- Database table names should use prefixes as namespaces to separate and clarify the responsibility of different tables. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1) `source_document_id=srcdoc_9f4095914c32736f8a8a2bf6266d2c02` `source_revision_id=srcrev_b15d3133f2efee003ce55dff172e8e31` `chunk_id=srcchunk_c63a883711e8973ad1a527800728183a` `native_locator=https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1` `source_timestamp=2026-05-05T05:40:56Z`
- Tables associated with wallet data should use the prefix 'wallet_', e.g., wallet_merkle_proof. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1) `source_document_id=srcdoc_9f4095914c32736f8a8a2bf6266d2c02` `source_revision_id=srcrev_b15d3133f2efee003ce55dff172e8e31` `chunk_id=srcchunk_c63a883711e8973ad1a527800728183a` `native_locator=https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1` `source_timestamp=2026-05-05T05:40:56Z`
- Tables related to NFTs should use the prefix 'nft_', e.g., nft_collection, nft_allowlist, nft_token. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1) `source_document_id=srcdoc_9f4095914c32736f8a8a2bf6266d2c02` `source_revision_id=srcrev_b15d3133f2efee003ce55dff172e8e31` `chunk_id=srcchunk_c63a883711e8973ad1a527800728183a` `native_locator=https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1` `source_timestamp=2026-05-05T05:40:56Z`
- Tables related to content (stories) should use the prefix 'story_', e.g., story_franchise, story_info, story_chapter. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1) `source_document_id=srcdoc_9f4095914c32736f8a8a2bf6266d2c02` `source_revision_id=srcrev_b15d3133f2efee003ce55dff172e8e31` `chunk_id=srcchunk_c63a883711e8973ad1a527800728183a` `native_locator=https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1` `source_timestamp=2026-05-05T05:40:56Z`

## Sources

- `source_document_id`: `srcdoc_9f4095914c32736f8a8a2bf6266d2c02`
- `source_revision_id`: `srcrev_b15d3133f2efee003ce55dff172e8e31`
- `source_url`: [Notion source](https://www.notion.so/DB-Naming-schema-638f8f681dd140c493b15a7764ffa0c1)
