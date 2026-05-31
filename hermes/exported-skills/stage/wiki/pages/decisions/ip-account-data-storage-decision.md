---
title: "IPAccount Data Storage Decision"
type: "decision"
slug: "decisions/ip-account-data-storage-decision"
freshness: "2024-03-15T22:42:00Z"
tags:
  - "architecture"
  - "data-storage"
  - "ip-account"
owners: []
source_revision_ids:
  - "srcrev_e81d198c591d4bd25a6c9aed004e8af7"
conflict_state: "none"
---

# IPAccount Data Storage Decision

## Summary

Decision finalized via voting bypass for PR #2 to adopt the Open Data Access system over the Module-based Data System for IPAccount storage. IPAccount will hold all relevant data in a generic bytes storage, with modules writing to assigned namespaces.

## Claims

- The Open Data Access system (Option D.2) was finalized and implemented via PR #2, bypassing a vote. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d) `source_document_id=srcdoc_2c2340b8ebaab7669798ad3181340fcd` `source_revision_id=srcrev_e81d198c591d4bd25a6c9aed004e8af7` `chunk_id=srcchunk_269afe4336da4c28977a8af3d49372b5` `native_locator=https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d` `source_timestamp=2024-03-15T22:42:00Z`
- Previously, IPAccount data was dispersed across modules; the Licensing Module held policy data, and the Attribution Module potentially held attribution data. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d) `source_document_id=srcdoc_2c2340b8ebaab7669798ad3181340fcd` `source_revision_id=srcrev_e81d198c591d4bd25a6c9aed004e8af7` `chunk_id=srcchunk_269afe4336da4c28977a8af3d49372b5` `native_locator=https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d` `source_timestamp=2024-03-15T22:42:00Z`
- Under the Open Data Access system, modules must write their data to assigned namespaces in IPAccount’s generic bytes storage. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d) `source_document_id=srcdoc_2c2340b8ebaab7669798ad3181340fcd` `source_revision_id=srcrev_e81d198c591d4bd25a6c9aed004e8af7` `chunk_id=srcchunk_269afe4336da4c28977a8af3d49372b5` `native_locator=https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d` `source_timestamp=2024-03-15T22:42:00Z`
- The Open Data Access system incurs more gas due to external calls for modules to read/write on IPAccount storage. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d) `source_document_id=srcdoc_2c2340b8ebaab7669798ad3181340fcd` `source_revision_id=srcrev_e81d198c591d4bd25a6c9aed004e8af7` `chunk_id=srcchunk_269afe4336da4c28977a8af3d49372b5` `native_locator=https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d` `source_timestamp=2024-03-15T22:42:00Z`
- The Module-based Data System benefits from gas efficiency and easier auditability because logic and storage reside in the same contract. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d) `source_document_id=srcdoc_2c2340b8ebaab7669798ad3181340fcd` `source_revision_id=srcrev_e81d198c591d4bd25a6c9aed004e8af7` `chunk_id=srcchunk_269afe4336da4c28977a8af3d49372b5` `native_locator=https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d` `source_timestamp=2024-03-15T22:42:00Z`

## Open Questions

- How to distinguish between private data that should stay in a module versus data written to IPAccount storage?
- How to manage namespace collisions when modules register globally unique namespaces?

## Sources

- `source_document_id`: `srcdoc_2c2340b8ebaab7669798ad3181340fcd`
- `source_revision_id`: `srcrev_e81d198c591d4bd25a6c9aed004e8af7`
- `source_url`: [Notion source](https://www.notion.so/1-IPAccount-Open-Data-Access-or-Module-based-Data-System-7144211a4f1e46e9a76ee758b793ab6d)
