---
title: "Kled Trace API"
type: "system"
slug: "systems/kled-trace-api"
freshness: "2026-05-26T22:27:00Z"
tags:
  - "api"
  - "kled"
  - "schema"
  - "staging"
  - "trace"
  - "webhook"
owners: []
source_revision_ids:
  - "srcrev_dbcb471c659c564a8ce104cd6ca05a4a"
conflict_state: "none"
---

# Kled Trace API

## Summary

Staging V1 of the Kled Trace API is deployed and load-tested as of May 21, 2026. It standardizes provider payloads into the Trace Schema v1.0 and provides webhook batch endpoints for writing, and read/search APIs for the Trace frontend. The staging architecture uses story-api webhook, SQS, and data-audit-ingestor for durable ingestion.

## Claims

- Staging V1 of the Kled Trace API is deployed and load-tested as of May 21, 2026. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_3137401b4ea5357cc0ae175c4548442a` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1` `source_timestamp=2026-05-26T22:27:00Z`
- The staging architecture routes provider clients (Kled/Otto) via Cloudflare to the story-api webhook batch endpoint; accepted chunks are enqueued to SQS Standard; the data-audit-ingestor persists durable audit rows and explicit index rows; story-api serves read/search APIs for the Trace frontend. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_3137401b4ea5357cc0ae175c4548442a` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1` `source_timestamp=2026-05-26T22:27:00Z`
- Returns HTTP 202 Accepted only after request validation and SQS acceptance; durable persistence happens asynchronously. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_3137401b4ea5357cc0ae175c4548442a` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1` `source_timestamp=2026-05-26T22:27:00Z`
- SQS Standard queue provides at-least-once delivery; duplicate delivery is expected and handled idempotently. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_3137401b4ea5357cc0ae175c4548442a` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1` `source_timestamp=2026-05-26T22:27:00Z`
- The ingestor removes messages from SQS only after they are persisted, idempotently skipped, or explicitly rejected. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_3137401b4ea5357cc0ae175c4548442a` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1` `source_timestamp=2026-05-26T22:27:00Z`
- A DLQ exists for conflict and poison messages; automatic replay is not configured. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_3137401b4ea5357cc0ae175c4548442a` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1` `source_timestamp=2026-05-26T22:27:00Z`
- Kled clients should retry transient errors (502, 503, 504, 429, network failures) using the same request body and X-Batch-Id header. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_3137401b4ea5357cc0ae175c4548442a` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-1` `source_timestamp=2026-05-26T22:27:00Z`
- Read API endpoints include: GET /api/v1/data-audit/{data_id}/metadatas, GET /api/v1/data-audit/{data_id}/metadatas?provider=kled, and various search endpoints filtering by field (e.g., file.media_category, provider, source_record_id, file.content_sha256, file.hashes.phash64). `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-2) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_7f307fb7a171de7c177b1e1d391a7441` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-2` `source_timestamp=2026-05-26T22:27:00Z`
- Trace Schema v1.0 standardizes payload fields: file (content_sha256, mime_type, media_category, size_bytes, hashes), file_specific (video, image, document sub-objects), user (source_user_id, kyc_status, tax_status, account_verification_status), app (platform_name, terms_of_service, privacy_policy), timestamps (occurred_at, uploaded_at, captured_at), attestation (payload_hash, signature, key_id), and provider_payload (full original provider public payload). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-2) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_7f307fb7a171de7c177b1e1d391a7441` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-2` `source_timestamp=2026-05-26T22:27:00Z`
- The attestation section includes a payload_hash, signature, and key_id; Kled is expected to sign the deterministic payload hash. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-2) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_7f307fb7a171de7c177b1e1d391a7441` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-2` `source_timestamp=2026-05-26T22:27:00Z`
- Explicit schema versioning is maintained in the audit path via the schema_version field in the payload. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-2) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_7f307fb7a171de7c177b1e1d391a7441` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-2` `source_timestamp=2026-05-26T22:27:00Z`
- Future features not yet implemented: automatic DLQ replay tooling, on-chain anchoring, and analytics-style ad hoc search indexing. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-3) `source_document_id=srcdoc_6904027f6ebaceb800260ddfeb5c2c2c` `source_revision_id=srcrev_dbcb471c659c564a8ce104cd6ca05a4a` `chunk_id=srcchunk_983e34215f87033902157cf03c93196a` `native_locator=https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857#chunk-3` `source_timestamp=2026-05-26T22:27:00Z`

## Open Questions

- Confirm JSON canonicalization standard for deterministic payload_hash.
- Confirm payload size tests against the 350 KiB serialized record limit.
- Confirm revision semantics and monotonic seq for mutable fields such as KYC, terms of service, and privacy policy state.
- Confirm signature contract: exact bytes/hash signed, key ID format, and verification method.
- Confirm whether Kled needs direct read API access on staging, and which read-capable API key should be provisioned.
- Finalize public stable ID format and provide exact sample values. Current meeting answer points to 16 hex characters, but hashing method remains to be finalized.

## Sources

- `source_document_id`: `srcdoc_6904027f6ebaceb800260ddfeb5c2c2c`
- `source_revision_id`: `srcrev_dbcb471c659c564a8ce104cd6ca05a4a`
- `source_url`: [Notion source](https://www.notion.so/Kled-Trace-API-36c051299a5480b7b6c4e7609b6d4857)
