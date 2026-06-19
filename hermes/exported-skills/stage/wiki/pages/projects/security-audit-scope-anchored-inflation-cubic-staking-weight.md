---
title: "Security Audit Scope: Anchored Inflation and Cubic Staking Weight"
type: "project"
slug: "projects/security-audit-scope-anchored-inflation-cubic-staking-weight"
freshness: "2026-06-19T17:36:00Z"
tags:
  - "audit"
  - "cubic-staking"
  - "inflation"
  - "security"
  - "story"
owners: []
source_revision_ids:
  - "srcrev_d838f410a73735dab799d1770e29df97"
conflict_state: "none"
---

# Security Audit Scope: Anchored Inflation and Cubic Staking Weight

## Summary

External security audit of two features: anchored inflation and cubic staking weight (SIP-12), implemented across two private forks with cross-cutting interaction.

## Claims

- The audit scopes two features: anchored inflation in story-private-fork (branch release/1.10) and cubic staking weight in cosmos-sdk-private-fork (branch hans/v0.50.14-sip12-cubic). `claim:claim_3_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- The two repos interact via story-private-fork's go.mod pinning the cubic branch of the SDK fork, and the v1.9.0 upgrade handler that iterates delegations and recomputes cubic RewardsShares. `claim:claim_3_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- The audit diff must be taken against the latest mainnet public releases: story v1.8.0 (commit 0395c719e6072d8d19156a02c84dff5c87bb5ccb) and cosmos-sdk v0.50.14-piplabs-v1.1 (commit 31e0389e1121099d4489e413b3e325c4e4ca3f37). `claim:claim_3_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- To reproduce the diff for anchored inflation, use: git fetch https://github.com/piplabs/story.git v1.8.0 && git diff 0395c719e6072d8d19156a02c84dff5c87bb5ccb release/1.10 -- client/x/mint/types/inflation.go client/x/mint/keeper/abci.go client/app/upgrades/v_1_9_0/upgrades.go `claim:claim_3_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_f87e3999902751cac4ab0629e17b8c7a` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:36:00Z`
- To reproduce the diff for cubic staking weight, use: git fetch https://github.com/piplabs/cosmos-sdk.git v0.50.14-piplabs-v1.1 && git diff 31e0389e1121099d4489e413b3e325c4e4ca3f37 hans/v0.50.14-sip12-cubic -- x/staking/keeper/cubic.go x/staking/keeper/delegation.go x/staking/keeper/slash.go x/staking/keeper/invariants.go x/staking/types/validator.go `claim:claim_3_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_f87e3999902751cac4ab0629e17b8c7a` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:36:00Z`
- Cross-cutting review notes: consensus determinism of the cubic flag (must be a pure function of chainID and height, not derived from CLI flags or config), inflation anchor without on-chain state drift, and cubic vs linear consistency under slashing (verify DelegatorRewardsSharesInvariant). `claim:claim_3_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- The cubic flag setter in app.go is excluded as wiring; if the auditor needs to confirm the pure-function property, they may bring it back into scope. `claim:claim_3_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_f87e3999902751cac4ab0629e17b8c7a` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:36:00Z`

## Related Pages

- `anchored-inflation`
- `cubic-staking-weight`

## Sources

- `source_document_id`: `srcdoc_0dad80217a5ee14b1346d6b0c7c30f89`
- `source_revision_id`: `srcrev_d838f410a73735dab799d1770e29df97`
- `source_url`: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1)
