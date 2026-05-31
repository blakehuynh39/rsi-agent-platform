---
title: "KMS Access Control"
type: "policy"
slug: "policies/kms-access-control"
freshness: "2025-08-11T19:42:00Z"
tags:
  - "access-control"
  - "mvp"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_3a52d49ae3bf7610381c82ceadcc9acd"
conflict_state: "none"
---

# KMS Access Control

## Summary

Access control policy for the KMS, with MVP relying on infrastructure-side controls and future plans for bucket permission verification.

## Claims

- For MVP, access control is performed on the infra side. `claim:claim_4_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`
- Future access control will verify bucket permission. `claim:claim_4_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d) `source_document_id=srcdoc_750823276c4748346e6d08c3aad2079a` `source_revision_id=srcrev_3a52d49ae3bf7610381c82ceadcc9acd` `chunk_id=srcchunk_8526c5293d1bba235ee14736befbff43` `native_locator=https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d` `source_timestamp=2025-08-11T19:42:00Z`

## Open Questions

- What specific bucket permission model will be implemented in the future?

## Related Pages

- `kms-core-system`

## Sources

- `source_document_id`: `srcdoc_750823276c4748346e6d08c3aad2079a`
- `source_revision_id`: `srcrev_3a52d49ae3bf7610381c82ceadcc9acd`
- `source_url`: [Notion source](https://www.notion.so/KMS-Core-PRD-1ed051299a54803fabc4f4c5c2f1589d)
