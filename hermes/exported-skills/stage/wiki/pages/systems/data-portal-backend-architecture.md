---
title: "Data Portal Backend Architecture"
type: "system"
slug: "systems/data-portal-backend-architecture"
freshness: "2026-05-18T21:10:00Z"
tags:
  - "architecture"
  - "audit-api"
  - "backend"
  - "data-portal"
owners: []
source_revision_ids:
  - "srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4"
  - "srcrev_690ff4cad493be0ef9a141a4fe69926a"
conflict_state: "none"
---

# Data Portal Backend Architecture

## Summary

Architecture overview of the Data Portal Backend, including an embedded diagram, high-level smart contract ideas for metadata versioning, and a detailed Data Audit API specification for durable, provider-scoped audit record submission.

## Claims

- The Data Portal Backend Architecture page contains an embedded architecture diagram image. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_690ff4cad493be0ef9a141a4fe69926a` `chunk_id=srcchunk_ff723e425dd5c46c74c9a8f7c98dcb61` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914` `source_timestamp=2026-05-12T23:27:00Z`
- Metadata updates should append a new immutable version/event, while the IP record stores only the current metadata pointer for efficient reads. `claim:claim_2_1` `confidence:0.90`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_690ff4cad493be0ef9a141a4fe69926a` `chunk_id=srcchunk_ff723e425dd5c46c74c9a8f7c98dcb61` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914` `source_timestamp=2026-05-12T23:27:00Z`
- The proposed data model uses a structure where data/{dataId} stores the current canonical state (currentMetadataRoot, currentSchemaId, latestSeq, latestEventHash, status) and event/{dataId}/{seq} stores immutable lifecycle events. `claim:claim_3_1` `confidence:0.90`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_690ff4cad493be0ef9a141a4fe69926a` `chunk_id=srcchunk_ff723e425dd5c46c74c9a8f7c98dcb61` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914` `source_timestamp=2026-05-12T23:27:00Z`
- It is assumed that the lifecycle events are bounded to less than 20 metadata updates during the whole life of a data IP. `claim:claim_4_1` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_690ff4cad493be0ef9a141a4fe69926a` `chunk_id=srcchunk_ff723e425dd5c46c74c9a8f7c98dcb61` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914` `source_timestamp=2026-05-12T23:27:00Z`
- The Data Audit API allows external providers to submit durable audit records for initial data ID registration and metadata updates for an existing data ID. `claim:claim_5_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_6603da74fd92c49c8a7fd857be543144` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-18T21:10:00Z`
- The API is provider-scoped, and a data ID is identified by the combination of provider and data_id, allowing different providers to submit records without colliding. `claim:claim_6_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_6603da74fd92c49c8a7fd857be543144` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-18T21:10:00Z`
- All write endpoints require the headers X-API-Key, X-Provider, X-Batch-Id, and Content-Type (application/json for standard batch requests). `claim:claim_7_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_6603da74fd92c49c8a7fd857be543144` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-18T21:10:00Z`
- Provider names are normalized to lowercase and may contain only lowercase letters, numbers, hyphens, and underscores. `claim:claim_8_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_6603da74fd92c49c8a7fd857be543144` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-18T21:10:00Z`
- The data_id must be a UUID. `claim:claim_9_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_6603da74fd92c49c8a7fd857be543144` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-18T21:10:00Z`
- The maximum request body size is 25 MiB, the maximum metadata updates per data ID is 20, and the maximum serialized record size is 350 KiB. `claim:claim_10_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_57f9fff42d0a04c3917b6d667e3b8baa` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-18T21:10:00Z`
- Delivery is at least once, so duplicate submissions may occur. Duplicate submissions of the same event are treated idempotently. `claim:claim_11_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_57f9fff42d0a04c3917b6d667e3b8baa` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-18T21:10:00Z`
- Same provider, same data_id, same metadata sequence, and different event content is treated as a conflict. `claim:claim_12_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_57f9fff42d0a04c3917b6d667e3b8baa` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-18T21:10:00Z`
- Metadata updates may arrive before the initial data ID registration. `claim:claim_13_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_57f9fff42d0a04c3917b6d667e3b8baa` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-18T21:10:00Z`
- Audit data is durable and does not expire. `claim:claim_14_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4` `chunk_id=srcchunk_57f9fff42d0a04c3917b6d667e3b8baa` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-2` `source_timestamp=2026-05-18T21:10:00Z`

## Sources

- `source_document_id`: `srcdoc_f33f716b82984e27937f90590ba0afd6`
- `source_revision_id`: `srcrev_28dbdbbb9ea2bd6313604e3dbe7774a4`
- `source_url`: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914)
