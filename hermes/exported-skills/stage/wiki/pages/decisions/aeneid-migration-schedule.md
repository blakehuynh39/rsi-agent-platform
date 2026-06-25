---
title: "Aeneid Migration Schedule Decision"
type: "decision"
slug: "decisions/aeneid-migration-schedule"
freshness: "2026-01-30T01:03:00Z"
tags:
  - "aeneid"
  - "migration"
  - "schedule"
  - "validators"
owners:
  - "U079ZJ48D62"
  - "U07A7AUGL5V"
  - "U07KLPN0JN6"
  - "U07TNT9N4JC"
  - "U082UKSD3BR"
  - "U0883L0RBRR"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_58a5055208377193fec7e9069176f109"
  - "srcrev_610ef8eff4a2acaa934d039cc88780ef"
  - "srcrev_a7455a72581ef09e8364b80a5b00178d"
  - "srcrev_c98f258f04ec72173914d2f41cca9dae"
  - "srcrev_f60559c7cd8d8e7fe373593b9bb57525"
  - "srcrev_f7c55148aacfece9f1c814514d37e881"
conflict_state: "none"
---

# Aeneid Migration Schedule Decision

## Summary

The migration of Aeneid RPC and validators was initially scheduled for tomorrow 11PM PT, but after a date confusion it was rescheduled to Monday 3PM BJT. A technical sync issue on validator4 was resolved by adding a peer manually.

## Claims

- The initial plan was to migrate Aeneid RPC and validators tomorrow 11PM PT without downtime. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_58a5055208377193fec7e9069176f109` `chunk_id=srcchunk_23eb033bcf78af39a99319d9ae61d766` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674617.908059` `source_timestamp=2026-01-29T11:17:29Z`
- Validator use1-aeneid-validator4 experienced block sync errors, with geth reporting missing finalized blocks. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f60559c7cd8d8e7fe373593b9bb57525` `chunk_id=srcchunk_d211a0c443638f4434efda509e900c54` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678206.437309` `source_timestamp=2026-01-29T09:16:46Z`
- The CL logs indicated that the execution engine was syncing and blocks were being processed while syncing. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f7c55148aacfece9f1c814514d37e881` `chunk_id=srcchunk_794a1f7289b9b9c443ac9fcff3645248` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678252.258289` `source_timestamp=2026-01-29T09:17:32Z`
- The sync issue was resolved by adding a peer manually, after which everything worked fine. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a7455a72581ef09e8364b80a5b00178d` `chunk_id=srcchunk_0c93db5007818e1eae81bb9ad758c83d` `native_locator=slack:C0547N89JUB:1769674617.908059:1769685706.820279` `source_timestamp=2026-01-29T11:23:20Z`
- Due to date confusion, the migration was rescheduled to Monday 3PM BJT. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_610ef8eff4a2acaa934d039cc88780ef` `chunk_id=srcchunk_148071300eebf431f70d10d4252da234` `native_locator=slack:C0547N89JUB:1769674617.908059:1769724531.593729` `source_timestamp=2026-01-29T22:08:51Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_c98f258f04ec72173914d2f41cca9dae` `chunk_id=srcchunk_8d612c9c5145cacd90bef4e7284eeafd` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734980.602749` `source_timestamp=2026-01-30T01:03:00Z`

## Sources

- `source_document_id`: `srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d`
- `source_revision_id`: `srcrev_161498e4abc3a93f291bac397c708d53`
