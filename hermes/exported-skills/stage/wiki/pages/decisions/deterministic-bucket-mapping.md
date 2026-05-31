---
title: "Deterministic Bucket to Nodes Mapping"
type: "decision"
slug: "decisions/deterministic-bucket-mapping"
freshness: "2025-08-11T19:42:00Z"
tags:
  - "architecture"
  - "consistent-hashing"
  - "distribution"
owners: []
source_revision_ids:
  - "srcrev_3a52d49ae3bf7610381c82ceadcc9acd"
conflict_state: "none"
---

# Deterministic Bucket to Nodes Mapping

## Summary

Decision to use consistent hashing for deterministically mapping buckets to service nodes for share distribution.

## Claims

- To decide which nodes hold shares, use consistent hashing. `claim:claim_3_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`

## Open Questions

- What is the specific consistent hashing algorithm and ring configuration?

## Related Pages

- `kms-core-system`

## Sources

- `source_document_id`: `srcdoc_750823276c4748346e6d08c3aad2079a`
- `source_revision_id`: `srcrev_3a52d49ae3bf7610381c82ceadcc9acd`
- `source_url`: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d)
