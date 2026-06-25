---
title: "Aeneid RPC and Validators Migration"
type: "project"
slug: "projects/aeneid-migration"
freshness: "2026-01-30T01:03:00Z"
tags:
  - "aeneid"
  - "blockchain"
  - "migration"
  - "validators"
owners:
  - "U079ZJ48D62"
source_revision_ids:
  - "srcrev_26ab6d9a1433b7c3745d46ec10bec869"
  - "srcrev_41a04ac561f3e2c3a235f4af56b7243e"
  - "srcrev_58a5055208377193fec7e9069176f109"
  - "srcrev_610ef8eff4a2acaa934d039cc88780ef"
  - "srcrev_7ac036b36f4fae468f0626cee4689743"
  - "srcrev_7d72295e39178558b9556d7301178cc8"
  - "srcrev_a27d6860bc5a97fd0b5f61b1d3941711"
  - "srcrev_a7455a72581ef09e8364b80a5b00178d"
  - "srcrev_b8dad48e3109a47f6c713bc38fd5fd48"
  - "srcrev_c98f258f04ec72173914d2f41cca9dae"
  - "srcrev_f60559c7cd8d8e7fe373593b9bb57525"
  - "srcrev_f7c55148aacfece9f1c814514d37e881"
conflict_state: "none"
---

# Aeneid RPC and Validators Migration

## Summary

Migration of Aeneid RPC and validators, planned with no downtime. A syncing issue was resolved by adding peers manually. The date was adjusted from Jan 30 to Jan 26 initially, then to Monday Jan 28 3PM BJT to avoid a conflict with an opening ceremony.

## Claims

- Migration was initially planned for tomorrow (approximately 2026-01-30 11PM PT) with no downtime. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_58a5055208377193fec7e9069176f109` `chunk_id=srcchunk_23eb033bcf78af39a99319d9ae61d766` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674617.908059` `source_timestamp=2026-01-29T11:17:29Z`
- Syncing issue observed on use1-aeneid-validator4: geth reported 'Number of finalized block is missing' and cosmovisor logged warnings about processing finalized payload while execution engine syncing. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f60559c7cd8d8e7fe373593b9bb57525` `chunk_id=srcchunk_d211a0c443638f4434efda509e900c54` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678206.437309` `source_timestamp=2026-01-29T09:16:46Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f7c55148aacfece9f1c814514d37e881` `chunk_id=srcchunk_794a1f7289b9b9c443ac9fcff3645248` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678252.258289` `source_timestamp=2026-01-29T09:17:32Z`
- Troubleshooting included checking snap sync, suspecting EL block data issues, and noting that monitoring was not fully connected. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_7d72295e39178558b9556d7301178cc8` `chunk_id=srcchunk_3197e837388bdcf9d1f818fd7e49395a` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678354.810879` `source_timestamp=2026-01-29T09:19:14Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_41a04ac561f3e2c3a235f4af56b7243e` `chunk_id=srcchunk_5cc0ce899fa53a6040fce0ce67d7b391` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678508.876709` `source_timestamp=2026-01-29T09:21:48Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_26ab6d9a1433b7c3745d46ec10bec869` `chunk_id=srcchunk_233fff9679af47ab878cb631a42b97f4` `native_locator=slack:C0547N89JUB:1769674617.908059:1769679337.169059` `source_timestamp=2026-01-29T09:35:37Z`
- The syncing issue was resolved by manually adding peers. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a7455a72581ef09e8364b80a5b00178d` `chunk_id=srcchunk_0c93db5007818e1eae81bb9ad758c83d` `native_locator=slack:C0547N89JUB:1769674617.908059:1769685706.820279` `source_timestamp=2026-01-29T11:23:20Z`
- The scheduled date was found to be incorrect and was changed to today (Jan 26, 2026). `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_b8dad48e3109a47f6c713bc38fd5fd48` `chunk_id=srcchunk_2eeee0b00c4a41cadf92132772d9e842` `native_locator=slack:C0547N89JUB:1769674617.908059:1769696549.060869` `source_timestamp=2026-01-29T14:22:29Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_610ef8eff4a2acaa934d039cc88780ef` `chunk_id=srcchunk_148071300eebf431f70d10d4252da234` `native_locator=slack:C0547N89JUB:1769674617.908059:1769724531.593729` `source_timestamp=2026-01-29T22:08:51Z`
- Due to a conflict with an opening ceremony, the migration was rescheduled to Monday, Jan 28, 2026, 3PM BJT. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_7ac036b36f4fae468f0626cee4689743` `chunk_id=srcchunk_f1602e20542e3033baf078eac01c1da5` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734118.854689` `source_timestamp=2026-01-30T00:48:38Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a27d6860bc5a97fd0b5f61b1d3941711` `chunk_id=srcchunk_88539399bcf9757cb54c4b11a2d0d50e` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734797.849139` `source_timestamp=2026-01-30T00:59:57Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_c98f258f04ec72173914d2f41cca9dae` `chunk_id=srcchunk_8d612c9c5145cacd90bef4e7284eeafd` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734980.602749` `source_timestamp=2026-01-30T01:03:00Z`

## Open Questions

- Monitoring integration for Aeneid validators is not fully complete (as of Jan 29).

## Sources

- `source_document_id`: `srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d`
- `source_revision_id`: `srcrev_26ab6d9a1433b7c3745d46ec10bec869`
