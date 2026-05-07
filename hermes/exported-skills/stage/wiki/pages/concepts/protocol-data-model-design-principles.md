---
title: "Protocol Data Model Design Principles"
type: "concept"
slug: "concepts/protocol-data-model-design-principles"
freshness: "2023-05-17T18:35:00Z"
tags:
  - "data-model"
  - "design-principles"
  - "on-chain"
owners: []
source_revision_ids:
  - "srcrev_5b8694592b19a0759c2c700a8d15b864"
conflict_state: "none"
---

# Protocol Data Model Design Principles

## Summary

Design principles for the protocol data model, focusing on on-chain data properties and criteria for storing data on-chain.

## Claims

- On-chain data is immutable and traceable: once written to a block, the data is immutable in that block, and history is preserved at every block. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2) `source_document_id=srcdoc_c385c17549da5e7ade34384340948fcd` `source_revision_id=srcrev_5b8694592b19a0759c2c700a8d15b864` `chunk_id=srcchunk_2a9a09cf4ed1b2688b04cd37c5a4d0d5` `native_locator=https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2` `source_timestamp=2023-05-17T18:35:00Z`
- On-chain data is expensive to write (costs gas) but free to read. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2) `source_document_id=srcdoc_c385c17549da5e7ade34384340948fcd` `source_revision_id=srcrev_5b8694592b19a0759c2c700a8d15b864` `chunk_id=srcchunk_2a9a09cf4ed1b2688b04cd37c5a4d0d5` `native_locator=https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2` `source_timestamp=2023-05-17T18:35:00Z`
- On-chain data is open and permissionless; anyone can access the entire blockchain data. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2) `source_document_id=srcdoc_c385c17549da5e7ade34384340948fcd` `source_revision_id=srcrev_5b8694592b19a0759c2c700a8d15b864` `chunk_id=srcchunk_2a9a09cf4ed1b2688b04cd37c5a4d0d5` `native_locator=https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2` `source_timestamp=2023-05-17T18:35:00Z`
- The data model should save valuable, nontrivial, and non-sensitive data on-chain. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2) `source_document_id=srcdoc_c385c17549da5e7ade34384340948fcd` `source_revision_id=srcrev_5b8694592b19a0759c2c700a8d15b864` `chunk_id=srcchunk_2a9a09cf4ed1b2688b04cd37c5a4d0d5` `native_locator=https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2` `source_timestamp=2023-05-17T18:35:00Z`
- Valuable data adds value to the protocol and must be concise without much duplication. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2) `source_document_id=srcdoc_c385c17549da5e7ade34384340948fcd` `source_revision_id=srcrev_5b8694592b19a0759c2c700a8d15b864` `chunk_id=srcchunk_2a9a09cf4ed1b2688b04cd37c5a4d0d5` `native_locator=https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2` `source_timestamp=2023-05-17T18:35:00Z`
- Nontrivial data won't be changed too frequently and easily. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2) `source_document_id=srcdoc_c385c17549da5e7ade34384340948fcd` `source_revision_id=srcrev_5b8694592b19a0759c2c700a8d15b864` `chunk_id=srcchunk_2a9a09cf4ed1b2688b04cd37c5a4d0d5` `native_locator=https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2` `source_timestamp=2023-05-17T18:35:00Z`
- Non-sensitive data shouldn't be stored on-chain, but zero-knowledge technology can offer verifiable privacy protection on-chain. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2) `source_document_id=srcdoc_c385c17549da5e7ade34384340948fcd` `source_revision_id=srcrev_5b8694592b19a0759c2c700a8d15b864` `chunk_id=srcchunk_2a9a09cf4ed1b2688b04cd37c5a4d0d5` `native_locator=https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2` `source_timestamp=2023-05-17T18:35:00Z`

## Sources

- `source_document_id`: `srcdoc_c385c17549da5e7ade34384340948fcd`
- `source_revision_id`: `srcrev_5b8694592b19a0759c2c700a8d15b864`
- `source_url`: [Notion source](https://www.notion.so/Protocol-Data-Model-and-Module-Discussion-f004711a228047e7abd2d9b48911b9b2)
