---
title: "Story Kernel Changelog (v0.1.0 to Feature)"
type: "project"
slug: "projects/story-kernel-changelog"
freshness: "2026-06-17T07:45:00Z"
tags:
  - "changelog"
  - "dkg"
  - "light-client"
  - "security"
  - "story-kernel"
owners: []
source_revision_ids:
  - "srcrev_f27c37876ea2b92518e3601b12a91b7b"
conflict_state: "none"
---

# Story Kernel Changelog (v0.1.0 to Feature)

## Summary

Changelog of merged PRs for Story kernel from base v0.1.0 to a feature branch, covering versioning, light-client improvements, DKG protocol fixes, and resilience enhancements.

## Claims

- Added a story-kernel version command that prints version and git commit/timestamp, using ldflags for build-time injection due to SGX reproducible build requirements. `claim:version-command` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_f27c37876ea2b92518e3601b12a91b7b` `chunk_id=srcchunk_e0115d3ff1346e21be58fe6c7ae8bd70` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-17T07:45:00Z`
- initializeQueryClient now only falls back to config trusted block (clearing sealed DB state) on genuinely invalid state errors (ErrOldHeaderExpired/ErrInvalidHeader); transient errors return immediately without touching DB. `claim:lc-db-wipe-narrowing` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_f27c37876ea2b92518e3601b12a91b7b` `chunk_id=srcchunk_e0115d3ff1346e21be58fe6c7ae8bd70` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-17T07:45:00Z`
- Kernel refuses to start if sealed light-client state is missing or expired, moves sealed-state check before ClearTrustedState, enforces TDH2Verify result in PartialDecryptTDH2, and documents recovery procedure. `claim:lc-recovery-hardening` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_f27c37876ea2b92518e3601b12a91b7b` `chunk_id=srcchunk_e0115d3ff1346e21be58fe6c7ae8bd70` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-17T07:45:00Z`
- ProcessDeals/Responses/Justifications now return a rejected_* list of items that could not be absorbed, allowing individual retries; idempotent re-submissions are silently skipped and excluded from rejected lists. `claim:rejected-lists` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_f27c37876ea2b92518e3601b12a91b7b` `chunk_id=srcchunk_e0115d3ff1346e21be58fe6c7ae8bd70` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-17T07:45:00Z`
- Fixed DKG stall by introducing ErrLightClientLag sentinel, retrying GetOrLoadRoundContext on it, and allowing early deals to prevent finalization deadlock. `claim:lc-lag-dkg-stall` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_f27c37876ea2b92518e3601b12a91b7b` `chunk_id=srcchunk_e0115d3ff1346e21be58fe6c7ae8bd70` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-17T07:45:00Z`
- Kernel now waits for finalization stage before computing participants root to ensure it uses the post-invalidation registration set (partial due to truncation). `claim:participants-root-resharing` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_f27c37876ea2b92518e3601b12a91b7b` `chunk_id=srcchunk_e0115d3ff1346e21be58fe6c7ae8bd70` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-17T07:45:00Z`

## Open Questions

- The changelog source is truncated; the participants-root fix description is incomplete and there may be additional changes not captured.

## Sources

- `source_document_id`: `srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c`
- `source_revision_id`: `srcrev_f27c37876ea2b92518e3601b12a91b7b`
- `source_url`: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340)
