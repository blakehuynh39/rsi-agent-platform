---
title: "Aeneid RPC and Validators Migration"
type: "decision"
slug: "decisions/aeneid-rpc-validators-migration"
freshness: "2026-01-30T01:03:00Z"
tags:
  - "aeneid"
  - "migration"
  - "rpc"
  - "scheduling"
  - "sync-issue"
  - "validators"
owners:
  - "U079ZJ48D62"
source_revision_ids:
  - "srcrev_26ab6d9a1433b7c3745d46ec10bec869"
  - "srcrev_58a5055208377193fec7e9069176f109"
  - "srcrev_7d72295e39178558b9556d7301178cc8"
  - "srcrev_9a49ad1db798e9b0ffab1b63889f16bf"
  - "srcrev_a7455a72581ef09e8364b80a5b00178d"
  - "srcrev_c98f258f04ec72173914d2f41cca9dae"
  - "srcrev_e1292878eb2153155fdd65b0aa674f7a"
  - "srcrev_f60559c7cd8d8e7fe373593b9bb57525"
  - "srcrev_f7c55148aacfece9f1c814514d37e881"
conflict_state: "none"
---

# Aeneid RPC and Validators Migration

## Summary

Planned migration of Aeneid RPC and validators, initially scheduled for tomorrow 11PM PT, then rescheduled to Monday 3PM BJT to avoid opening ceremony conflict. Encountered block sync issue on validator node, resolved by manually adding peer.

## Claims

- Migration of Aeneid RPC and validators was initially planned for tomorrow 11 PM PT. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_58a5055208377193fec7e9069176f109` `chunk_id=srcchunk_23eb033bcf78af39a99319d9ae61d766` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674617.908059` `source_timestamp=2026-01-29T11:17:29Z`
- Validator node use1-aeneid-validator4 encountered block sync issues, including 'Number of finalized block is missing' and 'Push finalized payload while evm syncing' warnings. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f60559c7cd8d8e7fe373593b9bb57525` `chunk_id=srcchunk_d211a0c443638f4434efda509e900c54` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678206.437309` `source_timestamp=2026-01-29T09:16:46Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f7c55148aacfece9f1c814514d37e881` `chunk_id=srcchunk_794a1f7289b9b9c443ac9fcff3645248` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678252.258289` `source_timestamp=2026-01-29T09:17:32Z`
- It was suggested that the node might need snap sync. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_7d72295e39178558b9556d7301178cc8` `chunk_id=srcchunk_3197e837388bdcf9d1f818fd7e49395a` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678354.810879` `source_timestamp=2026-01-29T09:19:14Z`
- The sync issue was resolved by manually adding a peer to the node. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a7455a72581ef09e8364b80a5b00178d` `chunk_id=srcchunk_0c93db5007818e1eae81bb9ad758c83d` `native_locator=slack:C0547N89JUB:1769674617.908059:1769685706.820279` `source_timestamp=2026-01-29T11:23:20Z`
- The migration was rescheduled to Monday 3 PM BJT due to a conflict with an opening ceremony. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_c98f258f04ec72173914d2f41cca9dae` `chunk_id=srcchunk_8d612c9c5145cacd90bef4e7284eeafd` `native_locator=slack:C0547N89JUB:1769674617.908059:1769734980.602749` `source_timestamp=2026-01-30T01:03:00Z`
- There was a suggestion to share the migration announcement with partners or on social media. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_e1292878eb2153155fdd65b0aa674f7a` `chunk_id=srcchunk_29a84ab097a5aa2459982476e9fb67d2` `native_locator=slack:C0547N89JUB:1769674617.908059:1769674667.077879` `source_timestamp=2026-01-29T08:17:47Z`
- The validator node is not fully connected to the monitoring system (Grafana). `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_9a49ad1db798e9b0ffab1b63889f16bf` `chunk_id=srcchunk_9ab75628afc0dc3b812cba8e945c3c8e` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678976.250819` `source_timestamp=2026-01-29T09:29:36Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_26ab6d9a1433b7c3745d46ec10bec869` `chunk_id=srcchunk_233fff9679af47ab878cb631a42b97f4` `native_locator=slack:C0547N89JUB:1769674617.908059:1769679337.169059` `source_timestamp=2026-01-29T09:35:37Z`

## Open Questions

- Should the migration be publicly announced to partners/social media?

## Sources

- `source_document_id`: `srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d`
- `source_revision_id`: `srcrev_7c500560f89b8ad6cfaa8ade4b793cae`
