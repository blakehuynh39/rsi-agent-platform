---
title: "Rotating Temporal Certificates for Staging Environment"
type: "runbook"
slug: "runbooks/rotating-temporal-certificates-staging"
freshness: "2026-01-31T06:58:36Z"
tags:
  - "certificates"
  - "incident"
  - "staging"
  - "temporal"
owners: []
source_revision_ids:
  - "srcrev_002d5d0807c3e49b40c88f24d001d73e"
  - "srcrev_69acf536608d4d9030068729b87926e4"
  - "srcrev_b72afba2c732d2e1198ca28d5e4e08fd"
  - "srcrev_bfb7b465804b59bd1da96a254d58c6d2"
  - "srcrev_bff2f894b5d6b822b67c485d1e361b37"
  - "srcrev_d92d027b76a3e8d252b0a61c76337458"
  - "srcrev_d972635da0ff6141bd62828bbe7a825c"
conflict_state: "none"
---

# Rotating Temporal Certificates for Staging Environment

## Summary

Procedure and incident notes for rotating Temporal certificates used by staging workers. On 2026-01-31, after rotating expiring Temporal certificates, staging pods crash-looped and required restarting deployments to recover.

## Claims

- On 2026-01-31, all staging pods were crash looping. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bfb7b465804b59bd1da96a254d58c6d2` `chunk_id=srcchunk_67d04dede4e89406df966aaad7b93645` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841556.201809` `source_timestamp=2026-01-31T06:39:16Z`
- User U08332YRB7W rotated the Temporal certificates because they were expiring. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bff2f894b5d6b822b67c485d1e361b37` `chunk_id=srcchunk_88bdae9c12e3823413d87116d3d5f9b4` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841992.302049` `source_timestamp=2026-01-31T06:46:32Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d972635da0ff6141bd62828bbe7a825c` `chunk_id=srcchunk_170ae3977dabcd5c972b46ec96671e54` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842104.994089` `source_timestamp=2026-01-31T06:48:24Z`
- The rotation steps: generated CA cert, uploaded to Temporal, copied leaf certs to Vault for staging workers. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_002d5d0807c3e49b40c88f24d001d73e` `chunk_id=srcchunk_c82f57b42460e9df95ddc49d9cfb68fe` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842505.278199` `source_timestamp=2026-01-31T06:55:05Z`
- After the certificate rotation, staging pods crash looped, requiring a restart of all deployments to resolve. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d92d027b76a3e8d252b0a61c76337458` `chunk_id=srcchunk_3c0472c4b7a73e5a7201af82ee83386e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842716.717719` `source_timestamp=2026-01-31T06:58:36Z`
- An email was received about Cloud SQL instance SSL/TLS server certificate expiring within 90 days; it was raised as a possible related cause but not confirmed. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_b72afba2c732d2e1198ca28d5e4e08fd` `chunk_id=srcchunk_717f1a850eafa456293f3e998160668e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842385.828349` `source_timestamp=2026-01-31T06:53:05Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_69acf536608d4d9030068729b87926e4` `chunk_id=srcchunk_46e808123b9d39cc4fdd89a51dd8a525` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842394.236109` `source_timestamp=2026-01-31T06:53:14Z`

## Open Questions

- Was the Cloud SQL certificate expiry related to the incident?
- What exact secrets were updated in Vault?
- Why did the pods crash loop after certificate rotation?

## Sources

- `source_document_id`: `srcdoc_dd1eadae89838bea673b7011a794bcec`
- `source_revision_id`: `srcrev_61e8d77f284d3bb1a8ca3df91732a1df`
