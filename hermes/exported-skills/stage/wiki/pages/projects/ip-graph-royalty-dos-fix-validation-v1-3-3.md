---
title: "IP Graph Royalty DoS Fix Validation (v1.3.3)"
type: "project"
slug: "projects/ip-graph-royalty-dos-fix-validation-v1-3-3"
freshness: "2026-06-25T08:00:00Z"
tags:
  - "DoS"
  - "fix"
  - "IPGraph"
  - "royalty"
  - "validation"
owners:
  - "Hans"
source_revision_ids:
  - "srcrev_c2f1ba990cce55659de1c0f732c93705"
conflict_state: "none"
---

# IP Graph Royalty DoS Fix Validation (v1.3.3)

## Summary

Reproduction and validation of the IPGraph royalty gas-underpricing DoS fix on devnet0.

## Claims

- The IP_GRAPH precompile topological sort over ancestor graph is mispriced, costing a flat ~29k gas per call while wall-time scales with V+E, allowing a 16M gas transaction to loop ~575 calls and exceed the block build window. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_c2f1ba990cce55659de1c0f732c93705` `chunk_id=srcchunk_a4288e48057ea2212718621d4b1be56a` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:00:00Z`
- Pre-fix, dense-64, deep-300, lattice-183, and apex-975 graph scenarios caused chain DoS (empty blocks, all txs blocked), while tip-64 (linear, 64 ancestors) did not. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_c2f1ba990cce55659de1c0f732c93705` `chunk_id=srcchunk_832eb48f397ee3a23b7a6c4f43ea978d` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:00:00Z`
- The v1.3.3 fix routes `getPolicyRoyalty` to an expensive precompile selector, increasing per-call gas cost to 222,665 and reducing per-tx reach to ~75 calls, which keeps execution under the build window. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_c2f1ba990cce55659de1c0f732c93705` `chunk_id=srcchunk_a4288e48057ea2212718621d4b1be56a` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:00:00Z`
- Post-fix, all five staller transactions mined as OOG failures, and normal transfers mined in the same or next block, confirming no chain DoS. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_c2f1ba990cce55659de1c0f732c93705` `chunk_id=srcchunk_832eb48f397ee3a23b7a6c4f43ea978d` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:00:00Z`
- The upgrade to fixed implementation was executed via Safe with UPGRADER role through ProtocolAccessManager (schedule tx 0x543b..., execute tx 0xf7ab...). `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_c2f1ba990cce55659de1c0f732c93705` `chunk_id=srcchunk_832eb48f397ee3a23b7a6c4f43ea978d` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:00:00Z`
- The per-transaction gas cap on devnet0 is 16M (block limit 36M), limiting attacker's per-tx reach and making a dense 16-parent deep mesh unbuildable. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_c2f1ba990cce55659de1c0f732c93705` `chunk_id=srcchunk_a4288e48057ea2212718621d4b1be56a` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:00:00Z`

## Sources

- `source_document_id`: `srcdoc_904e0df3db2d9f74c32ad8fd9af152d8`
- `source_revision_id`: `srcrev_c2f1ba990cce55659de1c0f732c93705`
- `source_url`: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c)
