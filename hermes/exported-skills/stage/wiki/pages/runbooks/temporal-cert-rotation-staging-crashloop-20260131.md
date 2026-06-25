---
title: "Temporal Certificate Rotation Caused Staging Pod Crashlooping (2026-01-31)"
type: "runbook"
slug: "runbooks/temporal-cert-rotation-staging-crashloop-20260131"
freshness: "2026-01-31T06:58:36Z"
tags:
  - "certificates"
  - "crashloop"
  - "incident"
  - "staging"
  - "temporal"
owners:
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_002d5d0807c3e49b40c88f24d001d73e"
  - "srcrev_00762005d6d4ff7ed686566a00c836e6"
  - "srcrev_976300a02790fd3b6b5c3252237280cb"
  - "srcrev_b72afba2c732d2e1198ca28d5e4e08fd"
  - "srcrev_bfb7b465804b59bd1da96a254d58c6d2"
  - "srcrev_bff2f894b5d6b822b67c485d1e361b37"
  - "srcrev_c95dcaaddf2ee33494a9b886516d3264"
  - "srcrev_d92d027b76a3e8d252b0a61c76337458"
conflict_state: "none"
---

# Temporal Certificate Rotation Caused Staging Pod Crashlooping (2026-01-31)

## Summary

On 2026-01-31, rotating Temporal worker certificates on the staging cluster led to all staging pods crash looping. The issue was resolved by restarting deployments after updating vault secrets. This page documents the timeline and root cause analysis.

## Claims

- All staging pods were crash looping on 2026-01-31. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bfb7b465804b59bd1da96a254d58c6d2` `chunk_id=srcchunk_67d04dede4e89406df966aaad7b93645` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841556.201809` `source_timestamp=2026-01-31T06:39:16Z`
- U08332YRB7W rotated the Temporal certificates on staging. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bff2f894b5d6b822b67c485d1e361b37` `chunk_id=srcchunk_88bdae9c12e3823413d87116d3d5f9b4` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841992.302049` `source_timestamp=2026-01-31T06:46:32Z`
- The Vault secrets for staging were updated with the new leaf certificates. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_002d5d0807c3e49b40c88f24d001d73e` `chunk_id=srcchunk_c82f57b42460e9df95ddc49d9cfb68fe` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842505.278199` `source_timestamp=2026-01-31T06:55:05Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_00762005d6d4ff7ed686566a00c836e6` `chunk_id=srcchunk_4838d7291799f92b202cf5b60ee57971` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842594.572599` `source_timestamp=2026-01-31T06:56:34Z`
- After restarting all deployments, the pods became healthy. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_c95dcaaddf2ee33494a9b886516d3264` `chunk_id=srcchunk_06c82f6a8d18881f2fe4fb9696651046` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842713.942099` `source_timestamp=2026-01-31T06:58:33Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d92d027b76a3e8d252b0a61c76337458` `chunk_id=srcchunk_3c0472c4b7a73e5a7201af82ee83386e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842716.717719` `source_timestamp=2026-01-31T06:58:36Z`
- A Cloud SQL SSL/TLS server certificate expiration warning was received but was not the cause of the staging crashloop, as it was still 90 days from expiry. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_b72afba2c732d2e1198ca28d5e4e08fd` `chunk_id=srcchunk_717f1a850eafa456293f3e998160668e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842385.828349` `source_timestamp=2026-01-31T06:53:05Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_976300a02790fd3b6b5c3252237280cb` `chunk_id=srcchunk_9b8d654e1b28cfda57b23af2b9699fd7` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842433.976199` `source_timestamp=2026-01-31T06:53:53Z`

## Open Questions

- What is the proper procedure to rotate Temporal certificates without causing downtime?
- Why did the cert rotation cause crash looping? Was it because the pods did not automatically pick up the new certs and required a restart?

## Sources

- `source_document_id`: `srcdoc_dd1eadae89838bea673b7011a794bcec`
- `source_revision_id`: `srcrev_00762005d6d4ff7ed686566a00c836e6`
