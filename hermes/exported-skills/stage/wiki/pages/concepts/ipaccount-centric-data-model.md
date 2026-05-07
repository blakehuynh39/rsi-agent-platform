---
title: "IPAccount Centric Data Model"
type: "concept"
slug: "concepts/ipaccount-centric-data-model"
freshness: "2024-12-05T16:42:00Z"
tags:
  - "data-model"
  - "IPAccount"
  - "modules"
  - "namespaced-storage"
owners: []
source_revision_ids:
  - "srcrev_88ca2a7e5818d5cf5fcb4b94efead38c"
conflict_state: "none"
---

# IPAccount Centric Data Model

## Summary

The IPAccount Centric Data Model centralizes all IP-related data within the IPAccount structure, enabling efficient data handling by Modules. Modules read from and write back to the IPAccount, with data isolated by unique namespaces to prevent conflicts. Additional data like NFT tokens is stored in a Registry.

## Claims

- The IPAccount Centric Data Model centralizes all IP-related data within the IPAccount structure. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_31cbbb336c4a19f4855a0dacb8c89f59` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- Modules retrieve necessary data from the IPAccount and subsequently write back any outputs or data changes into it. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_31cbbb336c4a19f4855a0dacb8c89f59` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- To prevent conflicts during data write operations between different Modules, each Module's data within the IPAccount is isolated by a unique namespace. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_31cbbb336c4a19f4855a0dacb8c89f59` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- IP data is stored in two primary locations: the IPAccount (main repository) and the Registry (for data that cannot be efficiently stored within the IPAccount, such as NFT Tokens). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_31cbbb336c4a19f4855a0dacb8c89f59` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- The IPAccount uses a namespaced storage pattern where a namespace is generated using the Module's address. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_31cbbb336c4a19f4855a0dacb8c89f59` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- Every Module can read data from any namespace, but only the owning Module (whose address matches the namespace) can write data into its respective namespace. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_31cbbb336c4a19f4855a0dacb8c89f59` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-1` `source_timestamp=2024-12-05T16:42:00Z`
- Modules implement version control for compatibility checking; View Modules use version information to determine if they can accurately interpret and display data from a specific Module version. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-2) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_42a922e19622199c55009229b9b35797` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-2` `source_timestamp=2024-12-05T16:42:00Z`
- All View Modules implement a function `isSupported(IPAccount ipAccount)` that checks whether the View Module can support a given IPAccount by examining the module version used. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-2) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_42a922e19622199c55009229b9b35797` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-2` `source_timestamp=2024-12-05T16:42:00Z`
- In the License Creation for a Book workflow, metadata is written into the IPAccount, the license module creates a license and writes policy information into the IPAccount, and the BookLicenseView Module provides a tokenURI() method that aggregates and displays book license information from both the IPAccount and the License Registry. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-2) `source_document_id=srcdoc_8b58c15aea50597fe2a57a4c754ea8e1` `source_revision_id=srcrev_88ca2a7e5818d5cf5fcb4b94efead38c` `chunk_id=srcchunk_42a922e19622199c55009229b9b35797` `native_locator=https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc#chunk-2` `source_timestamp=2024-12-05T16:42:00Z`

## Sources

- `source_document_id`: `srcdoc_8b58c15aea50597fe2a57a4c754ea8e1`
- `source_revision_id`: `srcrev_88ca2a7e5818d5cf5fcb4b94efead38c`
- `source_url`: [Notion source](https://www.notion.so/IPAccount-Centric-Data-Model-c553225fe8ad477286fce289407a92fc)
