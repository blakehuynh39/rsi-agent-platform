---
title: "Revenue Distribution to Vaults"
type: "concept"
slug: "concepts/revenue-distribution-to-vaults"
freshness: "2025-04-27T01:25:04Z"
tags:
  - "revenue"
  - "royalty"
  - "sdk"
  - "vault"
owners: []
source_revision_ids:
  - "srcrev_60b239a69b58ff15ba75e302cb5369c6"
conflict_state: "none"
---

# Revenue Distribution to Vaults

## Summary

Revenue from royalty payments is automatically distributed to ancestor vaults via payRoyaltyOnBehalf or payLicenseMintingFee, without need for manual transferToVault calls. Claiming revenue requires a separate manual function. This page clarifies the flow.

## Claims

- Revenue distribution to ancestor vaults is automatic when payRoyaltyOnBehalf or payLicenseMintingFee is called. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_60b239a69b58ff15ba75e302cb5369c6` `chunk_id=srcchunk_1ab76a5b56e2b54d8a2adaa10d1dfd48` `native_locator=slack:C04T5307FNU:1745619184.594809:1745717104.323939` `source_timestamp=2025-04-27T01:25:04Z`
- Ancestors do not need to manually call transferToVault to move funds to their vaults. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_60b239a69b58ff15ba75e302cb5369c6` `chunk_id=srcchunk_1ab76a5b56e2b54d8a2adaa10d1dfd48` `native_locator=slack:C04T5307FNU:1745619184.594809:1745717104.323939` `source_timestamp=2025-04-27T01:25:04Z`
- Claiming revenue from a vault requires a separate call to claimAllRevenue or similar functions. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0517a88031549e215fac23163f60df18` `source_revision_id=srcrev_60b239a69b58ff15ba75e302cb5369c6` `chunk_id=srcchunk_1ab76a5b56e2b54d8a2adaa10d1dfd48` `native_locator=slack:C04T5307FNU:1745619184.594809:1745717104.323939` `source_timestamp=2025-04-27T01:25:04Z`

## Open Questions

- What is the use case for transferToVault if revenue distribution to vaults is automatic when payRoyaltyOnBehalf or payLicenseMintingFee is called?
- When might a developer need to call transferToVault directly?

## Sources

- `source_document_id`: `srcdoc_0517a88031549e215fac23163f60df18`
- `source_revision_id`: `srcrev_60b239a69b58ff15ba75e302cb5369c6`
