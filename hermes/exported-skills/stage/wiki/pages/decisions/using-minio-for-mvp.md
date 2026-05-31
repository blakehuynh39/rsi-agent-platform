---
title: "Using MinIO for MVP"
type: "decision"
slug: "decisions/using-minio-for-mvp"
freshness: "2025-08-11T19:06:00Z"
tags:
  - "bft"
  - "dsync"
  - "minio"
  - "mvp"
  - "storage"
owners: []
source_revision_ids:
  - "srcrev_d4c31179ecfe6a8eaee983d580515eaa"
conflict_state: "none"
---

# Using MinIO for MVP

## Summary

A three-step plan for integrating MinIO into the MVP storage layer, starting with an honest-node assumption and progressing toward a BFT-compatible, heterogeneous storage backend.

## Claims

- The dsync package provides distributed locking for up to 16 nodes, requiring n/2 + 1 positive responses to acquire a lock, and is not BFT. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c) `source_document_id=srcdoc_213d428c566d9737e45678ecae9c2eca` `source_revision_id=srcrev_d4c31179ecfe6a8eaee983d580515eaa` `chunk_id=srcchunk_384aeb69fc24220e70f7b2016ded97cd` `native_locator=https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c` `source_timestamp=2025-08-11T19:06:00Z`
- dsync must be removed if untrusted parties are allowed to run storage nodes because it has strong trust assumptions. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c) `source_document_id=srcdoc_213d428c566d9737e45678ecae9c2eca` `source_revision_id=srcrev_d4c31179ecfe6a8eaee983d580515eaa` `chunk_id=srcchunk_384aeb69fc24220e70f7b2016ded97cd` `native_locator=https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c` `source_timestamp=2025-08-11T19:06:00Z`
- The first step assumes all nodes are honest and uses MinIO in a single multi-region, multi-drive cluster with on-chain events triggering bucket creation with quota and retention policy, and OpenID Connect with SIWE for authentication. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c) `source_document_id=srcdoc_213d428c566d9737e45678ecae9c2eca` `source_revision_id=srcrev_d4c31179ecfe6a8eaee983d580515eaa` `chunk_id=srcchunk_384aeb69fc24220e70f7b2016ded97cd` `native_locator=https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c` `source_timestamp=2025-08-11T19:06:00Z`
- The second step involves investigating non-BFT components of MinIO (such as dsync and homogeneous node capability assumptions) and replacing them with BFT code. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c) `source_document_id=srcdoc_213d428c566d9737e45678ecae9c2eca` `source_revision_id=srcrev_d4c31179ecfe6a8eaee983d580515eaa` `chunk_id=srcchunk_384aeb69fc24220e70f7b2016ded97cd` `native_locator=https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c` `source_timestamp=2025-08-11T19:06:00Z`
- The third step considers heterogeneous clusters where not all storage nodes run MinIO, and may include S3, SeaweedFS, or PostgreSQL. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c) `source_document_id=srcdoc_213d428c566d9737e45678ecae9c2eca` `source_revision_id=srcrev_d4c31179ecfe6a8eaee983d580515eaa` `chunk_id=srcchunk_384aeb69fc24220e70f7b2016ded97cd` `native_locator=https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c` `source_timestamp=2025-08-11T19:06:00Z`
- Multipart uploads (MPU) are disabled, object overwrite is never permitted, and deletes are serialized client-side or happen outside MinIO (e.g., via object expiry). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c) `source_document_id=srcdoc_213d428c566d9737e45678ecae9c2eca` `source_revision_id=srcrev_d4c31179ecfe6a8eaee983d580515eaa` `chunk_id=srcchunk_384aeb69fc24220e70f7b2016ded97cd` `native_locator=https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c` `source_timestamp=2025-08-11T19:06:00Z`

## Sources

- `source_document_id`: `srcdoc_213d428c566d9737e45678ecae9c2eca`
- `source_revision_id`: `srcrev_d4c31179ecfe6a8eaee983d580515eaa`
- `source_url`: [Notion source](https://www.notion.so/Using-MinIO-for-MVP-1e0051299a5480a6980fe8c1eaa5188c)
