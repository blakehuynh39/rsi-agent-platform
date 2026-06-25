---
title: "Staging Pod CrashLoop After Temporal Cert Rotation"
type: "runbook"
slug: "runbooks/staging-pod-crashloop-after-temporal-cert-rotation"
freshness: "2026-01-31T06:58:36Z"
tags:
  - "certificates"
  - "incident"
  - "staging"
  - "temporal"
  - "vault"
owners:
  - "U07TNT9N4JC"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_002d5d0807c3e49b40c88f24d001d73e"
  - "srcrev_00762005d6d4ff7ed686566a00c836e6"
  - "srcrev_b72afba2c732d2e1198ca28d5e4e08fd"
  - "srcrev_bfb7b465804b59bd1da96a254d58c6d2"
  - "srcrev_bff2f894b5d6b822b67c485d1e361b37"
  - "srcrev_c95dcaaddf2ee33494a9b886516d3264"
  - "srcrev_d92d027b76a3e8d252b0a61c76337458"
  - "srcrev_d972635da0ff6141bd62828bbe7a825c"
conflict_state: "none"
---

# Staging Pod CrashLoop After Temporal Cert Rotation

## Summary

On 2026-01-31, staging pods started crash looping after a developer rotated Temporal worker certificates in Vault. Restarting all deployments restored health. Root cause remains under investigation.

## Claims

- Staging pods were crash looping, as seen in ArgoCD for use1-stage-story-api and use1-stage-story-api-proxy. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bfb7b465804b59bd1da96a254d58c6d2` `chunk_id=srcchunk_67d04dede4e89406df966aaad7b93645` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841556.201809` `source_timestamp=2026-01-31T06:39:16Z`
- User U08332YRB7W rotated temporal worker certificates stored in Vault because they were expiring. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bff2f894b5d6b822b67c485d1e361b37` `chunk_id=srcchunk_88bdae9c12e3823413d87116d3d5f9b4` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841992.302049` `source_timestamp=2026-01-31T06:46:32Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d972635da0ff6141bd62828bbe7a825c` `chunk_id=srcchunk_170ae3977dabcd5c972b46ec96671e54` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842104.994089` `source_timestamp=2026-01-31T06:48:24Z`
- The cert rotation followed the Temporal documentation: a CA cert was generated and uploaded to Temporal, and leaf certs were copied to Vault for respective workers. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_002d5d0807c3e49b40c88f24d001d73e` `chunk_id=srcchunk_c82f57b42460e9df95ddc49d9cfb68fe` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842505.278199` `source_timestamp=2026-01-31T06:55:05Z`
- After the rotation, secrets in Vault for staging were updated. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_00762005d6d4ff7ed686566a00c836e6` `chunk_id=srcchunk_4838d7291799f92b202cf5b60ee57971` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842594.572599` `source_timestamp=2026-01-31T06:56:34Z`
- A Cloud SQL instance had an SSL/TLS server certificate that will expire within 90 days, which may be unrelated to the crash loop. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_b72afba2c732d2e1198ca28d5e4e08fd` `chunk_id=srcchunk_717f1a850eafa456293f3e998160668e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842385.828349` `source_timestamp=2026-01-31T06:53:05Z`
- After restarting all deployments, staging pods became healthy again. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_c95dcaaddf2ee33494a9b886516d3264` `chunk_id=srcchunk_06c82f6a8d18881f2fe4fb9696651046` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842713.942099` `source_timestamp=2026-01-31T06:58:33Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d92d027b76a3e8d252b0a61c76337458` `chunk_id=srcchunk_3c0472c4b7a73e5a7201af82ee83386e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842716.717719` `source_timestamp=2026-01-31T06:58:36Z`

## Open Questions

- Was the cert rotation applied incorrectly or did pods fail to pick up new certs without a restart?
- What exactly caused the crash loop after the cert rotation?
- Why did restarting all deployments resolve the issue?

## Sources

- `source_document_id`: `srcdoc_dd1eadae89838bea673b7011a794bcec`
- `source_revision_id`: `srcrev_69acf536608d4d9030068729b87926e4`
