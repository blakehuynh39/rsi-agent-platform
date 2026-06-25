---
title: "Staging Pods CrashLoop After Temporal Certs Rotation"
type: "runbook"
slug: "runbooks/staging-pods-crashloop-after-temporal-certs-rotation"
freshness: "2026-01-31T06:58:36Z"
tags:
  - "certificates"
  - "crashloop"
  - "staging"
  - "temporal"
  - "vault"
owners:
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_002d5d0807c3e49b40c88f24d001d73e"
  - "srcrev_976300a02790fd3b6b5c3252237280cb"
  - "srcrev_b72afba2c732d2e1198ca28d5e4e08fd"
  - "srcrev_bfb7b465804b59bd1da96a254d58c6d2"
  - "srcrev_bff2f894b5d6b822b67c485d1e361b37"
  - "srcrev_c95dcaaddf2ee33494a9b886516d3264"
  - "srcrev_d92d027b76a3e8d252b0a61c76337458"
conflict_state: "none"
---

# Staging Pods CrashLoop After Temporal Certs Rotation

## Summary

Runbook for diagnosing and resolving staging pod crash loops that occurred after rotating Temporal worker certificates stored in Vault.

## Claims

- All staging pods were crash looping. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bfb7b465804b59bd1da96a254d58c6d2` `chunk_id=srcchunk_67d04dede4e89406df966aaad7b93645` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841556.201809` `source_timestamp=2026-01-31T06:39:16Z`
- Temporal certificates were rotated. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bff2f894b5d6b822b67c485d1e361b37` `chunk_id=srcchunk_88bdae9c12e3823413d87116d3d5f9b4` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841992.302049` `source_timestamp=2026-01-31T06:46:32Z`
- Certificate rotation process involved generating a CA certificate, uploading it to Temporal, and copying leaf certificates to Vault for respective workers. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_002d5d0807c3e49b40c88f24d001d73e` `chunk_id=srcchunk_c82f57b42460e9df95ddc49d9cfb68fe` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842505.278199` `source_timestamp=2026-01-31T06:55:05Z`
- A Cloud SQL instance SSL/TLS certificate expiration warning email was received, but it was within 90 days and not immediately related. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_b72afba2c732d2e1198ca28d5e4e08fd` `chunk_id=srcchunk_717f1a850eafa456293f3e998160668e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842385.828349` `source_timestamp=2026-01-31T06:53:05Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_976300a02790fd3b6b5c3252237280cb` `chunk_id=srcchunk_9b8d654e1b28cfda57b23af2b9699fd7` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842433.976199` `source_timestamp=2026-01-31T06:53:53Z`
- After restarting all deployments, the staging pods became healthy. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_c95dcaaddf2ee33494a9b886516d3264` `chunk_id=srcchunk_06c82f6a8d18881f2fe4fb9696651046` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842713.942099` `source_timestamp=2026-01-31T06:58:33Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d92d027b76a3e8d252b0a61c76337458` `chunk_id=srcchunk_3c0472c4b7a73e5a7201af82ee83386e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842716.717719` `source_timestamp=2026-01-31T06:58:36Z`

## Open Questions

- Why did the pods require a restart after certificate rotation? Investigation pending.

## Sources

- `source_document_id`: `srcdoc_dd1eadae89838bea673b7011a794bcec`
- `source_revision_id`: `srcrev_976300a02790fd3b6b5c3252237280cb`
