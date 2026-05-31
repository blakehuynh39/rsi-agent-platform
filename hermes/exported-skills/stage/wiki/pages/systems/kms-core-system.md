---
title: "KMS Core System"
type: "system"
slug: "systems/kms-core-system"
freshness: "2025-08-11T19:42:00Z"
tags:
  - "encryption"
  - "key-management"
  - "kms"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_3a52d49ae3bf7610381c82ceadcc9acd"
conflict_state: "none"
---

# KMS Core System

## Summary

A key management service that splits secrets using Shamir's Secret Sharing and distributes shares across service nodes, with a coordinator handling create and reconstruct requests.

## Claims

- The system consists of a set of service nodes storing shares of secret keys. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`
- Secrets are split using Shamir's Secret Sharing (SSS). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`
- The system exposes an HTTP API for key creation (split + distribute) and key reconstruction (gather + combine). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`
- A per-bucket strategy deterministically maps which nodes store which shares. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`
- The system is designed to scale to more nodes later. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`

## Open Questions

- How will scaling to more nodes be implemented?
- What is the exact consistent hashing scheme for bucket-to-node mapping?

## Related Pages

- `deterministic-bucket-mapping`
- `kms-access-control`
- `kms-coordinator`
- `shamir-secret-sharing`

## Sources

- `source_document_id`: `srcdoc_750823276c4748346e6d08c3aad2079a`
- `source_revision_id`: `srcrev_3a52d49ae3bf7610381c82ceadcc9acd`
- `source_url`: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d)
