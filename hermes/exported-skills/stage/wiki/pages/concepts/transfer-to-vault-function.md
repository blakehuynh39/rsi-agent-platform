---
title: "transferToVault Function"
type: "concept"
slug: "concepts/transfer-to-vault-function"
freshness: "2025-04-27T01:25:04Z"
tags:
  - "revenue-claiming"
  - "royalty-module"
  - "sdk"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_3dd9b7e875d9ec770d664e6bbfecb64c"
  - "srcrev_60b239a69b58ff15ba75e302cb5369c6"
  - "srcrev_a23c24e3d077fa7197d70607dbab1920"
  - "srcrev_bc1daebc3b34cf79b17e47123bb936a1"
conflict_state: "none"
---

# transferToVault Function

## Summary

The transferToVault function is part of the Story Protocol royalty module. It transfers revenue tokens to a vault for claiming. However, it is typically handled automatically during revenue distribution calls like payRoyaltyOnBehalf or payLicenseMintingFee, and direct user call is generally unnecessary. The automatic distribution occurs to ancestor vaults, while claiming revenue from the vault requires a separate manual call.

## Claims

- transferToVault is a function in the Story Protocol royalty module that transfers an amount of revenue tokens to a vault claimable via a royalty policy. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_bc1daebc3b34cf79b17e47123bb936a1` `chunk_id=srcchunk_ecb8cd3418ef35f02f589f53ed03fd6f` `native_locator=slack:C04T5307FNU:1745619184.594809:1745619184.594809` `source_timestamp=2025-04-25T22:14:12Z`
- The function is likely handled under the hood during claimRevenueOnBehalf, and direct user calls may not be necessary. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_3dd9b7e875d9ec770d664e6bbfecb64c` `chunk_id=srcchunk_6144ddc67283f4cba2c05d91d89d229f` `native_locator=slack:C04T5307FNU:1745619184.594809:1745684903.918709` `source_timestamp=2025-04-26T16:28:23Z`
- At the developer level, it is unclear if a direct call to transferToVault is ever required before claiming revenue; this remains an open question. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_a23c24e3d077fa7197d70607dbab1920` `chunk_id=srcchunk_b33bc684a94f3b2d2e3e9703e103b861` `native_locator=slack:C04T5307FNU:1745619184.594809:1745687493.724419` `source_timestamp=2025-04-26T17:11:33Z`
- Automatic distribution of revenue to ancestor vaults happens when payRoyaltyOnBehalf or payLicenseMintingFee is called, so ancestors do not need to manually call transferToVault. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_60b239a69b58ff15ba75e302cb5369c6` `chunk_id=srcchunk_1ab76a5b56e2b54d8a2adaa10d1dfd48` `native_locator=slack:C04T5307FNU:1745619184.594809:1745717104.323939` `source_timestamp=2025-04-27T01:25:04Z`
- Despite automatic distribution, claiming revenue from a vault requires a separate manual call to functions like claimAllRevenue. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_60b239a69b58ff15ba75e302cb5369c6` `chunk_id=srcchunk_1ab76a5b56e2b54d8a2adaa10d1dfd48` `native_locator=slack:C04T5307FNU:1745619184.594809:1745717104.323939` `source_timestamp=2025-04-27T01:25:04Z`

## Open Questions

- Is transferToVault ever required to be called directly by dapp developers, or is it entirely internal to the royalty module?

## Sources

- `source_document_id`: `srcdoc_0517a88031549e215fac23163f60df18`
- `source_revision_id`: `srcrev_0c81803a577d34d08f1ad49b8e178bd4`
