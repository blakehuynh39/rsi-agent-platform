---
title: "Temporal Certificate Renewal Runbook"
type: "runbook"
slug: "runbooks/temporal-certificate-renewal-runbook"
freshness: "2026-01-12T22:51:04Z"
tags:
  - "access-control"
  - "certificate-expiry"
  - "runbook"
  - "staging"
  - "temporal"
owners: []
source_revision_ids:
  - "srcrev_72e1d0f12428312b3517e1ebc75835df"
  - "srcrev_a0a5b6a88650296c2fb535db762adfca"
  - "srcrev_b8cb42cb4ce3f47278b0cd8ef299703b"
  - "srcrev_cf4e57121c4d1b5a5223a61ca9876e1e"
conflict_state: "none"
---

# Temporal Certificate Renewal Runbook

## Summary

Runbook for handling Temporal certificate expiry, including notification, permission escalation, and renewal.

## Claims

- The staging-gcp.koyiy TLS certificate for Temporal is expiring. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8ac415b93a82fd7c1ee6823b17773027` `source_revision_id=srcrev_a0a5b6a88650296c2fb535db762adfca` `chunk_id=srcchunk_becb58d8eff9a1a57346fff523a8d40b` `native_locator=slack:C0547N89JUB:1768201916.033099:1768201916.033099` `source_timestamp=2026-01-12T07:11:56Z`
- Notification was forwarded via email by Andy. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8ac415b93a82fd7c1ee6823b17773027` `source_revision_id=srcrev_a0a5b6a88650296c2fb535db762adfca` `chunk_id=srcchunk_becb58d8eff9a1a57346fff523a8d40b` `native_locator=slack:C0547N89JUB:1768201916.033099:1768201916.033099` `source_timestamp=2026-01-12T07:11:56Z`
- The reporter had only write permissions on the certificate and wanted to ensure tracking. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8ac415b93a82fd7c1ee6823b17773027` `source_revision_id=srcrev_a0a5b6a88650296c2fb535db762adfca` `chunk_id=srcchunk_becb58d8eff9a1a57346fff523a8d40b` `native_locator=slack:C0547N89JUB:1768201916.033099:1768201916.033099` `source_timestamp=2026-01-12T07:11:56Z`
- Due to unfamiliarity with Temporal and mainnet migration workload, another engineer was asked to investigate. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8ac415b93a82fd7c1ee6823b17773027` `source_revision_id=srcrev_72e1d0f12428312b3517e1ebc75835df` `chunk_id=srcchunk_de3dfdca498c99e8bc1ebb532854f62a` `native_locator=slack:C0547N89JUB:1768201916.033099:1768202338.955079` `source_timestamp=2026-01-12T07:21:08Z`
- The investigating engineer requested Temporal admin permissions. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8ac415b93a82fd7c1ee6823b17773027` `source_revision_id=srcrev_b8cb42cb4ce3f47278b0cd8ef299703b` `chunk_id=srcchunk_eff5af4ad8a106f97c0acf3914494565` `native_locator=slack:C0547N89JUB:1768201916.033099:1768256486.533149` `source_timestamp=2026-01-12T22:21:26Z`
- Temporal admin permissions were granted. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8ac415b93a82fd7c1ee6823b17773027` `source_revision_id=srcrev_cf4e57121c4d1b5a5223a61ca9876e1e` `chunk_id=srcchunk_c568eeacf3ef67120de3d5de99fbe85e` `native_locator=slack:C0547N89JUB:1768201916.033099:1768258264.935929` `source_timestamp=2026-01-12T22:51:04Z`

## Open Questions

- Was the Temporal staging-gcp.koyiy certificate successfully renewed?

## Sources

- `source_document_id`: `srcdoc_8ac415b93a82fd7c1ee6823b17773027`
- `source_revision_id`: `srcrev_72e1d0f12428312b3517e1ebc75835df`
