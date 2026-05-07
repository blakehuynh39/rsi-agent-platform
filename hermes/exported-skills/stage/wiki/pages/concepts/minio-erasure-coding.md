---
title: "MinIO Erasure Coding"
type: "concept"
slug: "concepts/minio-erasure-coding"
freshness: "2025-05-09T05:57:00Z"
tags:
  - "erasure-coding"
  - "minio"
owners: []
source_revision_ids:
  - "srcrev_5e4d1c6bec79a677faabbcd501afa9f6"
conflict_state: "none"
---

# MinIO Erasure Coding

## Summary

MinIO's erasure coding implementation uses per-object inline Reed-Solomon coding with configurable redundancy, object-level healing, and high fault tolerance.

## Claims

- MinIO uses per-object inline Reed-Solomon erasure coding with configurable redundancy levels. `claim:claim_minio_ec_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1) `source_document_id=srcdoc_fda12217a5552b410689a3227d6f89c3` `source_revision_id=srcrev_5e4d1c6bec79a677faabbcd501afa9f6` `chunk_id=srcchunk_b45688245d70d499139ad7d596d2f3ec` `native_locator=https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1` `source_timestamp=2025-05-09T05:57:00Z`
- Healing is performed at the object level and can heal multiple objects independently. `claim:claim_minio_ec_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1) `source_document_id=srcdoc_fda12217a5552b410689a3227d6f89c3` `source_revision_id=srcrev_5e4d1c6bec79a677faabbcd501afa9f6` `chunk_id=srcchunk_b45688245d70d499139ad7d596d2f3ec` `native_locator=https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1` `source_timestamp=2025-05-09T05:57:00Z`
- At maximum parity of N/2, MinIO ensures uninterrupted read and write operations with only (N/2)+1 operational drives. `claim:claim_minio_ec_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1) `source_document_id=srcdoc_fda12217a5552b410689a3227d6f89c3` `source_revision_id=srcrev_5e4d1c6bec79a677faabbcd501afa9f6` `chunk_id=srcchunk_b45688245d70d499139ad7d596d2f3ec` `native_locator=https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1` `source_timestamp=2025-05-09T05:57:00Z`

## Related Pages

- `projects/erasure-coding-investigation`

## Sources

- `source_document_id`: `srcdoc_fda12217a5552b410689a3227d6f89c3`
- `source_revision_id`: `srcrev_5e4d1c6bec79a677faabbcd501afa9f6`
- `source_url`: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa)
