---
title: "IP Account Centric Model"
type: "concept"
slug: "concepts/ip-account-centric-model"
freshness: "2024-12-30T07:49:00Z"
tags:
  - "AccessController"
  - "Architecture"
  - "IPAccount"
  - "Modules"
  - "Orchestrator"
owners: []
source_revision_ids:
  - "srcrev_41b1b750d39ffc0c9a336b3a38bce7e5"
conflict_state: "none"
---

# IP Account Centric Model

## Summary

The IP Account Centric Model conceptualizes IPAccount as the core identity (noun) and Modules as actions (verbs). IP Accounts represent IP as real entities on Ethereum, with IP NFTs serving as their identity. The model emphasizes composability, future-proofing through modules, and fine-grained access control via an Orchestrator and AccessController.

## Claims

- IPAccount serves as the core identity for all actions in the system. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- An IP NFT represents the identity of an IP, analogous to a passport. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- IP Account represents the IP as a real entity. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- Ownership of an IP NFT implies ownership of the corresponding IP Account. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- IPAccount enables IP to be a first-class citizen on Ethereum. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- IPAccount can hold NFTs representing licenses and royalties. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- IPAccount facilitates integration with other protocols and contracts. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- IPAccount can create licenses, set up royalty distributions, create relationships, and collect assets. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- IPAccount stores state information, including links and licenses. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- IPAccount acts as the caller to initiate actions and invoke modules. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- IPAccount can leverage new modules to extend capabilities, ensuring future-proofing. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- An Orchestrator provides authorization between IPAccount and Modules. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`
- Modules are stateless and can be dynamically enabled or disabled by the Orchestrator. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-2) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_b8bdc689c9d643f41efd25cfe1e10a11` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-2` `source_timestamp=2024-12-30T07:49:00Z`
- The AccessController provides fine-grained access control per caller at the function level. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-2) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_b8bdc689c9d643f41efd25cfe1e10a11` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-2` `source_timestamp=2024-12-30T07:49:00Z`
- The design principle is simple with composability, where the protocol defines only basic opcodes and all actions are considered modules. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1) `source_document_id=srcdoc_41c373a2be12df4e63581fc6d62ab2af` `source_revision_id=srcrev_41b1b750d39ffc0c9a336b3a38bce7e5` `chunk_id=srcchunk_c587580bbad1ca80b92e66f302ffcd05` `native_locator=https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288#chunk-1` `source_timestamp=2024-12-30T07:49:00Z`

## Sources

- `source_document_id`: `srcdoc_41c373a2be12df4e63581fc6d62ab2af`
- `source_revision_id`: `srcrev_41b1b750d39ffc0c9a336b3a38bce7e5`
- `source_url`: [Notion source](https://www.notion.so/IP-Account-Centric-Model-924c2777de614fa58e60bcae1faae288)
