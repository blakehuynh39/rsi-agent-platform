---
title: "Erasure Coding Investigation"
type: "project"
slug: "projects/erasure-coding-investigation"
freshness: "2025-05-09T05:57:00Z"
tags:
  - "erasure-coding"
  - "investigation"
  - "minio"
owners: []
source_revision_ids:
  - "srcrev_5e4d1c6bec79a677faabbcd501afa9f6"
  - "srcrev_d7f43e0edda91b3da414096e586811c3"
conflict_state: "none"
---

# Erasure Coding Investigation

## Summary

Investigation into erasure coding, with MinIO as a candidate technology.

## Claims

- Erasure coding is supported by MinIO. `claim:claim_4_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Technical-Challenges-1dc051299a5480ff8fe8e471808534a0) `source_document_id=srcdoc_308f44f99cbee47555ac43ecc4b44bd3` `source_revision_id=srcrev_d7f43e0edda91b3da414096e586811c3` `chunk_id=srcchunk_77b29b70add3731122648850031ee9b7` `native_locator=https://www.notion.so/Technical-Challenges-1dc051299a5480ff8fe8e471808534a0` `source_timestamp=2025-04-21T20:33:00Z`
- Further investigation into erasure coding is planned. `claim:claim_4_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Technical-Challenges-1dc051299a5480ff8fe8e471808534a0) `source_document_id=srcdoc_308f44f99cbee47555ac43ecc4b44bd3` `source_revision_id=srcrev_d7f43e0edda91b3da414096e586811c3` `chunk_id=srcchunk_77b29b70add3731122648850031ee9b7` `native_locator=https://www.notion.so/Technical-Challenges-1dc051299a5480ff8fe8e471808534a0` `source_timestamp=2025-04-21T20:33:00Z`
- MinIO uses per-object inline Reed-Solomon erasure coding with configurable redundancy levels. `claim:claim_minio_ec_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1) `source_document_id=srcdoc_fda12217a5552b410689a3227d6f89c3` `source_revision_id=srcrev_5e4d1c6bec79a677faabbcd501afa9f6` `chunk_id=srcchunk_b45688245d70d499139ad7d596d2f3ec` `native_locator=https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1` `source_timestamp=2025-05-09T05:57:00Z`
- Healing is performed at the object level and can heal multiple objects independently. `claim:claim_minio_ec_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1) `source_document_id=srcdoc_fda12217a5552b410689a3227d6f89c3` `source_revision_id=srcrev_5e4d1c6bec79a677faabbcd501afa9f6` `chunk_id=srcchunk_b45688245d70d499139ad7d596d2f3ec` `native_locator=https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1` `source_timestamp=2025-05-09T05:57:00Z`
- At maximum parity of N/2, MinIO ensures uninterrupted read and write operations with only (N/2)+1 operational drives. `claim:claim_minio_ec_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1) `source_document_id=srcdoc_fda12217a5552b410689a3227d6f89c3` `source_revision_id=srcrev_5e4d1c6bec79a677faabbcd501afa9f6` `chunk_id=srcchunk_b45688245d70d499139ad7d596d2f3ec` `native_locator=https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-1` `source_timestamp=2025-05-09T05:57:00Z`
- MinIO requires homogeneous capability in a cluster. `claim:claim_minio_heterogeneous` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-2) `source_document_id=srcdoc_fda12217a5552b410689a3227d6f89c3` `source_revision_id=srcrev_5e4d1c6bec79a677faabbcd501afa9f6` `chunk_id=srcchunk_d0fe56c398be1146e7082c15b3b38eec` `native_locator=https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa#chunk-2` `source_timestamp=2025-05-09T05:57:00Z`

## Related Pages

- `concepts/minio-erasure-coding`

## Sources

- `source_document_id`: `srcdoc_fda12217a5552b410689a3227d6f89c3`
- `source_revision_id`: `srcrev_5e4d1c6bec79a677faabbcd501afa9f6`
- `source_url`: [Notion source](https://www.notion.so/MinIO-1df051299a5480efbd86c2c12ddb84fa)
