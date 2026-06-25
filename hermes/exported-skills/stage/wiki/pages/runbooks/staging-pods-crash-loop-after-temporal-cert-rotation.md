---
title: "Staging Pods Crash Loop After Temporal Cert Rotation"
type: "runbook"
slug: "runbooks/staging-pods-crash-loop-after-temporal-cert-rotation"
freshness: "2026-01-31T06:58:36Z"
tags:
  - "cert-rotation"
  - "incident"
  - "staging"
  - "temporal"
owners:
  - "U07TNT9N4JC"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_002d5d0807c3e49b40c88f24d001d73e"
  - "srcrev_00762005d6d4ff7ed686566a00c836e6"
  - "srcrev_3e6fed046d94ad0be05bb412803e3925"
  - "srcrev_bfb7b465804b59bd1da96a254d58c6d2"
  - "srcrev_bff2f894b5d6b822b67c485d1e361b37"
  - "srcrev_c95dcaaddf2ee33494a9b886516d3264"
  - "srcrev_d92d027b76a3e8d252b0a61c76337458"
  - "srcrev_d972635da0ff6141bd62828bbe7a825c"
conflict_state: "none"
---

# Staging Pods Crash Loop After Temporal Cert Rotation

## Summary

On 2026-01-31, all staging pods in use1-stage-story-api and use1-stage-story-api-proxy entered crash loops after temporal worker certificate rotation. The certs were rotated due to expiration, CA cert uploaded to Temporal, leaf certs copied to Vault. The issue resolved after restarting all deployments.

## Claims

- All staging pods were crash looping as reported on 2026-01-31. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bfb7b465804b59bd1da96a254d58c6d2` `chunk_id=srcchunk_67d04dede4e89406df966aaad7b93645` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841556.201809` `source_timestamp=2026-01-31T06:39:16Z`
- The crash looping affected the use1-stage-story-api and use1-stage-story-api-proxy ArgoCD applications. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bfb7b465804b59bd1da96a254d58c6d2` `chunk_id=srcchunk_67d04dede4e89406df966aaad7b93645` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841556.201809` `source_timestamp=2026-01-31T06:39:16Z`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_3e6fed046d94ad0be05bb412803e3925` `chunk_id=srcchunk_aacc19df876b40c9617e9af89ab506ba` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841587.297479` `source_timestamp=2026-01-31T06:39:47Z`
- The user U08332YRB7W rotated the temporal certificates. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_bff2f894b5d6b822b67c485d1e361b37` `chunk_id=srcchunk_88bdae9c12e3823413d87116d3d5f9b4` `native_locator=slack:C0547N89JUB:1769841556.201809:1769841992.302049` `source_timestamp=2026-01-31T06:46:32Z`
- Temporal worker certs were stored in Vault and were rotated due to expiration. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d972635da0ff6141bd62828bbe7a825c` `chunk_id=srcchunk_170ae3977dabcd5c972b46ec96671e54` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842104.994089` `source_timestamp=2026-01-31T06:48:24Z`
- The rotation process followed Temporal Cloud documentation: generated CA cert, uploaded to Temporal, and copied leaf certs to Vault for respective workers. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_002d5d0807c3e49b40c88f24d001d73e` `chunk_id=srcchunk_c82f57b42460e9df95ddc49d9cfb68fe` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842505.278199` `source_timestamp=2026-01-31T06:55:05Z`
- Secrets were updated in Vault for staging. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_00762005d6d4ff7ed686566a00c836e6` `chunk_id=srcchunk_4838d7291799f92b202cf5b60ee57971` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842594.572599` `source_timestamp=2026-01-31T06:56:34Z`
- The staging pods became healthy after restarting all deployments. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_c95dcaaddf2ee33494a9b886516d3264` `chunk_id=srcchunk_06c82f6a8d18881f2fe4fb9696651046` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842713.942099` `source_timestamp=2026-01-31T06:58:33Z`
- The exact reason for the crash looping is unknown; the person who rotated certs was investigating. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dd1eadae89838bea673b7011a794bcec` `source_revision_id=srcrev_d92d027b76a3e8d252b0a61c76337458` `chunk_id=srcchunk_3c0472c4b7a73e5a7201af82ee83386e` `native_locator=slack:C0547N89JUB:1769841556.201809:1769842716.717719` `source_timestamp=2026-01-31T06:58:36Z`

## Open Questions

- Why did the staging pods crash loop after temporal certificate rotation, even though the rotation process was followed correctly?

## Sources

- `source_document_id`: `srcdoc_dd1eadae89838bea673b7011a794bcec`
- `source_revision_id`: `srcrev_d92d027b76a3e8d252b0a61c76337458`
