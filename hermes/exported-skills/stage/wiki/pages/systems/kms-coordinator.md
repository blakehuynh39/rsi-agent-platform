---
title: "KMS Coordinator"
type: "system"
slug: "systems/kms-coordinator"
freshness: "2025-08-11T19:42:00Z"
tags:
  - "api"
  - "coordinator"
  - "key-management"
owners: []
source_revision_ids:
  - "srcrev_3a52d49ae3bf7610381c82ceadcc9acd"
conflict_state: "none"
---

# KMS Coordinator

## Summary

The coordinator (or any service node) handles key creation and reconstruction requests, splitting keys, distributing shares, and combining them.

## Claims

- The coordinator receives a create request, generates a random AES key (32 bytes), splits it into shares (n shares, threshold t), uses deterministic bucket-to-nodes mapping to send shares, and sends them to service nodes. `claim:claim_5_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`
- The coordinator receives a reconstruct request, contacts t nodes responsible for the bucket, collects shares, combines the secret using SSS, and returns the decrypted key. `claim:claim_5_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`
- The coordinator exposes POST /create and GET /reconstruct handlers. `claim:claim_5_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`

## Open Questions

- Can any service node act as coordinator, or is there a dedicated coordinator?

## Related Pages

- `deterministic-bucket-mapping`
- `kms-core-system`

## Sources

- `source_document_id`: `srcdoc_750823276c4748346e6d08c3aad2079a`
- `source_revision_id`: `srcrev_3a52d49ae3bf7610381c82ceadcc9acd`
- `source_url`: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d)
