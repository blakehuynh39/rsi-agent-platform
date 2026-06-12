---
title: "Story-kernel Change Log: v0.1.0 to Feature"
type: "project"
slug: "projects/story-kernel-changelog-v0-1-0-to-feature"
freshness: "2026-06-12T04:45:00Z"
tags:
  - "changelog"
  - "DKG"
  - "light-client"
  - "SGX"
  - "story-kernel"
owners: []
source_revision_ids:
  - "srcrev_33d2c38861421ef64092e89f2409ccee"
conflict_state: "none"
---

# Story-kernel Change Log: v0.1.0 to Feature

## Summary

Log of changes from base v0.1.0 to Feature branch, including PRs for version command, light-client recovery improvements, typed rejected lists, and light-client lag fix.

## Claims

- PR #57 adds a `story-kernel version` command and startup log line printing version plus git commit/timestamp. Git info is injected at build time via Makefile ldflags -X instead of runtime/debug, because SGX reproducible build uses -buildvcs=false for deterministic MRENCLAVE. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_33d2c38861421ef64092e89f2409ccee` `chunk_id=srcchunk_dd3ec3f48081d646c51b5e56dac011c9` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T04:45:00Z`
- PR #61 narrows when the light-client DB is wiped: `initializeQueryClient` only falls back to config trusted block (clearing sealed DB state) when `LoadVerifiedQueryClient` fails with `ErrOldHeaderExpired` or `ErrInvalidHeader` (genuinely invalid state). All other errors (transient network, `ErrNoWitnesses`, etc.) return immediately without touching the DB. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_33d2c38861421ef64092e89f2409ccee` `chunk_id=srcchunk_dd3ec3f48081d646c51b5e56dac011c9` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T04:45:00Z`
- PR #62 hardens light-client recovery after DKG registration: kernel refuses to start (forcing re-registration) when sealed light-client state is missing or expired, instead of silently falling back to config.toml; moves sealed-state check before ClearTrustedState so valid DB state is not destroyed; enforces (not just logs) the TDH2Verify result in PartialDecryptTDH2; and documents the recovery procedure in the README. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_33d2c38861421ef64092e89f2409ccee` `chunk_id=srcchunk_dd3ec3f48081d646c51b5e56dac011c9` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T04:45:00Z`
- PR #63 makes Process{Deals,Responses,Justifications} responses return a typed `rejected_*` list of items the kernel could not absorb, so the consensus layer can retry them individually. Idempotent re-submissions (deal/response/justification already received) are silently skipped and excluded from `rejected_*`, preventing gossip duplicates from amplifying retry traffic. Backward-compatible (new fields at unused proto tags). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_33d2c38861421ef64092e89f2409ccee` `chunk_id=srcchunk_dd3ec3f48081d646c51b5e56dac011c9` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T04:45:00Z`
- PR #67 fixes the kernel side of the light-client-lag DKG stall (story#826). Adds an `ErrLightClientLag` sentinel for when the light client reads Total=0 but already sees registrations, and causes the registration to wait for the light client to catch up instead of failing. (Chunk truncated, exact behavior not fully described.) `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340) `source_document_id=srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c` `source_revision_id=srcrev_33d2c38861421ef64092e89f2409ccee` `chunk_id=srcchunk_dd3ec3f48081d646c51b5e56dac011c9` `native_locator=https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340` `source_timestamp=2026-06-12T04:45:00Z`

## Related Pages

- `dkg`
- `light-client`
- `story-kernel`

## Sources

- `source_document_id`: `srcdoc_0b44382a4d934d2ac194b21a6fa2ed5c`
- `source_revision_id`: `srcrev_33d2c38861421ef64092e89f2409ccee`
- `source_url`: [source](https://app.notion.com/p/CDR-Story-kernel-Change-Log-37d051299a54800b9954ce9c49999340)
