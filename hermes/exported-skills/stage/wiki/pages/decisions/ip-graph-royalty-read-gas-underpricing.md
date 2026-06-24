---
title: "IP Graph Royalty Read Gas Underpricing"
type: "decision"
slug: "decisions/ip-graph-royalty-read-gas-underpricing"
freshness: "2026-06-24T16:47:00Z"
tags:
  - "gas"
  - "precompile"
  - "royalty"
  - "security"
  - "story-geth"
owners: []
source_revision_ids:
  - "srcrev_f58cd235b0ca0e7db3f911ef6b669a8f"
conflict_state: "none"
---

# IP Graph Royalty Read Gas Underpricing

## Summary

A fix was introduced to address a gas underpricing vulnerability in IP graph royalty reads. Three precompile call sites were switched from cheap internal selectors to more expensive external selectors to mitigate an attack where an untrusted caller could force expensive graph traversals on every validating node for very low gas cost. The root cause lies in story-geth's gas pricing, which uses a flat constant based on average ancestor count, independent of actual graph size.

## Claims

- The fix is in commit 2c29808 on branch fix/ipgraph-royalty-gas-underpricing. It changes three precompile call sites from cheap/internal to external selectors: RoyaltyModule._hasAncestorIp → hasAncestorIpExt, RoyaltyPolicyLAP._getRoyaltyLAP → getRoyaltyExt, RoyaltyPolicyLRP._getRoyaltyLRP → getRoyaltyExt. It also removes the dead helper RoyaltyModule._getAncestorCount, and intentionally keeps the cheap getRoyaltyStack in _getRoyaltyStackLAP/LRP. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6) `source_document_id=srcdoc_11286398b5cf997ee6d2913f90690722` `source_revision_id=srcrev_f58cd235b0ca0e7db3f911ef6b669a8f` `chunk_id=srcchunk_9f9780fd31ed4f906e8ea46be8c1ba99` `native_locator=https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6` `source_timestamp=2026-06-24T16:47:00Z`
- The root cause is that story-geth's RequiredGas function in core/vm/ipgraph.go prices precompile gas based only on the 4-byte selector. Two tiers exist: ipGraphReadGas = 10 for cheap/internal selectors, and ipGraphExternalReadGas = 2100 for external selectors. Both dispatch to the same Go implementation that performs full DFS and topological sorts, with gas independent of actual graph size (protocol limits: MAX_ANCESTORS=1024, MAX_PARENTS=16). The external tier provides a 210x safety margin for untrusted callers. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6) `source_document_id=srcdoc_11286398b5cf997ee6d2913f90690722` `source_revision_id=srcrev_f58cd235b0ca0e7db3f911ef6b669a8f` `chunk_id=srcchunk_9f9780fd31ed4f906e8ea46be8c1ba99` `native_locator=https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6` `source_timestamp=2026-06-24T16:47:00Z`
- The attack targets public, unauthenticated functions in RoyaltyPolicyLAP/LRP and RoyaltyModule that accept arbitrary ipId/ancestorIpId. Using the cheap selectors, an attacker can loop calls that trigger full DFS/topological traversals on every validating node, consuming only ~600–900 gas per call by manipulating a deep ancestor graph approaching 1024 ancestors. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6) `source_document_id=srcdoc_11286398b5cf997ee6d2913f90690722` `source_revision_id=srcrev_f58cd235b0ca0e7db3f911ef6b669a8f` `chunk_id=srcchunk_9f9780fd31ed4f906e8ea46be8c1ba99` `native_locator=https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6` `source_timestamp=2026-06-24T16:47:00Z`

## Sources

- `source_document_id`: `srcdoc_11286398b5cf997ee6d2913f90690722`
- `source_revision_id`: `srcrev_f58cd235b0ca0e7db3f911ef6b669a8f`
- `source_url`: [source](https://app.notion.com/p/IP_GRAPH-royalty-read-gas-underpricing-auditor-brief-389051299a5480b58721db9cda02c4c6)
