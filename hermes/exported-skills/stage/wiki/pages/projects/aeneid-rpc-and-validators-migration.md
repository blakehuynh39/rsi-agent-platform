---
title: "Aeneid RPC and Validators Migration"
type: "project"
slug: "projects/aeneid-rpc-and-validators-migration"
freshness: "2026-01-30T01:03:00Z"
tags:
  - "aeneid"
  - "migration"
  - "rpc"
  - "scheduling"
  - "validators"
owners:
  - "U079ZJ48D62"
  - "U07A7AUGL5V"
  - "U07KLPN0JN6"
  - "U07TNT9N4JC"
  - "U082UKSD3BR"
source_revision_ids:
  - "srcrev_58a5055208377193fec7e9069176f109"
  - "srcrev_7ac036b36f4fae468f0626cee4689743"
  - "srcrev_a7455a72581ef09e8364b80a5b00178d"
  - "srcrev_c98f258f04ec72173914d2f41cca9dae"
  - "srcrev_f60559c7cd8d8e7fe373593b9bb57525"
  - "srcrev_f7c55148aacfece9f1c814514d37e881"
conflict_state: "none"
---

# Aeneid RPC and Validators Migration

## Summary

Planning, scheduling changes, and execution issues for the Aeneid RPC and validators migration.

## Claims

- Initial migration was planned for 11 PM PT with no expected downtime. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_58a5055208377193fec7e9069176f109` `chunk_id=srcchunk_23eb033bcf78af39a99319d9ae61d766` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674617.908059` `source_timestamp=2026-01-29T11:17:29Z`
- Validator node use1-aeneid-validator4 encountered sync errors after resync, including 'Number of finalized block is missing' (EL) and 'execution engine is syncing' causing proposal rejection (CL). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f60559c7cd8d8e7fe373593b9bb57525` `chunk_id=srcchunk_d211a0c443638f4434efda509e900c54` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678206.437309` `source_timestamp=2026-01-29T09:16:46Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f7c55148aacfece9f1c814514d37e881` `chunk_id=srcchunk_794a1f7289b9b9c443ac9fcff3645248` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678252.258289` `source_timestamp=2026-01-29T09:17:32Z`
- The sync issue was caused by a missing peer; manually adding the peer resolved the problem and allowed migration to proceed. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a7455a72581ef09e8364b80a5b00178d` `chunk_id=srcchunk_0c93db5007818e1eae81bb9ad758c83d` `native_locator=slack:C0547N89JUB:1769674617.908059:1769685706.820279` `source_timestamp=2026-01-29T11:23:20Z`
- Migration was rescheduled to Monday at 3 PM BJT due to conflict with an opening ceremony. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_7ac036b36f4fae468f0626cee4689743` `chunk_id=srcchunk_f1602e20542e3033baf078eac01c1da5` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734118.854689` `source_timestamp=2026-01-30T00:48:38Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_c98f258f04ec72173914d2f41cca9dae` `chunk_id=srcchunk_8d612c9c5145cacd90bef4e7284eeafd` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734980.602749` `source_timestamp=2026-01-30T01:03:00Z`

## Open Questions

- Is Grafana monitoring fully integrated for Aeneid validators?

## Sources

- `source_document_id`: `srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d`
- `source_revision_id`: `srcrev_e1292878eb2153155fdd65b0aa674f7a`
