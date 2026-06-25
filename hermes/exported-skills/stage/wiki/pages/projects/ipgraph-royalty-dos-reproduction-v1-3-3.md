---
title: "IP_GRAPH Royalty DoS Reproduction and Fix (v1.3.3)"
type: "project"
slug: "projects/ipgraph-royalty-dos-reproduction-v1-3-3"
freshness: "2026-06-25T08:19:00Z"
tags:
  - "devnet0"
  - "dos"
  - "gas-pricing"
  - "ip-graph"
  - "royalty"
  - "security"
owners:
  - "Hans"
source_revision_ids:
  - "srcrev_deff00a6b5c51725bbe4bb3a1fa551a5"
conflict_state: "none"
---

# IP_GRAPH Royalty DoS Reproduction and Fix (v1.3.3)

## Summary

Reproduction of the IPGRAPH royalty gas-underpricing denial-of-service attack on devnet0 (chainId 1512) before and (pending) after the v1.3.3 fix. Five graph scenarios confirmed the attack stalls the chain pre-fix; evidence shows that edge density, not ancestor count, drives the DoS. The fix raises per-call gas cost ~7.6×, reducing reachable calls from ~575 to ~75, which keeps execution within the 2-second block-build window. Governance upgrades use the Safe multi-sig and ProtocolAccessManager.

## Claims

- The attacker deploys a contract that loops RoyaltyPolicyLAP.getPolicyRoyalty(node, X) in a single 16M-gas transaction; each call triggers the IP_GRAPH precompile (0x101) topological sort over the node’s ancestor graph (O(V+E)). `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_deff00a6b5c51725bbe4bb3a1fa551a5` `chunk_id=srcchunk_b762c6f63c9068e661361675cf23e72a` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:19:00Z`
- Pre-fix, a single getPolicyRoyalty call costs 29,134 gas, allowing ~575 such calls in one 16M transaction. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_deff00a6b5c51725bbe4bb3a1fa551a5` `chunk_id=srcchunk_b762c6f63c9068e661361675cf23e72a` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:19:00Z`
- Post-fix, the expensive getRoyaltyExt selector raises per‑call gas to 222,665, yielding ~75 calls per 16M transaction — a ~7.6× increase. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_deff00a6b5c51725bbe4bb3a1fa551a5` `chunk_id=srcchunk_b762c6f63c9068e661361675cf23e72a` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:19:00Z`
- A single transaction is capped at 16,777,216 gas (RPC-enforced), preventing an out‑of‑gas block but not protecting against heavy compute within that limit. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_deff00a6b5c51725bbe4bb3a1fa551a5` `chunk_id=srcchunk_b762c6f63c9068e661361675cf23e72a` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:19:00Z`
- Pre-fix, a staller transaction exceeding the 2‑s block‑build window cannot be mined; the proposer ships empty blocks, and all other transactions are stuck. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_deff00a6b5c51725bbe4bb3a1fa551a5` `chunk_id=srcchunk_b762c6f63c9068e661361675cf23e72a` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-1` `source_timestamp=2026-06-25T08:19:00Z`
- Five pre‑fix graph scenarios were reproduced on devnet0 with on‑chain evidence: tip‑64 (64 ancestors, linear), dense‑64 (64 ancestors, dense mesh), deep‑300 (300‑deep linear), lattice‑183 (16×19 diamond mesh), and apex‑975 (15 × 64‑deep chains, matching mainnet node 0x6fea2dda). All except tip‑64 caused a DoS. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_deff00a6b5c51725bbe4bb3a1fa551a5` `chunk_id=srcchunk_897f008877091e69465b286504c33565` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:19:00Z`
- The DoS is driven by the topological‑sort computational cost, approximated by V+E (nodes + edges), not by the raw ancestor count. For example, dense‑64 (768 edges, DoS) has the same 64 ancestors as tip‑64 (64 edges, no DoS). `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_deff00a6b5c51725bbe4bb3a1fa551a5` `chunk_id=srcchunk_897f008877091e69465b286504c33565` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:19:00Z`
- The royalty proxy downgrade and upgrade are performed through the real Safe (2-of-3) → ProtocolAccessManager schedule → wait(1s) → execute path, not an EOA shortcut. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2) `source_document_id=srcdoc_904e0df3db2d9f74c32ad8fd9af152d8` `source_revision_id=srcrev_deff00a6b5c51725bbe4bb3a1fa551a5` `chunk_id=srcchunk_897f008877091e69465b286504c33565` `native_locator=https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c#chunk-2` `source_timestamp=2026-06-25T08:19:00Z`

## Open Questions

- Post‑fix reproduction scenarios are pending the Safe upgrade; actual empty‑block rates and call counts need verification on devnet0 after the v1.3.3 implementation is live.

## Sources

- `source_document_id`: `srcdoc_904e0df3db2d9f74c32ad8fd9af152d8`
- `source_revision_id`: `srcrev_deff00a6b5c51725bbe4bb3a1fa551a5`
- `source_url`: [source](https://app.notion.com/p/devnet0-before-after-reproduction-graph-scenarios-v1-3-3-fix-38a051299a5481ec9d4df70330d4263c)
