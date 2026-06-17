---
title: "registerDerivative Error: Royalty Percent Exceeds Max Revenue Share"
type: "runbook"
slug: "runbooks/register-derivative-royalty-percent-exceeds-max-revenue-share"
freshness: "2025-03-21T00:13:59Z"
tags:
  - "error"
  - "ipasset"
  - "registerDerivative"
  - "sdk"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_3483b392d4ed9f080066535598830d0c"
  - "srcrev_ce1b723dda59c0130a1ac41a099f5842"
conflict_state: "none"
---

# registerDerivative Error: Royalty Percent Exceeds Max Revenue Share

## Summary

Troubleshooting the ValueError 'The royalty percent for the parent IP with id ... is greater than the maximum revenue share ...' when calling registerDerivative via the Story Protocol Python SDK.

## Claims

- The error message is: 'The royalty percent for the parent IP with id {parent_id} is greater than the maximum revenue share {internal_data['maxRevenueShare']}.' `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- The error originates in the `_validate_derivative_data` method of `IPAsset.py` at line 267. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- The user called `registerDerivative` with `max_revenue_share=100`, `child_ip_id='0x0695882509E74a8a4C4fA81bC583C9e11a80F95A'`, `parent_ip_ids=['0x59b8d4007d2194dECEeeaEe4d3187d8899d1e98f']`, `license_terms_ids=[955]`, and `max_rts=100000000`. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`
- Even when using SDK default values, the error persists, suggesting a deeper issue rather than a parameter misconfiguration. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_3483b392d4ed9f080066535598830d0c` `chunk_id=srcchunk_df7f62129a98174318fd35c8a19cf4a0` `native_locator=slack:C04T5307FNU:1742516039.214519` `source_timestamp=2025-03-21T00:13:59Z`
- The parent IP is at address 0x59b8d4007d2194dECEeeaEe4d3187d8899d1e98f and the child IP at 0x0695882509E74a8a4C4fA81bC583C9e11a80F95A on Aeneid testnet. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab4870cee8de03c781abcc48aac5c944` `source_revision_id=srcrev_ce1b723dda59c0130a1ac41a099f5842` `chunk_id=srcchunk_90d5630c1eb3a4107dd1aaa9dcef0cd3` `native_locator=slack:C04T5307FNU:1742505876.310699` `source_timestamp=2025-03-20T21:24:36Z`

## Open Questions

- How can this error be resolved? Does the parent IP need to be reconfigured, or is there a workaround?
- Why does the parent IP's royalty percentage exceed 100%? Is it due to an incorrect on-chain configuration or a bug in the SDK's royalty calculation?

## Sources

- `source_document_id`: `srcdoc_ab4870cee8de03c781abcc48aac5c944`
- `source_revision_id`: `srcrev_dc56a099ec9ed831440d51b32b629540`
