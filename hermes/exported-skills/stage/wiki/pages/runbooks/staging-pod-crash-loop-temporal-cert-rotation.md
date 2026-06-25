---
title: "Staging Pod Crash Loop After Temporal Certificate Rotation"
type: "runbook"
slug: "runbooks/staging-pod-crash-loop-temporal-cert-rotation"
freshness: "2026-01-31T06:58:36Z"
tags:
  - "certificate-rotation"
  - "crash-loop"
  - "incident"
  - "staging"
  - "temporal"
  - "vault"
owners:
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_002d5d0807c3e49b40c88f24d001d73e"
  - "srcrev_00762005d6d4ff7ed686566a00c836e6"
  - "srcrev_b72afba2c732d2e1198ca28d5e4e08fd"
  - "srcrev_bfb7b465804b59bd1da96a254d58c6d2"
  - "srcrev_bff2f894b5d6b822b67c485d1e361b37"
  - "srcrev_c95dcaaddf2ee33494a9b886516d3264"
  - "srcrev_d3be9559dda2ed3f3922d3b44769cdf6"
  - "srcrev_d92d027b76a3e8d252b0a61c76337458"
  - "srcrev_d972635da0ff6141bd62828bbe7a825c"
conflict_state: "none"
---

# Staging Pod Crash Loop After Temporal Certificate Rotation

## Summary

On 2026-01-31, staging pods entered a crash loop after temporal certificates were rotated. The issue was resolved by restarting all deployments. Root cause is under investigation.

## Claims

- All staging pods are crash looping. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bfb7b465804b59bd1da96a254d58c6d2` `chunk_id=srcchunk_67d04dede4e89406df966aaad7b93645` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841556.201809` `source_timestamp=2026-01-31T06:39:16Z`
- @U08332YRB7W rotated the temporal certs. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bff2f894b5d6b822b67c485d1e361b37` `chunk_id=srcchunk_88bdae9c12e3823413d87116d3d5f9b4` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841992.302049` `source_timestamp=2026-01-31T06:46:32Z`
- Temporal worker certs were stored in Vault and were rotated because they were expiring. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d972635da0ff6141bd62828bbe7a825c` `chunk_id=srcchunk_170ae3977dabcd5c972b46ec96671e54` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842104.994089` `source_timestamp=2026-01-31T06:48:24Z`
- Circulation supply data in Vault was not modified during the certificate rotation. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d3be9559dda2ed3f3922d3b44769cdf6` `chunk_id=srcchunk_9ce4038e644152df6e0b2968cfff76a2` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842301.623829` `source_timestamp=2026-01-31T06:51:41Z`
- A Cloud SQL instance SSL/TLS server certificate will expire within 90 days. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_b72afba2c732d2e1198ca28d5e4e08fd` `chunk_id=srcchunk_717f1a850eafa456293f3e998160668e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842385.828349` `source_timestamp=2026-01-31T06:53:05Z`
- Certificate rotation steps: generated CA cert, uploaded cert to Temporal, copied leaf certs to Vault for respective workers. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_002d5d0807c3e49b40c88f24d001d73e` `chunk_id=srcchunk_c82f57b42460e9df95ddc49d9cfb68fe` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842505.278199` `source_timestamp=2026-01-31T06:55:05Z`
- Vault secrets for staging were updated during the rotation. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_00762005d6d4ff7ed686566a00c836e6` `chunk_id=srcchunk_4838d7291799f92b202cf5b60ee57971` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842594.572599` `source_timestamp=2026-01-31T06:56:34Z`
- The staging environment recovered and looked healthy after restarting all deployments. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_c95dcaaddf2ee33494a9b886516d3264` `chunk_id=srcchunk_06c82f6a8d18881f2fe4fb9696651046` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842713.942099` `source_timestamp=2026-01-31T06:58:33Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d92d027b76a3e8d252b0a61c76337458` `chunk_id=srcchunk_3c0472c4b7a73e5a7201af82ee83386e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842716.717719` `source_timestamp=2026-01-31T06:58:36Z`

## Open Questions

- Is there a dependency on certificate validity that caused the crash loops?
- What is the connection between the temporal cert rotation and the staging pods crash looping?
- Why did restarting deployments resolve the issue?

## Sources

- `source_document_id`: `srcdoc_dd1eadae89838bea673b7011a794bcec`
- `source_revision_id`: `srcrev_735fdafb6dfd915eae68a1c13c547ef6`
