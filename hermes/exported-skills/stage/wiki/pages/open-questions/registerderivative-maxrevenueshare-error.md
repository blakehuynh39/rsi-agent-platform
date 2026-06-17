---
title: "registerDerivative Error: Royalty Percent Exceeds Maximum Revenue Share"
type: "open_question"
slug: "open-questions/registerderivative-maxrevenueshare-error"
freshness: "2025-03-21T00:13:59Z"
tags:
  - "derivative-registration"
  - "error"
  - "royalty"
  - "story-protocol-sdk"
owners: []
source_revision_ids:
  - "srcrev_3483b392d4ed9f080066535598830d0c"
  - "srcrev_ce1b723dda59c0130a1ac41a099f5842"
  - "srcrev_fe42a3d7f1c412faabb042d44dddd023"
conflict_state: "none"
---

# registerDerivative Error: Royalty Percent Exceeds Maximum Revenue Share

## Summary

A user encounters a ValueError when calling registerDerivative: 'The royalty percent for the parent IP with id ... is greater than the maximum revenue share ...', despite setting max_revenue_share=100. The issue persists with SDK defaults. Possible causes include misunderstanding, documentation mismatch, or SDK bug.

## Claims

- The error message displayed is 'The royalty percent for the parent IP with id {parent_id} is greater than the maximum revenue share {internal_data['maxRevenueShare']}.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_fe42a3d7f1c412faabb042d44dddd023` `chunk_id=srcchunk_74a437e9f0d91d54f8015dd8523737d6` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505848.636179` `source_timestamp=2025-03-20T21:24:08Z`
- The error occurred during a call to registerDerivative with parameters: child_ip_id=0x0695882509E74a8a4C4fA81bC583C9e11a80F95A, parent_ip_ids=['0x59b8d4007d2194dECEeeaEe4d3187d8899d1e98f'], license_terms_ids=[955], max_minting_fee=100, max_revenue_share=100, max_rts=100000000. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- The error is raised from the _validate_derivative_data method in IPAsset.py at line 267 of the story-protocol-python-sdk. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- The user reported that using the same parameter values or SDK defaults still results in the same error. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_3483b392d4ed9f080066535598830d0c` `chunk_id=srcchunk_df7f62129a98174318fd35c8a19cf4a0` `native_locator=slack:C04T5307FNU:1742505848.636179:1742516039.214519` `source_timestamp=2025-03-21T00:13:59Z`
- The user suspects possible causes: misunderstanding of the SDK, mismatch with documentation, or a bug in the SDK. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_fe42a3d7f1c412faabb042d44dddd023` `chunk_id=srcchunk_74a437e9f0d91d54f8015dd8523737d6` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505848.636179` `source_timestamp=2025-03-20T21:24:08Z`

## Open Questions

- Why does registerDerivative fail with 'royalty percent for parent IP exceeds max revenue share' even when max_revenue_share is set to 100?

## Sources

- `source_document_id`: `srcdoc_ab4870cee8de03c781abcc48aac5c944`
- `source_revision_id`: `srcrev_3483b392d4ed9f080066535598830d0c`
