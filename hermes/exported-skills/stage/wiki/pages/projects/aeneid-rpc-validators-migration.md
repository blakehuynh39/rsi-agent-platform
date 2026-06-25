---
title: "Aeneid RPC and Validators Migration"
type: "project"
slug: "projects/aeneid-rpc-validators-migration"
freshness: "2026-01-30T01:03:00Z"
tags:
  - "aeneid"
  - "migration"
  - "rpc"
  - "validators"
owners:
  - "U079ZJ48D62"
  - "U07TNT9N4JC"
source_revision_ids:
  - "srcrev_161498e4abc3a93f291bac397c708d53"
  - "srcrev_26ab6d9a1433b7c3745d46ec10bec869"
  - "srcrev_41a04ac561f3e2c3a235f4af56b7243e"
  - "srcrev_58a5055208377193fec7e9069176f109"
  - "srcrev_610ef8eff4a2acaa934d039cc88780ef"
  - "srcrev_7ac036b36f4fae468f0626cee4689743"
  - "srcrev_9a49ad1db798e9b0ffab1b63889f16bf"
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

# Aeneid RPC and Validators Migration

## Summary

Planned migration of Aeneid RPC and validators, initially scheduled for January 30, 2026 11 PM PT, postponed due to sync issues and event conflict, rescheduled to Monday at 3 PM BJT. Sync issues were caused by a peer connectivity problem and resolved by manual peer addition.

## Claims

- Migration of Aeneid RPC and validators was initially scheduled for January 30, 2026 at 11 PM PT with zero downtime promised. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_58a5055208377193fec7e9069176f109` `chunk_id=srcchunk_23eb033bcf78af39a99319d9ae61d766` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674617.908059` `source_timestamp=2026-01-29T11:17:29Z`
- A validator node encountered sync errors including 'Number of finalized block is missing' and warnings about pushing finalized payload while the execution engine was syncing. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f60559c7cd8d8e7fe373593b9bb57525` `chunk_id=srcchunk_d211a0c443638f4434efda509e900c54` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678206.437309` `source_timestamp=2026-01-29T09:16:46Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f7c55148aacfece9f1c814514d37e881` `chunk_id=srcchunk_794a1f7289b9b9c443ac9fcff3645248` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678252.258289` `source_timestamp=2026-01-29T09:17:32Z`
- The node was synced using snapshot data, but issues with EL block data persisted even after re-uploading the snapshot. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_41a04ac561f3e2c3a235f4af56b7243e` `chunk_id=srcchunk_5cc0ce899fa53a6040fce0ce67d7b391` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678508.876709` `source_timestamp=2026-01-29T09:21:48Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_161498e4abc3a93f291bac397c708d53` `chunk_id=srcchunk_641613ff5ed603f1a591e2b4c7ad0dbc` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678608.112989` `source_timestamp=2026-01-29T09:23:28Z`
- The validator node is not fully connected to the Grafana monitoring system, limiting log visibility. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_9a49ad1db798e9b0ffab1b63889f16bf` `chunk_id=srcchunk_9ab75628afc0dc3b812cba8e945c3c8e` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678976.250819` `source_timestamp=2026-01-29T09:29:36Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_26ab6d9a1433b7c3745d46ec10bec869` `chunk_id=srcchunk_233fff9679af47ab878cb631a42b97f4` `native_locator=slack:C0547N89JUB:1769674617.908059:1769679337.169059` `source_timestamp=2026-01-29T09:35:37Z`
- The root cause of the sync issues was a peer connectivity problem; manually adding a peer resolved the issue, allowing the migration to proceed. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a7455a72581ef09e8364b80a5b00178d` `chunk_id=srcchunk_0c93db5007818e1eae81bb9ad758c83d` `native_locator=slack:C0547N89JUB:1769674617.908059:1769685706.820279` `source_timestamp=2026-01-29T11:23:20Z`
- The initial migration date was considered incorrect by a team member, and the date was changed to the current day. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_b8dad48e3109a47f6c713bc38fd5fd48` `chunk_id=srcchunk_2eeee0b00c4a41cadf92132772d9e842` `native_locator=slack:C0547N89JUB:1769674617.908059:1769696549.060869` `source_timestamp=2026-01-29T14:22:29Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_610ef8eff4a2acaa934d039cc88780ef` `chunk_id=srcchunk_148071300eebf431f70d10d4252da234` `native_locator=slack:C0547N89JUB:1769674617.908059:1769724531.593729` `source_timestamp=2026-01-29T22:08:51Z`
- Due to a conflict with an 'opening ceremony', the migration was rescheduled to Monday at 3 PM BJT (Beijing Time). `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_7ac036b36f4fae468f0626cee4689743` `chunk_id=srcchunk_f1602e20542e3033baf078eac01c1da5` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734118.854689` `source_timestamp=2026-01-30T00:48:38Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f5ef5c14cf84a9f2520305dc40cc0b93` `chunk_id=srcchunk_516c8fadfb81a4df7fd4835f713a6264` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734207.472069` `source_timestamp=2026-01-30T00:50:07Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a27d6860bc5a97fd0b5f61b1d3941711` `chunk_id=srcchunk_88539399bcf9757cb54c4b11a2d0d50e` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734797.849139` `source_timestamp=2026-01-30T00:59:57Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_dba472f4960509786d97616071b002f7` `chunk_id=srcchunk_a4d1414a2d1c51500fb2c5e62de4cc0c` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734871.062749` `source_timestamp=2026-01-30T01:01:11Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_c98f258f04ec72173914d2f41cca9dae` `chunk_id=srcchunk_8d612c9c5145cacd90bef4e7284eeafd` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734980.602749` `source_timestamp=2026-01-30T01:03:00Z`

## Sources

- `source_document_id`: `srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d`
- `source_revision_id`: `srcrev_f60559c7cd8d8e7fe373593b9bb57525`
