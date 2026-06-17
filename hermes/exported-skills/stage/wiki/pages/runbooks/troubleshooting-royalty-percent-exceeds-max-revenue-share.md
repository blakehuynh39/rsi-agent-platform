---
title: "Troubleshooting: 'Royalty percent for parent IP exceeds max revenue share' Error"
type: "runbook"
slug: "runbooks/troubleshooting-royalty-percent-exceeds-max-revenue-share"
freshness: "2025-03-21T00:13:59Z"
tags:
  - "ipasset"
  - "python"
  - "register-derivative"
  - "sdk-error"
owners: []
source_revision_ids:
  - "srcrev_3483b392d4ed9f080066535598830d0c"
  - "srcrev_ce1b723dda59c0130a1ac41a099f5842"
  - "srcrev_fe42a3d7f1c412faabb042d44dddd023"
conflict_state: "none"
---

# Troubleshooting: 'Royalty percent for parent IP exceeds max revenue share' Error

## Summary

How to diagnose and resolve the ValueError: 'The royalty percent for the parent IP is greater than the maximum revenue share' when calling registerDerivative.

## Claims

- The SDK’s registerDerivative method validates that the parent IP’s royalty percent does not exceed the max_revenue_share parameter. If the parent’s royalty is higher, a ValueError is raised with message: 'The royalty percent for the parent IP with id {parent_id} is greater than the maximum revenue share {internal_data['maxRevenueShare']}.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_fe42a3d7f1c412faabb042d44dddd023` `chunk_id=srcchunk_74a437e9f0d91d54f8015dd8523737d6` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505848.636179` `source_timestamp=2025-03-20T21:24:08Z`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- In a reported case, the error occurred with child IP 0x0695882509E74a8a4C4fA81bC583C9e11a80F95A, parent IP 0x59b8d4007d2194dECEeeaEe4d3187d8899d1e98f, license terms 955, max_minting_fee 100, max_revenue_share 100, max_rts 100000000. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- The issue persisted even when using SDK defaults or the previously attempted parameter values. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_3483b392d4ed9f080066535598830d0c` `chunk_id=srcchunk_df7f62129a98174318fd35c8a19cf4a0` `native_locator=slack:C04T5307FNU:1742505848.636179:1742516039.214519` `source_timestamp=2025-03-21T00:13:59Z`

## Open Questions

- Why does the parent IP have a royalty percent that exceeds the provided max_revenue_share? Is it a misconfiguration, a scaling mismatch, or an expected constraint on derivative registration?

## Sources

- `source_document_id`: `srcdoc_ab4870cee8de03c781abcc48aac5c944`
- `source_revision_id`: `srcrev_fe42a3d7f1c412faabb042d44dddd023`
