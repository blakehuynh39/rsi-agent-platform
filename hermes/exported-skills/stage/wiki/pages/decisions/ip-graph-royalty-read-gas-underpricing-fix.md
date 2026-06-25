---
title: "IP Graph Royalty Read Gas Underpricing Fix"
type: "decision"
slug: "decisions/ip-graph-royalty-read-gas-underpricing-fix"
freshness: "2026-06-25T10:41:00Z"
tags:
  - "gas"
  - "ip-graph"
  - "precompile"
  - "royalty"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_791f79a038a5c59ebf2bcb333ce65c58"
conflict_state: "none"
---

# IP Graph Royalty Read Gas Underpricing Fix

## Summary

Decision to switch three public-facing IP graph royalty precompile call sites from cheap internal selectors to external selectors, raising gas cost from ~600-900 to ~126k-189k per call, to mitigate a gas underpricing attack that could force expensive traversals on all validating nodes for negligible cost.

## Claims

- The fix (commit 2c29808 on branch fix/ipgraph-royalty-gas-underpricing, repo piplabs/protocol-core-v1-private-patch) changes three precompile call sites from the cheap/internal selector to the external selector and removes the dead helper RoyaltyModule._getAncestorCount. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6) `source_document_id=srcdoc_11286398b5cf997ee6d2913f90690722` `source_revision_id=srcrev_791f79a038a5c59ebf2bcb333ce65c58` `chunk_id=srcchunk_54acc0e89d2bdb080c7cc42fb136136b` `native_locator=https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6` `source_timestamp=2026-06-25T10:41:00Z`
- The gas pricing for the IP graph precompile at 0x0101 is keyed solely to the 4-byte selector, with two tiers: ipGraphReadGas = 10 for cheap/internal selectors (getRoyalty, getRoyaltyStack, hasAncestorIp, getAncestorIps…) and ipGraphExternalReadGas = 2100 for external selectors (…Ext), a 210× difference. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6) `source_document_id=srcdoc_11286398b5cf997ee6d2913f90690722` `source_revision_id=srcrev_791f79a038a5c59ebf2bcb333ce65c58` `chunk_id=srcchunk_54acc0e89d2bdb080c7cc42fb136136b` `native_locator=https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6` `source_timestamp=2026-06-25T10:41:00Z`
- Both gas tiers dispatch to the identical Go implementation that performs findAncestors (DFS) and topologicalSort + getRoyaltyLap/Lrp, traversing the actual on-chain graph bounded by MAX_ANCESTORS=1024 and MAX_PARENTS=16, but gas cost is a flat constant set assuming averageAncestorIpCount=30, not the true graph size, so both tiers underprice large graphs. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6) `source_document_id=srcdoc_11286398b5cf997ee6d2913f90690722` `source_revision_id=srcrev_791f79a038a5c59ebf2bcb333ce65c58` `chunk_id=srcchunk_54acc0e89d2bdb080c7cc42fb136136b` `native_locator=https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6` `source_timestamp=2026-06-25T10:41:00Z`
- Pre-fix, public unauthenticated functions (RoyaltyPolicyLAP/LRP.getPolicyRoyalty, transferToVault, RoyaltyModule.hasAncestorIp) used the cheap selectors, enabling an attacker to force a full DFS/topological traversal on every validating validator for only ~600–900 gas by looping reads on a deep ancestor graph (up to 1024 ancestors). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6) `source_document_id=srcdoc_11286398b5cf997ee6d2913f90690722` `source_revision_id=srcrev_791f79a038a5c59ebf2bcb333ce65c58` `chunk_id=srcchunk_54acc0e89d2bdb080c7cc42fb136136b` `native_locator=https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6` `source_timestamp=2026-06-25T10:41:00Z`
- The fix mitigates the attack by switching those public-facing calls to external selectors, raising gas costs to 126,000–189,000 per call, while intentionally retaining the cheap getRoyaltyStack selectors for internal stack operations in _getRoyaltyStackLAP/LRP. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6) `source_document_id=srcdoc_11286398b5cf997ee6d2913f90690722` `source_revision_id=srcrev_791f79a038a5c59ebf2bcb333ce65c58` `chunk_id=srcchunk_54acc0e89d2bdb080c7cc42fb136136b` `native_locator=https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6` `source_timestamp=2026-06-25T10:41:00Z`

## Sources

- `source_document_id`: `srcdoc_11286398b5cf997ee6d2913f90690722`
- `source_revision_id`: `srcrev_791f79a038a5c59ebf2bcb333ce65c58`
- `source_url`: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6)
