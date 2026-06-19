---
title: "Cubic Staking Weight (SIP-12)"
type: "system"
slug: "systems/cubic-staking-weight"
freshness: "2026-06-19T17:36:00Z"
tags:
  - "audit"
  - "cubic"
  - "sip-12"
  - "staking"
  - "story"
owners: []
source_revision_ids:
  - "srcrev_d838f410a73735dab799d1770e29df97"
conflict_state: "none"
---

# Cubic Staking Weight (SIP-12)

## Summary

A validator's reward weight scales with the cube of stake, with cubic-aware reward-share accounting, partial-unbond reduction, and slashing.

## Claims

- Cubic staking weight scales a validator's reward weight by (bondAmt/scale)^3 × tokenMult × periodMult, with cubic-aware reward-share accounting, partial-unbond reduction, and slashing. `claim:claim_2_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- Implemented in cosmos-sdk-private-fork, branch hans/v0.50.14-sip12-cubic. `claim:claim_2_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- The diff baseline is public v0.50.14-piplabs-v1.1 release of cosmos-sdk (commit 31e0389e1121099d4489e413b3e325c4e4ca3f37). `claim:claim_2_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- In-scope source files for cubic staking weight: x/staking/keeper/cubic.go, x/staking/keeper/delegation.go, x/staking/keeper/slash.go, x/staking/keeper/invariants.go, x/staking/types/validator.go. `claim:claim_2_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_f87e3999902751cac4ab0629e17b8c7a` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:36:00Z`
- Cubic weight activation is controlled by an in-memory, height-based flag set every PreBlock from applyRuntimeForkFlags() based on netconf.IsV190; it is never persisted to the KV store. `claim:claim_2_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_ca58d887a6e28f9692c82c9ab412d521` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-1` `source_timestamp=2026-06-19T17:36:00Z`
- Under slashing, RemoveTokensCubic scales reward weight by (1−f)^3 while the distribution hook receives 1−(1−f)^3; the auditor must verify the DelegatorRewardsSharesInvariant is not violated by bounded drift. `claim:claim_2_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2) `source_document_id=srcdoc_0dad80217a5ee14b1346d6b0c7c30f89` `source_revision_id=srcrev_d838f410a73735dab799d1770e29df97` `chunk_id=srcchunk_f87e3999902751cac4ab0629e17b8c7a` `native_locator=https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1#chunk-2` `source_timestamp=2026-06-19T17:36:00Z`

## Related Pages

- `anchored-inflation`
- `security-audit-scope-anchored-inflation-cubic-staking-weight`

## Sources

- `source_document_id`: `srcdoc_0dad80217a5ee14b1346d6b0c7c30f89`
- `source_revision_id`: `srcrev_d838f410a73735dab799d1770e29df97`
- `source_url`: [source](https://app.notion.com/p/Security-Audit-Scope-Anchored-Inflation-Cubic-Staking-Weight-384051299a548150a032d032621aaee1)
