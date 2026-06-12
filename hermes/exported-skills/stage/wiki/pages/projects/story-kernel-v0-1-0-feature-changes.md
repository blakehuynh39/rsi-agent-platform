---
title: "Story-kernel v0.1.0 Feature Changes"
type: "project"
slug: "projects/story-kernel-v0-1-0-feature-changes"
freshness: "2026-06-12T06:38:00Z"
tags:
  - "release"
  - "SGX"
  - "story-kernel"
owners: []
source_revision_ids:
  - "srcrev_fae21b5d2b463f388f8d51f75b4d5d76"
conflict_state: "none"
---

# Story-kernel v0.1.0 Feature Changes

## Summary

Summary of merged changes from base v0.1.0 into feature branch as of 2026-06-12.

## Claims

- Added `story-kernel version` command (and startup log) printing version plus git commit/timestamp, using Makefile's `-ldflags -X` instead of `runtime/debug` for SGX reproducible build (due to `-buildvcs=false`). `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_fae21b5d2b463f388f8d51f75b4d5d76` `chunk_id=srcchunk_613b2397afebebb55936ca02f0f17dd1` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T06:38:00Z`
- Narrowed when light-client DB is wiped: `initializeQueryClient` only falls back to config trusted block (clearing sealed DB state) when `LoadVerifiedQueryClient` fails with `ErrOldHeaderExpired`/`ErrInvalidHeader`; all other errors (transient network, `ErrNoWitnesses`, etc.) return immediately without touching DB. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_fae21b5d2b463f388f8d51f75b4d5d76` `chunk_id=srcchunk_613b2397afebebb55936ca02f0f17dd1` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T06:38:00Z`
- Hardened light-client recovery after DKG registration: kernel refuses to start (forcing re-registration) when sealed light-client state is missing or expired, instead of silently falling back to `config.toml`; moves sealed-state check before `ClearTrustedState`; enforces `TDH2Verify` result in `PartialDecryptTDH2`; documented recovery procedure in README. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_fae21b5d2b463f388f8d51f75b4d5d76` `chunk_id=srcchunk_613b2397afebebb55936ca02f0f17dd1` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T06:38:00Z`
- `Process{Deals,Responses,Justifications}` now return a typed `rejected_*` list of items the kernel could not absorb, allowing consensus layer to retry individually instead of failing batch; idempotent re-submissions are silently skipped; backward-compatible via unused proto tags. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_fae21b5d2b463f388f8d51f75b4d5d76` `chunk_id=srcchunk_613b2397afebebb55936ca02f0f17dd1` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T06:38:00Z`
- Fixed kernel side of light-client-lag DKG stall (story#826): added `ErrLightClientLag` sentinel for when light client reads `Total=0` but sees registrations; taught `GetOrLoadRoundContext` to retry instead of bailing on generic count-mismatch; early deals no longer rejected, preventing stuck finalization. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_fae21b5d2b463f388f8d51f75b4d5d76` `chunk_id=srcchunk_613b2397afebebb55936ca02f0f17dd1` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T06:38:00Z`
- Fixed participants-root mismatch during resharing (kernel#71): kernel now waits via `waitForFinalizationRegistrations` (polls `GetDKGNetwork().Stage` until `DKG_STAGE_FINALIZATION`, 5×2s, else `ErrLightClientLag`) so monotonic trusted-height read reflects correct post-invalidation registration set, avoiding root mismatch. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_fae21b5d2b463f388f8d51f75b4d5d76` `chunk_id=srcchunk_613b2397afebebb55936ca02f0f17dd1` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T06:38:00Z`

## Sources

- `source_document_id`: `srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c`
- `source_revision_id`: `srcrev_fae21b5d2b463f388f8d51f75b4d5d76`
- `source_url`: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340)
