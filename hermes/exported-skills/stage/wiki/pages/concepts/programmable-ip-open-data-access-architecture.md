---
title: "Programmable IP Open Data Access Architecture"
type: "concept"
slug: "concepts/programmable-ip-open-data-access-architecture"
freshness: "2024-12-05T16:42:00Z"
tags:
  - "architecture"
  - "data-access"
  - "ip-account"
  - "programmable-ip"
owners: []
source_revision_ids:
  - "srcrev_0dbeb1e0e965ffe89272c598cd2b4cca"
conflict_state: "none"
---

# Programmable IP Open Data Access Architecture

## Summary

Defines the core principles and implementation structure for an open, equitable data access architecture in the Programmable IP ecosystem, centered around the IPAccount data model and namespaced storage.

## Claims

- Programmable IP represents a paradigm shift enabling users/developers to write their own programs to operate and manipulate IP data, hinging on data and functions. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_3dad1b381cbcfc2d63cdf6ac0adca87e` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- The success of Programmable IP rests on two key principles: Open Data Access (no super modules, all modules access data under the same rules) and Equal Access (every participant has equal opportunity, no privileged access). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_3dad1b381cbcfc2d63cdf6ac0adca87e` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- To achieve openness and equality, the architecture proposes Uniform Access Rules where any module can write its own data into any IPAccount and read any data from any IPAccount. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_3dad1b381cbcfc2d63cdf6ac0adca87e` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- Data Ownership Protocols dictate that only the owner of data can write/change it, while reading data is unrestricted, allowing anyone to read any data. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_3dad1b381cbcfc2d63cdf6ac0adca87e` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- The IPAccount acts as a structured storage, enforcing data ownership access rules, standardizing data access and sharing between modules, and identifying data ownership. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_3dad1b381cbcfc2d63cdf6ac0adca87e` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- Core metadata can be made immutable by restricting each IPAccount from writing metadata more than once via the CoreMetadataModule. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-2) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_462bd5e33686bbe1b7b3ffa0c94db876` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-2` `source_timestamp=2024-12-05T16:42:00Z`
- New user metadata can be added by creating a new UserMetadataModule, and a ViewModule can be created to read and display data from both Core Metadata and User Metadata. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-2) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_462bd5e33686bbe1b7b3ffa0c94db876` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-2` `source_timestamp=2024-12-05T16:42:00Z`
- Metadata upgrade/migration can be avoided by creating a MetadataModuleV2 for new metadata and a MetadataViewModuleV2 that reads from both previous and new metadata sources to display combined information. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-2) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_462bd5e33686bbe1b7b3ffa0c94db876` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-2` `source_timestamp=2024-12-05T16:42:00Z`
- A metadata module can be written to control which metadata is mutable and which is immutable. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-2) `source_document_id=srcdoc_539a5190b3d2399022d33187affb63e9` `source_revision_id=srcrev_0dbeb1e0e965ffe89272c598cd2b4cca` `chunk_id=srcchunk_462bd5e33686bbe1b7b3ffa0c94db876` `native_locator=https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610#chunk-2` `source_timestamp=2024-12-05T16:42:00Z`

## Sources

- `source_document_id`: `srcdoc_539a5190b3d2399022d33187affb63e9`
- `source_revision_id`: `srcrev_0dbeb1e0e965ffe89272c598cd2b4cca`
- `source_url`: [Notion source](https://www.notion.so/Open-Data-Access-Architecture-Embracing-Programmable-IP-10ba12c8c8514a87ae845d01a037a610)
