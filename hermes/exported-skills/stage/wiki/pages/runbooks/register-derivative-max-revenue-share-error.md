---
title: "Error: royalty percent greater than max revenue share when registering derivative"
type: "runbook"
slug: "runbooks/register-derivative-max-revenue-share-error"
freshness: "2025-03-21T00:13:59Z"
tags:
  - "derivative"
  - "error"
  - "ipasset"
  - "sdk"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_3483b392d4ed9f080066535598830d0c"
  - "srcrev_ce1b723dda59c0130a1ac41a099f5842"
  - "srcrev_fe42a3d7f1c412faabb042d44dddd023"
conflict_state: "none"
---

# Error: royalty percent greater than max revenue share when registering derivative

## Summary

When calling registerDerivative with max_revenue_share=100, the SDK validation raises ValueError because the parent IP's royalty percent exceeds the provided max_revenue_share. This occurs even when using SDK defaults.

## Claims

- registerDerivative call with max_revenue_share=100 results in ValueError: 'The royalty percent for the parent IP with id 0x59b8d4007d2194dECEeeaEe4d3187d8899d1e98f is greater than the maximum revenue share 100.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_fe42a3d7f1c412faabb042d44dddd023` `chunk_id=srcchunk_74a437e9f0d91d54f8015dd8523737d6` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505848.636179` `source_timestamp=2025-03-20T21:24:08Z`
- The error originates from IPAsset._validate_derivative_data in the story_protocol_python_sdk. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- The registerDerivative call used parameters: child_ip_id=0x0695882509E74a8a4C4fA81bC583C9e11a80F95A, parent_ip_ids=['0x59b8d4007d2194dECEeeaEe4d3187d8899d1e98f'], license_terms_ids=[955], max_minting_fee=100, max_revenue_share=100, max_rts=100000000. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505848.636179:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- Even when using SDK defaults instead of explicit parameter values, the same error occurs. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_3483b392d4ed9f080066535598830d0c` `chunk_id=srcchunk_df7f62129a98174318fd35c8a19cf4a0` `native_locator=slack:C04T5307FNU:1742505848.636179:1742516039.214519` `source_timestamp=2025-03-21T00:13:59Z`

## Open Questions

- Could this be a documentation mismatch or a bug in the SDK validation?
- Is max_revenue_share expected to be set higher than 100, or is there a different parameter for royalty percent?
- What is the parent IP's royalty percent, and why does it exceed max_revenue_share of 100?

## Sources

- `source_document_id`: `srcdoc_ab4870cee8de03c781abcc48aac5c944`
- `source_revision_id`: `srcrev_ce1b723dda59c0130a1ac41a099f5842`
