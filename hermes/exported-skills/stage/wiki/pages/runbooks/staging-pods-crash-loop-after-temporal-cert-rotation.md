---
title: "Staging Pods Crash Loop After Temporal Cert Rotation"
type: "runbook"
slug: "runbooks/staging-pods-crash-loop-after-temporal-cert-rotation"
freshness: "2026-01-31T06:58:36Z"
tags:
  - "certificates"
  - "crash-loop"
  - "incident"
  - "pods"
  - "staging"
  - "temporal"
owners: []
source_revision_ids:
  - "srcrev_002d5d0807c3e49b40c88f24d001d73e"
  - "srcrev_00762005d6d4ff7ed686566a00c836e6"
  - "srcrev_61e8d77f284d3bb1a8ca3df91732a1df"
  - "srcrev_69acf536608d4d9030068729b87926e4"
  - "srcrev_b72afba2c732d2e1198ca28d5e4e08fd"
  - "srcrev_bfb7b465804b59bd1da96a254d58c6d2"
  - "srcrev_bff2f894b5d6b822b67c485d1e361b37"
  - "srcrev_c95dcaaddf2ee33494a9b886516d3264"
  - "srcrev_d92d027b76a3e8d252b0a61c76337458"
  - "srcrev_d972635da0ff6141bd62828bbe7a825c"
conflict_state: "none"
---

# Staging Pods Crash Loop After Temporal Cert Rotation

## Summary

On 2026-01-31, after rotating Temporal worker certificates, all staging pods began crash looping. The pods recovered after restarting all deployments, though the exact cause of the crash loop remains uncertain.

## Claims

- All staging pods are crash looping. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bfb7b465804b59bd1da96a254d58c6d2` `chunk_id=srcchunk_67d04dede4e89406df966aaad7b93645` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841556.201809` `source_timestamp=2026-01-31T06:39:16Z`
- The crash might be related to another issue, cc'd @U05A515NBFC. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_61e8d77f284d3bb1a8ca3df91732a1df` `chunk_id=srcchunk_6eb50480847c80f43cde69514ae2f649` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841603.127209` `source_timestamp=2026-01-31T06:40:03Z`
- @U08332YRB7W rotated the Temporal certificates. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bff2f894b5d6b822b67c485d1e361b37` `chunk_id=srcchunk_88bdae9c12e3823413d87116d3d5f9b4` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841992.302049` `source_timestamp=2026-01-31T06:46:32Z`
- All temporal worker certificates were stored in Vault and were rotated because they were expiring. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d972635da0ff6141bd62828bbe7a825c` `chunk_id=srcchunk_170ae3977dabcd5c972b46ec96671e54` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842104.994089` `source_timestamp=2026-01-31T06:48:24Z`
- Certificate rotation process: generated CA cert, uploaded cert on Temporal, copied leaf certs to Vault for respective workers, per Temporal documentation. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_002d5d0807c3e49b40c88f24d001d73e` `chunk_id=srcchunk_c82f57b42460e9df95ddc49d9cfb68fe` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842505.278199` `source_timestamp=2026-01-31T06:55:05Z`
- Secrets in Vault for staging were updated. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_00762005d6d4ff7ed686566a00c836e6` `chunk_id=srcchunk_4838d7291799f92b202cf5b60ee57971` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842594.572599` `source_timestamp=2026-01-31T06:56:34Z`
- All staging pods looked healthy after restarting all deployments. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_c95dcaaddf2ee33494a9b886516d3264` `chunk_id=srcchunk_06c82f6a8d18881f2fe4fb9696651046` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842713.942099` `source_timestamp=2026-01-31T06:58:33Z`
- The crash loop resolved after restarting all deployments, but the reason why a restart was necessary is unknown. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d92d027b76a3e8d252b0a61c76337458` `chunk_id=srcchunk_3c0472c4b7a73e5a7201af82ee83386e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842716.717719` `source_timestamp=2026-01-31T06:58:36Z`
- A Cloud SQL instance has an SSL/TLS server certificate expiring within 90 days; it might be related to the crash loop. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_b72afba2c732d2e1198ca28d5e4e08fd` `chunk_id=srcchunk_717f1a850eafa456293f3e998160668e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842385.828349` `source_timestamp=2026-01-31T06:53:05Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_69acf536608d4d9030068729b87926e4` `chunk_id=srcchunk_46e808123b9d39cc4fdd89a51dd8a525` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842394.236109` `source_timestamp=2026-01-31T06:53:14Z`

## Open Questions

- Is the Cloud SQL SSL/TLS server certificate expiry related to the crash loop?
- Why did restarting all deployments resolve the crash loop after cert rotation?

## Related Pages

- `temporal-certificate-rotation-procedure`
- `vault-secrets-management`

## Sources

- `source_document_id`: `srcdoc_dd1eadae89838bea673b7011a794bcec`
- `source_revision_id`: `srcrev_bfb7b465804b59bd1da96a254d58c6d2`
