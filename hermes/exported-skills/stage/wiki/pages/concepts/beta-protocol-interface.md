---
title: "Beta Protocol Interface"
type: "concept"
slug: "concepts/beta-protocol-interface"
freshness: "2024-12-03T05:20:00Z"
tags:
  - "beta"
  - "interface"
  - "protocol"
owners: []
source_revision_ids:
  - "srcrev_2c61660abca955e5170ea57b4dab00eb"
conflict_state: "none"
---

# Beta Protocol Interface

## Summary

Design decisions and interface specifications for the Beta Protocol, including core components (IP Account, Registries) and modules (Licensing, Tagging, Dispute, Royalty).

## Claims

- The Beta Protocol Interface page documents design decisions and components. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_8536770001d813e5b97ad7923856200c` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- Core components include IPAccount, IPAccountRegistry, IPRecordRegistry, and LicenseRegistry. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_8536770001d813e5b97ad7923856200c` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- Modules include Module Registry, Access Control Manager, Registration Module, Licensing Module, Tagging Module, Dispute Module, and Royalty module. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_8536770001d813e5b97ad7923856200c` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- The IPAccount interface defines isValidSigner and execute functions. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_8536770001d813e5b97ad7923856200c` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- The IPAccountRegistry interface defines registerIpAccount, ipAccount, and isRegistered functions. `claim:claim_2_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_8536770001d813e5b97ad7923856200c` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- The AccessController interface defines setPolicy, getPolicy, and checkPolicy functions. `claim:claim_2_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_8536770001d813e5b97ad7923856200c` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- The ModuleRegistry interface defines registerModule and getModule functions. `claim:claim_2_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_8536770001d813e5b97ad7923856200c` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- The Licensing Module includes a LicenseRegistry with mintLicense and burnLicense functions, and a License struct containing licensePolicy and licensorIpIds. `claim:claim_2_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-2) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_f58f45a387f40f44d18337d636328bc3` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-2` `source_timestamp=2024-12-03T05:20:00Z`
- The page mentions a Tagging Module. `claim:claim_2_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-2) `source_document_id=srcdoc_1cc17f76c970edc6d1cdec37f0398209` `source_revision_id=srcrev_2c61660abca955e5170ea57b4dab00eb` `chunk_id=srcchunk_f58f45a387f40f44d18337d636328bc3` `native_locator=https://www.notion.so/Beta-Protocol-Interface-9b9088c760334da5bd30828ddd13d6ae#chunk-2` `source_timestamp=2024-12-03T05:20:00Z`

## Related Pages

- `concepts/protocol-home`

## Sources

- `source_document_id`: `srcdoc_aeb45af31362098b908ed27d9e2ce76a`
- `source_revision_id`: `srcrev_f8ab25602df939e8c386929f90fc3476`
- `source_url`: [Notion source](https://www.notion.so/Registration-Module-76974164cd8447bb92cb55c67e70f96e)
