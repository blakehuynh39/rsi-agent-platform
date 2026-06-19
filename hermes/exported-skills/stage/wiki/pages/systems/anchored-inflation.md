---
title: "Anchored Inflation"
type: "system"
slug: "systems/anchored-inflation"
freshness: "2026-06-19T17:36:00Z"
tags:
  - "audit"
  - "inflation"
  - "staking"
  - "story"
owners: []
source_revision_ids:
  - "srcrev_d838f410a73735dab799d1770e29df97"
conflict_state: "none"
---

# Anchored Inflation

## Summary

Decaying-rate inflation schedule applied per-block to a compounding supply, anchored by a v1.9.0 upgrade handler.

## Claims

- Anchored inflation is a decaying-rate inflation schedule with initial rate decaying each year, applied per-block to a compounding supply, plus the v1.9.0 on-chain upgrade handler that anchors the schedule and recomputes reward shares. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- The anchored inflation feature is implemented in the story-private-fork repository, branch release/1.10. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- The diff baseline for the auditor is the public v1.8.0 release of story (commit 0395c719e6072d8d19156a02c84dff5c87bb5ccb). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- In-scope source files for anchored inflation: client/x/mint/types/inflation.go, client/x/mint/keeper/abci.go, client/app/upgrades/v_1_9_0/upgrades.go. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_f87e3999902751cac4ab0629e17b8c7a` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:36:00Z`
- The year-1 anchor is computed once in the upgrade handler and planted in mint.Params; subsequent per-block mints are derived from it and must be byte-identical across all validators and across replay, with no accumulating rounding drift. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_f87e3999902751cac4ab0629e17b8c7a` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:36:00Z`

## Related Pages

- `cubic-staking-weight`
- `security-audit-scope-anchored-inflation-cubic-staking-weight`

## Sources

- `source_document_id`: `srcdoc_0dad80217a5ee14b1346d6b0c7c30f89`
- `source_revision_id`: `srcrev_d838f410a73735dab799d1770e29df97`
- `source_url`: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1)
