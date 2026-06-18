---
title: "TaxBandits Live Account Questionnaire"
type: "decision"
slug: "decisions/taxbandits-questionnaire-activation"
freshness: "2026-06-04T18:29:11Z"
tags:
  - "compliance"
  - "data-retention"
  - "integration"
  - "taxbandits"
owners:
  - "Engineering (@U05A515NBFC)"
  - "Legal (@U086FECSTP1)"
  - "Product (@U06A5AQ1VD3)"
  - "Project Lead (@U04L0DD6B6F)"
  - "Security (@U08332YRB7W)"
source_revision_ids:
  - "srcrev_08b3ada3e37675f26fbed7d1395637f8"
  - "srcrev_0c015264a1751cf1bc0d0460400ebe8b"
  - "srcrev_0ddbefb317fa719928f5b75ad29b4aec"
  - "srcrev_0fd1bad1e0cda1fef85f2e98fbff8ad4"
  - "srcrev_1da64b2b91065d981f41e8d7165e220f"
  - "srcrev_3f66e8b0fa2df527fe74a85d67a554fa"
  - "srcrev_41ac32e51b59d47cdc384d3f194994e0"
  - "srcrev_5140d316f9d7a51297d4da3bf8d310ea"
  - "srcrev_5b2df2e73228fd06c42ce57bd78a4d37"
  - "srcrev_7f06bf1b25f8b1db010bffef525abb96"
  - "srcrev_8725d09ac72d5e1eac2576a6e58a1035"
  - "srcrev_b399ce3bd34ba8f431b83462260da20d"
  - "srcrev_fc947e89b45417bc001f1d4ce26dc0dd"
conflict_state: "none"
---

# TaxBandits Live Account Questionnaire

## Summary

Completion of TaxBandits questionnaire required to activate live payments. Key decision: we will not store signed W-9/W-8 PDFs locally; TaxBandits retains them for 10 years. Domain whitelisting set for api.numolabs.ai. Reviews from Engineering, Security, Legal, Product are in progress.

## Claims

