---
title: "Aeneid Migration"
type: "project"
slug: "projects/aeneid-migration"
freshness: "2026-01-30T01:03:00Z"
tags:
  - "aeneid"
  - "migration"
  - "rpc"
  - "validators"
owners: []
source_revision_ids:
  - "srcrev_58a5055208377193fec7e9069176f109"
  - "srcrev_610ef8eff4a2acaa934d039cc88780ef"
  - "srcrev_7ac036b36f4fae468f0626cee4689743"
  - "srcrev_a27d6860bc5a97fd0b5f61b1d3941711"
  - "srcrev_a7455a72581ef09e8364b80a5b00178d"
  - "srcrev_b8dad48e3109a47f6c713bc38fd5fd48"
  - "srcrev_c98f258f04ec72173914d2f41cca9dae"
  - "srcrev_dba472f4960509786d97616071b002f7"
  - "srcrev_f5ef5c14cf84a9f2520305dc40cc0b93"
  - "srcrev_f60559c7cd8d8e7fe373593b9bb57525"
  - "srcrev_f7c55148aacfece9f1c814514d37e881"
conflict_state: "none"
---

# Aeneid Migration

## Summary

Migration of Aeneid RPC and validators, initially planned for 11PM PT, encountered sync issues due to peer configuration, and was rescheduled multiple times due to date confusion and conflict with an opening ceremony, finally settled for Monday 3PM BJT.

## Claims

- The migration of Aeneid RPC and validators was initially announced to occur the next day at 11PM PT without downtime. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_58a5055208377193fec7e9069176f109` `chunk_id=srcchunk_23eb033bcf78af39a99319d9ae61d766` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674617.908059` `source_timestamp=2026-01-29T11:17:29Z`
- After a resync attempt on validator4, errors appeared including 'Number of finalized block is missing' from geth and 'execution engine is syncing' warnings from the consensus layer. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f60559c7cd8d8e7fe373593b9bb57525` `chunk_id=srcchunk_d211a0c443638f4434efda509e900c54` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678206.437309` `source_timestamp=2026-01-29T09:16:46Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f7c55148aacfece9f1c814514d37e881` `chunk_id=srcchunk_794a1f7289b9b9c443ac9fcff3645248` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678252.258289` `source_timestamp=2026-01-29T09:17:32Z`
- The sync issue was identified as a peer problem, and after manually adding a peer, everything worked fine. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a7455a72581ef09e8364b80a5b00178d` `chunk_id=srcchunk_0c93db5007818e1eae81bb9ad758c83d` `native_locator=slack:C0547N89JUB:1769674617.908059:1769685706.820279` `source_timestamp=2026-01-29T11:23:20Z`
- Confusion arose that the migration date had been set incorrectly and was already past; the date was changed to 'today'. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_b8dad48e3109a47f6c713bc38fd5fd48` `chunk_id=srcchunk_2eeee0b00c4a41cadf92132772d9e842` `native_locator=slack:C0547N89JUB:1769674617.908059:1769696549.060869` `source_timestamp=2026-01-29T14:22:29Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_610ef8eff4a2acaa934d039cc88780ef` `chunk_id=srcchunk_148071300eebf431f70d10d4252da234` `native_locator=slack:C0547N89JUB:1769674617.908059:1769724531.593729` `source_timestamp=2026-01-29T22:08:51Z`
- The migration was then rescheduled to avoid an opening ceremony conflict, with a proposal to move earlier, but ultimately agreed to Monday 3PM BJT. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_7ac036b36f4fae468f0626cee4689743` `chunk_id=srcchunk_f1602e20542e3033baf078eac01c1da5` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734118.854689` `source_timestamp=2026-01-30T00:48:38Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f5ef5c14cf84a9f2520305dc40cc0b93` `chunk_id=srcchunk_516c8fadfb81a4df7fd4835f713a6264` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734207.472069` `source_timestamp=2026-01-30T00:50:07Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a27d6860bc5a97fd0b5f61b1d3941711` `chunk_id=srcchunk_88539399bcf9757cb54c4b11a2d0d50e` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734797.849139` `source_timestamp=2026-01-30T00:59:57Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_dba472f4960509786d97616071b002f7` `chunk_id=srcchunk_a4d1414a2d1c51500fb2c5e62de4cc0c` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734871.062749` `source_timestamp=2026-01-30T01:01:11Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_c98f258f04ec72173914d2f41cca9dae` `chunk_id=srcchunk_8d612c9c5145cacd90bef4e7284eeafd` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734980.602749` `source_timestamp=2026-01-30T01:03:00Z`

## Sources

- `source_document_id`: `srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d`
- `source_revision_id`: `srcrev_b8dad48e3109a47f6c713bc38fd5fd48`
