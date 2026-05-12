---
title: "Data Portal Backend Architecture"
type: "system"
slug: "systems/data-portal-backend-architecture"
freshness: "2026-05-12T23:27:00Z"
tags:
  - "architecture"
  - "backend"
  - "data-portal"
  - "metadata"
  - "smart-contract"
owners: []
source_revision_ids:
  - "srcrev_690ff4cad493be0ef9a141a4fe69926a"
conflict_state: "none"
---

# Data Portal Backend Architecture

## Summary

Architecture overview of the Data Portal Backend, including an embedded diagram and high-level smart contract ideas for metadata versioning.

## Claims

- The Data Portal Backend Architecture page contains an embedded architecture diagram image. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_690ff4cad493be0ef9a141a4fe69926a` `chunk_id=srcchunk_ff723e425dd5c46c74c9a8f7c98dcb61` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914` `source_timestamp=2026-05-12T23:27:00Z`
- Metadata updates should append a new immutable version/event, while the IP record stores only the current metadata pointer for efficient reads. `claim:claim_2_1` `confidence:0.90`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_690ff4cad493be0ef9a141a4fe69926a` `chunk_id=srcchunk_ff723e425dd5c46c74c9a8f7c98dcb61` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914` `source_timestamp=2026-05-12T23:27:00Z`
- The proposed data model uses a structure where data/{dataId} stores the current canonical state (currentMetadataRoot, currentSchemaId, latestSeq, latestEventHash, status) and event/{dataId}/{seq} stores immutable lifecycle events. `claim:claim_3_1` `confidence:0.90`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_690ff4cad493be0ef9a141a4fe69926a` `chunk_id=srcchunk_ff723e425dd5c46c74c9a8f7c98dcb61` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914` `source_timestamp=2026-05-12T23:27:00Z`
- It is assumed that the lifecycle events are bounded to less than 20 metadata updates during the whole life of a data IP. `claim:claim_4_1` `confidence:0.80`
  - citation: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_690ff4cad493be0ef9a141a4fe69926a` `chunk_id=srcchunk_ff723e425dd5c46c74c9a8f7c98dcb61` `native_locator=https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914` `source_timestamp=2026-05-12T23:27:00Z`

## Sources

- `source_document_id`: `srcdoc_f33f716b82984e27937f90590ba0afd6`
- `source_revision_id`: `srcrev_690ff4cad493be0ef9a141a4fe69926a`
- `source_url`: [Notion source](https://www.notion.so/Data-Portal-Backend-Architecture-35e051299a5480a3864be5b963962914)
