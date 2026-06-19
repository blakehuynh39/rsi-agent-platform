---
title: "Anchored Inflation + Cubic Staking Weight Security Audit Scope"
type: "project"
slug: "projects/anchored-inflation-cubic-staking-weight-security-audit-scope"
freshness: "2026-06-19T17:07:00Z"
tags:
  - "anchored-inflation"
  - "cosmos-sdk"
  - "cubic-staking-weight"
  - "security-audit"
  - "story"
owners: []
source_revision_ids:
  - "srcrev_5e47205d423d9644146d60d614215de4"
conflict_state: "none"
---

# Anchored Inflation + Cubic Staking Weight Security Audit Scope

## Summary

Security audit scope for the anchored inflation and cubic staking weight features across the Story and Cosmos SDK private forks.

## Claims

- The external security audit scopes two features: anchored inflation (in story-private-fork) and cubic staking weight (in cosmos-sdk-private-fork). `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_5e47205d423d9644146d60d614215de4` `chunk_id=srcchunk_106544184b9b0e9c3b21e3c0c15b8334` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:07:00Z`
- The anchored inflation feature implements a decaying-rate inflation schedule applied per-block to a compounding supply, plus a v1.9.0 on-chain upgrade handler. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_5e47205d423d9644146d60d614215de4` `chunk_id=srcchunk_106544184b9b0e9c3b21e3c0c15b8334` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:07:00Z`
- The cubic staking weight feature makes a validator's reward weight scale with the cube of stake. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_5e47205d423d9644146d60d614215de4` `chunk_id=srcchunk_106544184b9b0e9c3b21e3c0c15b8334` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:07:00Z`
- The cubic weight is activated by an in-memory, height-based flag set every PreBlock, never persisted to the KV store. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_5e47205d423d9644146d60d614215de4` `chunk_id=srcchunk_106544184b9b0e9c3b21e3c0c15b8334` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:07:00Z`
- The audit diff for story-private-fork is taken against public v1.8.0 release commit 0395c719e6072d8d19156a02c84dff5c87bb5ccb. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_5e47205d423d9644146d60d614215de4` `chunk_id=srcchunk_106544184b9b0e9c3b21e3c0c15b8334` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:07:00Z`
- The audit diff for cosmos-sdk-private-fork is taken against public v0.50.14-piplabs-v1.1 release commit 31e0389e1121099d4489e413b3e325c4e4ca3f37. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_5e47205d423d9644146d60d614215de4` `chunk_id=srcchunk_106544184b9b0e9c3b21e3c0c15b8334` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:07:00Z`
- The combined in-scope audit surface is approximately 1,134 changed lines of code across 11 core files. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_5e47205d423d9644146d60d614215de4` `chunk_id=srcchunk_eadc393b9229ee7c4a488d7aa8cf0343` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:07:00Z`
- The keeper file x/staking/keeper/keeper.go contains in-memory cubicWeightEnabled / cubicWeightScale fields that are never persisted to KV. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_5e47205d423d9644146d60d614215de4` `chunk_id=srcchunk_eadc393b9229ee7c4a488d7aa8cf0343` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:07:00Z`

## Sources

- `source_document_id`: `srcdoc_0dad80217a5ee14b1346d6b0c7c30f89`
- `source_revision_id`: `srcrev_5e47205d423d9644146d60d614215de4`
- `source_url`: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1)
