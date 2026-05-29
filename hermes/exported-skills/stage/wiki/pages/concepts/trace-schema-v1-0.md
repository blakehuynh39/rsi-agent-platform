---
title: "Trace Schema v1.0"
type: "concept"
slug: "concepts/trace-schema-v1-0"
freshness: "2026-05-29T06:00:00Z"
tags:
  - "data-model"
  - "schema"
  - "trace"
  - "v1"
owners: []
source_revision_ids:
  - "srcrev_613dddbdd8eb04e21da6d97960bbd4f3"
conflict_state: "none"
---

# Trace Schema v1.0

## Summary

The normalized Trace Schema v1.0 that Story uses to standardize provider data, with sections for file, user, app, timestamps, attestation, and provider payload.

## Claims

- Trace Schema v1.0 standardizes fields across providers while preserving provider-specific payloads. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_613dddbdd8eb04e21da6d97960bbd4f3` `chunk_id=srcchunk_fc768c8f667a0bfbcec456d1cd35d80f` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-29T06:00:00Z`
- The schema version is "trace-v1.0". `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_613dddbdd8eb04e21da6d97960bbd4f3` `chunk_id=srcchunk_fc768c8f667a0bfbcec456d1cd35d80f` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-29T06:00:00Z`
- The schema includes top-level sections: file, file_specific (video/image/document), user, app, timestamps, attestation, and provider_payload. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_613dddbdd8eb04e21da6d97960bbd4f3` `chunk_id=srcchunk_fc768c8f667a0bfbcec456d1cd35d80f` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-29T06:00:00Z`
- The attestation section contains a payload_hash, signature, and key_id; Kled currently expects to sign the deterministic payload hash. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_613dddbdd8eb04e21da6d97960bbd4f3` `chunk_id=srcchunk_fc768c8f667a0bfbcec456d1cd35d80f` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-29T06:00:00Z`
- The file section includes content_sha256, mime_type, media_category, size_bytes, and optional hashes like phash64. `claim:claim_2_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_613dddbdd8eb04e21da6d97960bbd4f3` `chunk_id=srcchunk_fc768c8f667a0bfbcec456d1cd35d80f` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-29T06:00:00Z`
- The user section tracks source_user_id, kyc_status, tax_status, and account_verification_status. `claim:claim_2_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_613dddbdd8eb04e21da6d97960bbd4f3` `chunk_id=srcchunk_fc768c8f667a0bfbcec456d1cd35d80f` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-29T06:00:00Z`

## Open Questions

- Does Kled need direct read API access on staging, and which API key should be provisioned?
- Has payload size been tested against the 350 KiB serialized record limit?
- What are the revision semantics and monotonic seq for mutable fields like KYC, terms of service, privacy policy?
- What is the exact signature contract (bytes/hash signed, key ID format, verification method)?
- What is the finalized public stable ID format and hashing method for data_id?
- Which JSON canonicalization standard should be used for deterministic payload_hash?

## Related Pages

- `trace-backend-architecture`

## Sources

- `source_document_id`: `srcdoc_f33f716b82984e27937f90590ba0afd6`
- `source_revision_id`: `srcrev_613dddbdd8eb04e21da6d97960bbd4f3`
- `source_url`: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914)
