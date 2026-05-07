---
title: "Poseidon Storage Node Responsibilities"
type: "system"
slug: "systems/poseidon-storage-node-responsibilities"
freshness: "2025-08-11T19:42:00Z"
tags:
  - "architecture"
  - "poseidon"
  - "responsibilities"
  - "storage-node"
owners: []
source_revision_ids:
  - "srcrev_14188fbf16ec0dd8f62194dbaf4015c6"
conflict_state: "none"
---

# Poseidon Storage Node Responsibilities

## Summary

Responsibilities of system components relevant to the Poseidon storage nodes, including Trident Platform (S3 API, KMS, Storage Client) and SN Database.

## Claims

- The Trident Platform S3 API accepts file uploads from users in an S3-compatible API, must support existing S3 SDKs including the official AWS S3 SDK, must authenticate users' uploads using headers, stores file uploads into internal AWS S3 buckets for future async operations, and removes files cached in S3 once async operations are complete. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-1) `source_document_id=srcdoc_984f4da8786cfde8c37f7411ac02a849` `source_revision_id=srcrev_14188fbf16ec0dd8f62194dbaf4015c6` `chunk_id=srcchunk_483ab19d01266db539dc03528cd20478` `native_locator=https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- The Trident Platform KMS creates an encryption key per bucket when buckets are created (for all buckets), stores a mapping of bucket ⇒ SSS key ⇒ storage node (since each storage node stores one SSS key of a bucket), and distributes & fetches SSS keys from each storage node. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-1) `source_document_id=srcdoc_984f4da8786cfde8c37f7411ac02a849` `source_revision_id=srcrev_14188fbf16ec0dd8f62194dbaf4015c6` `chunk_id=srcchunk_483ab19d01266db539dc03528cd20478` `native_locator=https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- The Trident Platform Storage Client runs durable workers that fetch file uploads from internal AWS S3 buckets (used by S3 API), encrypts data using the encryption key corresponding to a bucket of cached file uploads fetched from KMS, calculates the Merkle root of cached file uploads (which takes a long time for large files), and validates that the Merkle root of a cached file matches data object IDs. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-1) `source_document_id=srcdoc_984f4da8786cfde8c37f7411ac02a849` `source_revision_id=srcrev_14188fbf16ec0dd8f62194dbaf4015c6` `chunk_id=srcchunk_483ab19d01266db539dc03528cd20478` `native_locator=https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- The user uploads fragmented keys to the respective storage nodes, and each SN will only accept if the hash of the received fragmented key matches any hashes stored in the hashed key database. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-2) `source_document_id=srcdoc_984f4da8786cfde8c37f7411ac02a849` `source_revision_id=srcrev_14188fbf16ec0dd8f62194dbaf4015c6` `chunk_id=srcchunk_8a39795318f3d59787e440bad521253f` `native_locator=https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-2` `source_timestamp=2025-08-11T19:42:00Z`
- An SN client must store buyer addresses for access to decryption keys after observing events on L1 KMS contracts. When a buyer purchases a data license, KMS contracts will emit an event containing the buyer’s address and data object IDs (or storage ID). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-2) `source_document_id=srcdoc_984f4da8786cfde8c37f7411ac02a849` `source_revision_id=srcrev_14188fbf16ec0dd8f62194dbaf4015c6` `chunk_id=srcchunk_8a39795318f3d59787e440bad521253f` `native_locator=https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-2` `source_timestamp=2025-08-11T19:42:00Z`
- A database driver will be created to enable SN operators to use any databases. Initially, Pebble DB will be supported. MinIO was previously considered but will not be used due to strong trust assumptions; instead, isolated MinIO nodes may be used as storage nodes with custom erasure-coding at the storage client level, making raw databases more efficient. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-2) `source_document_id=srcdoc_984f4da8786cfde8c37f7411ac02a849` `source_revision_id=srcrev_14188fbf16ec0dd8f62194dbaf4015c6` `chunk_id=srcchunk_8a39795318f3d59787e440bad521253f` `native_locator=https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662#chunk-2` `source_timestamp=2025-08-11T19:42:00Z`

## Sources

- `source_document_id`: `srcdoc_984f4da8786cfde8c37f7411ac02a849`
- `source_revision_id`: `srcrev_14188fbf16ec0dd8f62194dbaf4015c6`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-Core-PRD-1eb051299a548055be62c13f164c9662)
