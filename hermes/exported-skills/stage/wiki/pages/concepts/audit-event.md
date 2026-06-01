---
title: "Audit Event"
type: "concept"
slug: "concepts/audit-event"
freshness: "2026-06-01T22:17:00Z"
tags:
  - "audit"
  - "event"
  - "hash"
  - "on-chain"
owners: []
source_revision_ids:
  - "srcrev_06ab24663fb480b2559cdd1aa60f66a5"
conflict_state: "none"
---

# Audit Event

## Summary

Structure, identity, and hash computation of an on-chain audit event.

## Claims

- Every audit event is identified on-chain by a 32-byte content hash; identical committed content yields the same hash, differing content yields different hashes. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- An event is either the initial registration of a data piece or a subsequent metadata update. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- Identity fields include data_id, provider, seq, event_type, event_hash, event_hash_version, event_hash_canonicalization, trace_schema_version, and occurred_at. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- event_hash is the canonical SHA-256 over the full event payload, serving as unique fingerprint and basis for on-chain registration hash. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- event_hash_version and event_hash_canonicalization allow evolving the hash definition without ambiguity. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- Payload fields: metadata_json, metadata_root, prev_metadata_root (updates only), initial_metadata_root (registration only), initial_metadata_json (registration only), source_record_id. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- Bookkeeping fields include batch_id, request_id, and tx_hash (empty until on-chain transaction). `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- For DataRegistered events (seq 0), the event_hash preimage commits to version labels, event_type='DataRegistered', provider, data_id, source_record_id (if present), initial_metadata_root, initial_metadata_json, and occurred_at. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- For MetadataUpdated events (seq ≥ 1), the event_hash preimage commits to version labels and a different set of fields (details truncated in source). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_ddd90d36514a936002f8f4c44d50f684` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-1` `source_timestamp=2026-06-01T22:17:00Z`
- The multihash prefix is fixed at 0x1220 (sha2-256) per contract; a different algorithm requires a new contract. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-2) `source_document_id=srcdoc_746fbb670e36fdd39d970096271294fd` `source_revision_id=srcrev_06ab24663fb480b2559cdd1aa60f66a5` `chunk_id=srcchunk_cc4605a7d89cc620ebef643cca70aaff` `native_locator=https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244#chunk-2` `source_timestamp=2026-06-01T22:17:00Z`

## Sources

- `source_document_id`: `srcdoc_746fbb670e36fdd39d970096271294fd`
- `source_revision_id`: `srcrev_06ab24663fb480b2559cdd1aa60f66a5`
- `source_url`: [Notion source](https://www.notion.so/Contract-Design-Proposal-372051299a54808da119e78b1759d244)
