---
title: "Aeneid RPC and Validator Migration"
type: "decision"
slug: "decisions/aeneid-rpc-validator-migration"
freshness: "2026-01-30T01:03:00Z"
tags:
  - "aeneid"
  - "migration"
  - "peer"
  - "rpc"
  - "sync"
  - "validator"
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
  - "srcrev_a7455a72581ef09e8364b80a5b00178d"
  - "srcrev_c98f258f04ec72173914d2f41cca9dae"
  - "srcrev_e1292878eb2153155fdd65b0aa674f7a"
  - "srcrev_f60559c7cd8d8e7fe373593b9bb57525"
conflict_state: "none"
---

# Aeneid RPC and Validator Migration

## Summary

Planned migration of Aeneid RPC and validators to new infrastructure, initially scheduled for 2026-01-29 23:00 PT, but rescheduled to Monday 3PM BJT (2026-02-02?) to resolve sync issues and avoid conflict with opening ceremony.

## Claims

- Initial plan: migrate Aeneid RPC and validators on 2026-01-30 11PM PT without downtime. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_58a5055208377193fec7e9069176f109` `chunk_id=srcchunk_23eb033bcf78af39a99319d9ae61d766` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674617.908059` `source_timestamp=2026-01-29T11:17:29Z`
- Validator node encountered 'Number of finalized block is missing' error during sync, despite block sync appearing to work after resync. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f60559c7cd8d8e7fe373593b9bb57525` `chunk_id=srcchunk_d211a0c443638f4434efda509e900c54` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678206.437309` `source_timestamp=2026-01-29T09:16:46Z`
- Node sync issue was caused by missing peers; after manually adding a peer, the node synced correctly. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a7455a72581ef09e8364b80a5b00178d` `chunk_id=srcchunk_0c93db5007818e1eae81bb9ad758c83d` `native_locator=slack:C0547N89JUB:1769674617.908059:1769685706.820279` `source_timestamp=2026-01-29T11:23:20Z`
- Migration date was questioned as possibly set to the wrong day, then changed to today (Jan 29) before being moved to Monday 3PM BJT to avoid conflict with opening ceremony. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_c98f258f04ec72173914d2f41cca9dae` `chunk_id=srcchunk_8d612c9c5145cacd90bef4e7284eeafd` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734980.602749` `source_timestamp=2026-01-30T01:03:00Z`
- Concern was raised about sharing migration plans with partners or on social media. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_e1292878eb2153155fdd65b0aa674f7a` `chunk_id=srcchunk_29a84ab097a5aa2459982476e9fb67d2` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674667.077879` `source_timestamp=2026-01-29T08:17:47Z`

## Open Questions

- Will the migration proceed on Monday without further sync or network issues?

## Sources

- `source_document_id`: `srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d`
- `source_revision_id`: `srcrev_c98f258f04ec72173914d2f41cca9dae`
