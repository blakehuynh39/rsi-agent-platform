---
title: "transferToVault Function in Royalty Module"
type: "concept"
slug: "concepts/royalty-transfer-to-vault"
freshness: "2025-04-27T01:25:04Z"
tags:
  - "revenue"
  - "royalty"
  - "sdk"
  - "smart-contracts"
owners: []
source_revision_ids:
  - "srcrev_60b239a69b58ff15ba75e302cb5369c6"
  - "srcrev_a23c24e3d077fa7197d70607dbab1920"
  - "srcrev_bc1daebc3b34cf79b17e47123bb936a1"
conflict_state: "none"
---

# transferToVault Function in Royalty Module

## Summary

Clarifies the purpose and usage of the transferToVault function within the Royalty Module, including how automatic revenue distribution and manual claiming work.

## Claims

- transferToVault transfers an amount of revenue tokens to a vault, making them claimable via a royalty policy. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_bc1daebc3b34cf79b17e47123bb936a1` `chunk_id=srcchunk_ecb8cd3418ef35f02f589f53ed03fd6f` `native_locator=slack:C04T5307FNU:1745619184.594809:1745619184.594809` `source_timestamp=2025-04-25T22:14:12Z`
- When payRoyaltyOnBehalf or payLicenseMintingFee is called, revenue is automatically distributed to ancestor vaults; a manual transferToVault call is not required for distribution. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_60b239a69b58ff15ba75e302cb5369c6` `chunk_id=srcchunk_1ab76a5b56e2b54d8a2adaa10d1dfd48` `native_locator=slack:C04T5307FNU:1745619184.594809:1745717104.323939` `source_timestamp=2025-04-27T01:25:04Z`
- Despite automatic distribution, claiming revenue from a vault still requires a separate manual call (e.g., claimAllRevenue). `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_60b239a69b58ff15ba75e302cb5369c6` `chunk_id=srcchunk_1ab76a5b56e2b54d8a2adaa10d1dfd48` `native_locator=slack:C04T5307FNU:1745619184.594809:1745717104.323939` `source_timestamp=2025-04-27T01:25:04Z`
- It is unclear whether a developer ever needs to directly call transferToVault to move revenue from a policy to a vault prior to a claim. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_a23c24e3d077fa7197d70607dbab1920` `chunk_id=srcchunk_b33bc684a94f3b2d2e3e9703e103b861` `native_locator=slack:C04T5307FNU:1745619184.594809:1745687493.724419` `source_timestamp=2025-04-26T17:11:33Z`

## Open Questions

- In which scenarios, if any, must a developer call transferToVault directly? Is it ever needed for pre-claim fund movement?

## Related Pages

- `claimallrevenue-/-claimrevenueonbehalf`
- `payroyaltyonbehalf-/-paylicensemintingfee`
- `royalty-module-overview`

## Sources

- `source_document_id`: `srcdoc_0517a88031549e215fac23163f60df18`
- `source_revision_id`: `srcrev_c36cc5db3c589915edf7e355f484afd7`
