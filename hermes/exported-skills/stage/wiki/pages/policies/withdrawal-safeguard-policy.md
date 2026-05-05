---
title: "Withdrawal Safeguard Policy"
type: "policy"
slug: "policies/withdrawal-safeguard-policy"
freshness: "2026-05-04T23:58:02Z"
tags:
  - "fraud-prevention"
  - "payments"
  - "withdrawal"
owners: []
source_revision_ids:
  - "srcrev_71781f0f634caab2dad82ec766528f7d"
conflict_state: "none"
---

# Withdrawal Safeguard Policy

## Summary

Safeguard mechanism to prevent spam and malicious withdrawals by restricting withdrawals to processed tasks and requiring a minimum approved task count.

## Claims

- Users can only withdraw rewards of processed tasks (approved or rejected), they cannot withdraw rewards of unprocessed tasks (ie. submission reward). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_056ec0e3d256e88b47958ac3dab627af` `source_revision_id=srcrev_71781f0f634caab2dad82ec766528f7d` `chunk_id=srcchunk_617e73aa68c20ea1ff6819dd9b2eaf4c` `native_locator=https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4` `source_timestamp=2026-05-04T23:58:02Z`
- Users must have X amount of approved tasks to withdraw, otherwise we flag as spam account and ban or show spam warnings. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_056ec0e3d256e88b47958ac3dab627af` `source_revision_id=srcrev_71781f0f634caab2dad82ec766528f7d` `chunk_id=srcchunk_617e73aa68c20ea1ff6819dd9b2eaf4c` `native_locator=https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4` `source_timestamp=2026-05-04T23:58:02Z`
- This safeguard is a P1 priority measure. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_056ec0e3d256e88b47958ac3dab627af` `source_revision_id=srcrev_71781f0f634caab2dad82ec766528f7d` `chunk_id=srcchunk_617e73aa68c20ea1ff6819dd9b2eaf4c` `native_locator=https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4` `source_timestamp=2026-05-04T23:58:02Z`

## Open Questions

- What is the minimum number of approved tasks required to withdraw (X)?
- Will the system ban users or show spam warnings upon insufficient approved tasks?

## Sources

- `source_document_id`: `srcdoc_056ec0e3d256e88b47958ac3dab627af`
- `source_revision_id`: `srcrev_71781f0f634caab2dad82ec766528f7d`
- `source_url`: https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4
