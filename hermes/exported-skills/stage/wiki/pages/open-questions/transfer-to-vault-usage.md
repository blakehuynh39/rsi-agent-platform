---
title: "Usage of transferToVault in Royalty Module"
type: "open_question"
slug: "open-questions/transfer-to-vault-usage"
freshness: "2025-04-27T01:25:04Z"
tags:
  - "revenue"
  - "royalty"
  - "transferToVault"
  - "vault"
owners: []
source_revision_ids:
  - "srcrev_3dd9b7e875d9ec770d664e6bbfecb64c"
  - "srcrev_60b239a69b58ff15ba75e302cb5369c6"
  - "srcrev_bc1daebc3b34cf79b17e47123bb936a1"
conflict_state: "none"
---

# Usage of transferToVault in Royalty Module

## Summary

Clarifying when and whether developers or users need to call transferToVault directly.

## Claims

- transferToVault transfers an amount of revenue tokens claimable via a royalty policy. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_bc1daebc3b34cf79b17e47123bb936a1` `chunk_id=srcchunk_ecb8cd3418ef35f02f589f53ed03fd6f` `native_locator=slack:C04T5307FNU:1745619184.594809:1745619184.594809` `source_timestamp=2025-04-25T22:14:12Z`
- Revenue distribution to ancestor vaults is automatic when payRoyaltyOnBehalf or payLicenseMintingFee is called, and ancestors do not need to manually call transferToVault; claiming revenue from a vault requires a separate call to claimAllRevenue or similar functions. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_60b239a69b58ff15ba75e302cb5369c6` `chunk_id=srcchunk_1ab76a5b56e2b54d8a2adaa10d1dfd48` `native_locator=slack:C04T5307FNU:1745619184.594809:1745717104.323939` `source_timestamp=2025-04-27T01:25:04Z`
- transferToVault may be handled internally by claimRevenueOnBehalf. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_3dd9b7e875d9ec770d664e6bbfecb64c` `chunk_id=srcchunk_6144ddc67283f4cba2c05d91d89d229f` `native_locator=slack:C04T5307FNU:1745619184.594809:1745684903.918709` `source_timestamp=2025-04-26T16:28:23Z`

## Open Questions

- Are there edge cases where automatic distribution fails and manual transferToVault is needed?
- Does claimRevenueOnBehalf internally call transferToVault?
- Under what circumstances, if any, must a developer call transferToVault directly?

## Sources

- `source_document_id`: `srcdoc_0517a88031549e215fac23163f60df18`
- `source_revision_id`: `srcrev_a23c24e3d077fa7197d70607dbab1920`