- TaxBandits retains submitted W-9/W-8BEN forms for up to 10 years, consistent with IRS recordkeeping rules, and retrieval is guaranteed even after account termination or downgrade. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_3f66e8b0fa2df527fe74a85d67a554fa` `chunk_id=srcchunk_6cd0eeceb5645414c35404cefdf7e0a8` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780588733.213589` `source_timestamp=2026-06-04T15:59:12Z`
- RSI will not store signed PDFs of W-9/W-8 forms locally; instead relies on TaxBandits' retention and on-demand retrieval. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_08b3ada3e37675f26fbed7d1395637f8` `chunk_id=srcchunk_b834549607961031d7afe9556b4e75b3` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780590499.495579` `source_timestamp=2026-06-04T16:28:19Z`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_0ddbefb317fa719928f5b75ad29b4aec` `chunk_id=srcchunk_a0723b6e87807d3a80d3d283cfa66648` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780588864.879859` `source_timestamp=2026-06-04T16:01:04Z`
- Domain whitelisting uses: https://staging-api.numolabs.ai and https://api.numolabs.ai. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_1da64b2b91065d981f41e8d7165e220f` `chunk_id=srcchunk_5a0b985e9e2060ed8afe895a9f539d7c` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780597606.493749` `source_timestamp=2026-06-04T18:26:46Z`
- Backup and restore testing: Postgres backups cover tax form metadata, redacted snapshots, and webhook/audit tables; PDF restore testing not applicable because PDFs are not stored locally. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_41ac32e51b59d47cdc384d3f194994e0` `chunk_id=srcchunk_419656d85e6c8c25a71fd6aa7597da33` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780597625.293579` `source_timestamp=2026-06-04T18:27:05Z`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_7f06bf1b25f8b1db010bffef525abb96` `chunk_id=srcchunk_ce7eb992f79c7b9b035495c720a79998` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780597751.196419` `source_timestamp=2026-06-04T18:29:11Z`
- Cyber insurance coverage is confirmed. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_5b2df2e73228fd06c42ce57bd78a4d37` `chunk_id=srcchunk_1e7f5d7dd4ebd05acc5f1b920b1e2da1` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780534847.413429` `source_timestamp=2026-06-04T01:00:47Z`
- An email address for 'live account' notifications is requested, pending assignment. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_5140d316f9d7a51297d4da3bf8d310ea` `chunk_id=srcchunk_5db396dbb3c7cee126db67f2b7f6fbb8` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780534105.129679` `source_timestamp=2026-06-04T00:48:25Z`
- Engineering review covers questionnaire sections: §1 Use Case, §2 Testing, §3 Error Handling, technical parts of §5 Security (encryption, secrets, webhook, backups, logs), and IP-vs-Domain whitelisting choice in §6. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_fc947e89b45417bc001f1d4ce26dc0dd` `chunk_id=srcchunk_312004bff03d9ee7fb85d990275357c5` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780533660.991579` `source_timestamp=2026-06-04T00:41:24Z`
- Security review covers: breach history, vuln/pen testing, PII & PDF storage, MFA, incident handling, SOC 2. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_fc947e89b45417bc001f1d4ce26dc0dd` `chunk_id=srcchunk_312004bff03d9ee7fb85d990275357c5` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780533660.991579` `source_timestamp=2026-06-04T00:41:24Z`
- Legal review covers: compliance (lawsuits, CCPA, cross-border transfers, retention, sub-processors, audit rights) and US-source/withholding questions. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_fc947e89b45417bc001f1d4ce26dc0dd` `chunk_id=srcchunk_312004bff03d9ee7fb85d990275357c5` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780533660.991579` `source_timestamp=2026-06-04T00:41:24Z`
- Product review covers use-case framing and app flow screenshots/video (1.2). `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_fc947e89b45417bc001f1d4ce26dc0dd` `chunk_id=srcchunk_312004bff03d9ee7fb85d990275357c5` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780533660.991579` `source_timestamp=2026-06-04T00:41:24Z`
- TaxBandits integration provides a drop-in form or link that allows users to download the filled PDF after submission. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_b399ce3bd34ba8f431b83462260da20d` `chunk_id=srcchunk_723c56005e17b8ad48a9af6480ffbeeb` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780576790.624019` `source_timestamp=2026-06-04T12:39:50Z`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_0fd1bad1e0cda1fef85f2e98fbff8ad4` `chunk_id=srcchunk_c6ab20e4cc98f7082348425f604b5ecb` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780577088.807969` `source_timestamp=2026-06-04T12:44:48Z`
- Minimized Postgres records store status, treaty/country fields, redacted JSON, S3 keys, but no raw PII (TIN/DOB/name). `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_8725d09ac72d5e1eac2576a6e58a1035` `chunk_id=srcchunk_717360f5bda740fe34700c7db61ac909` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780571653.201329` `source_timestamp=2026-06-04T11:14:46Z`
- Engineering raised concern about storing full signed PDFs, citing increased PII surface, CCPA/SOC2 impact, and questioning necessity given TaxBandits' system-of-record role. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_81a7bff9cb47a5d52a8c2656dd59923d` `source_revision_id=srcrev_0c015264a1751cf1bc0d0460400ebe8b` `chunk_id=srcchunk_101477e55102f7bafe6a6ab546b1fa71` `native_locator=slack:C0AL7EKNHDF:1780533660.991579:1780574454.312759` `source_timestamp=2026-06-04T12:00:54Z`

## Open Questions

- Does vendor-held-and-retrievable (TaxBandits) satisfy withholding agent recordkeeping requirements, or must we retain our own copies? (Legal sign-off pending)
- Product review of use-case framing and screenshots/video is outstanding.
- Security and Legal reviews are still outstanding; their feedback is awaited.
- What email address should be provided for live account notifications?

## Sources

- `source_document_id`: `srcdoc_81a7bff9cb47a5d52a8c2656dd59923d`
- `source_revision_id`: `srcrev_e27f3c5d22b37613ea26d7dc8dfc934b`
