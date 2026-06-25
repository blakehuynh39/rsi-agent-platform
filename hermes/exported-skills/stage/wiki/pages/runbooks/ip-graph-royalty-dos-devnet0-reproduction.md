---
title: "IP Graph Royalty DoS Fix Reproduction on Devnet0"
type: "runbook"
slug: "runbooks/ip-graph-royalty-dos-devnet0-reproduction"
freshness: "2026-06-25T08:43:00Z"
tags:
  - "devnet0"
  - "dos"
  - "ip-graph"
  - "security"
  - "v1.3.3"
owners:
  - "Hans"
source_revision_ids:
  - "srcrev_7138f4f8f97dba51312e153701cb67cf"
conflict_state: "none"
---

# IP Graph Royalty DoS Fix Reproduction on Devnet0

## Summary

Reproduction of the IPGraph royalty gas-underpricing DoS on devnet0 before and after v1.3.3 fix, validating the attack stalls the chain pre-fix and is bounded post-fix across 5 graph shapes and all 3 attacker-reachable reads. Fix validated with on-chain tx evidence.

## Claims

- Goal: prove the attack stalls the chain pre-fix and is bounded post-fix, with on-chain tx evidence for every scenario. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_612ff539f1104c16f8519a0680ff962b` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:43:00Z`
- Status: COMPLETE — BEFORE + AFTER done. Fix validated across 5 graph shapes AND all 3 attacker-reachable reads. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_612ff539f1104c16f8519a0680ff962b` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:43:00Z`
- Attack mechanism: attacker contract loops `RoyaltyPolicyLAP.getPolicyRoyalty(node, X)` in one 16M-gas tx, causing IP_GRAPH precompile (0x101) to run O(V+E) topological sort over node's ancestor graph. Pre-fix flat gas ~29k/call allows ~575 calls, exceeding the 2s block-build window and forcing empty blocks. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_612ff539f1104c16f8519a0680ff962b` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:43:00Z`
- Post-fix: `getRoyaltyExt` selector costs ~222,665 gas per call (7.6x), limiting a 16M tx to ~75 calls and restoring execution under the build window. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_612ff539f1104c16f8519a0680ff962b` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:43:00Z`
- Hard per-transaction gas cap is 16,777,216 (16M), enforced by RPC. This caps attacker's per-tx reach and also limits graph construction (e.g., dense 16-parent mesh unbuildable past depth ~5 due to registerDerivative cost). `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_612ff539f1104c16f8519a0680ff962b` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:43:00Z`
- Tested graph shapes and results: tip-64 (linear 64 ancestors, 128 V+E → NO jam); dense-64 (mesh 64 ancestors, 832 V+E → DoS); deep-300 (linear 300 ancestors, 600 V+E → DoS); lattice-183 (diamond mesh 183 ancestors, 520 V+E → DoS); apex-975 (wide tree 975 ancestors, 1950 V+E → DoS). `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_b82ea724ff6d56c654c48cd8e80e26a7` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:43:00Z`
- Key insight: edge density, not ancestor count, drives the DoS. Dense-64 (same 64 ancestors as tip-64 but 768 edges) stalls, while tip-64 (64 edges) does not. The cost tracks the precompile's topological sort O(V+E). `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_b82ea724ff6d56c654c48cd8e80e26a7` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:43:00Z`
- Governance upgrade path: royalty proxies were downgraded/upgraded via the real 2-of-3 Safe → ProtocolAccessManager schedule → wait(1s) → execute path, not an EOA shortcut. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_b82ea724ff6d56c654c48cd8e80e26a7` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:43:00Z`
- All three attacker-reachable reads were patched: RoyaltyPolicyLAP.getPolicyRoyalty (pre ~29k → post ~222k), RoyaltyPolicyLRP.getPolicyRoyalty (~29k → ~163k), RoyaltyModule.hasAncestorIp (~29k → ~159k). `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-3) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_e60e88162bc87735f583033b808798dd` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-3` `source_timestamp=2026-06-25T08:43:00Z`
- Recovery during pre-fix DoS was possible only because the tester controlled the staller sender key (replaced tx with same nonce). A real attacker's staller cannot be replaced by validators. Real mitigations are the v1.3.3 contract fix plus a geth-level wall-clock watchdog/sender denylist. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-3) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_e60e88162bc87735f583033b808798dd` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-3` `source_timestamp=2026-06-25T08:43:00Z`
- Post-fix validation: the staller tx mines as a harmless 16M out-of-gas, and a normal transfer from another account mines successfully, confirming the DoS is bounded. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-3) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_e60e88162bc87735f583033b808798dd` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-3` `source_timestamp=2026-06-25T08:43:00Z`
- Reproduction uses deterministic addresses (create3 seed 6 = mainnet-identical), so the same proxy addresses and exploit/build contracts can be used unchanged on Aeneid and mainnet. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-3) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_7138f4f8f97dba51312e153701cb67cf` `chunk_id=srcchunk_e60e88162bc87735f583033b808798dd` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-3` `source_timestamp=2026-06-25T08:43:00Z`

## Sources

- `source_document_id`: `srcdoc_904e0df3db2d9f74c32ad8fd9af152d8`
- `source_revision_id`: `srcrev_7138f4f8f97dba51312e153701cb67cf`
- `source_url`: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c)
