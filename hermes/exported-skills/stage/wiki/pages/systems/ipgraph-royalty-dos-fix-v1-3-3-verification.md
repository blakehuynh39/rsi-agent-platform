---
title: "IPGraph Royalty DoS v1.3.3 Fix Verification"
type: "system"
slug: "systems/ipgraph-royalty-dos-fix-v1-3-3-verification"
freshness: "2026-06-25T14:50:00Z"
tags:
  - "devnet0"
  - "DoS"
  - "IPGraph"
  - "royalty"
  - "security"
  - "v1.3.3"
owners:
  - "Barry (QA)"
source_revision_ids:
  - "srcrev_e30b7280293e21f8cd8429d7aba813ee"
conflict_state: "none"
---

# IPGraph Royalty DoS v1.3.3 Fix Verification

## Summary

End-to-end verification of the v1.3.3 fix for the IPGraph royalty-read gas underpricing DoS on devnet0. The fix is a partial mitigation: it bounds sparse/shallow graphs but does NOT prevent the dense worst-case graph from stalling the chain. Full mitigation requires an ancestors cap (≤16) and retention of the geth watchdog PR #37.

## Claims

- The v1.3.3 fix is a partial mitigation, not a full fix for the IPGraph royalty-read DoS. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_6313ced2672530b6d23554ed60529ad4` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T14:50:00Z`
- Pre-fix, the DoS was reproduced on 4/5 graph shapes (including the mainnet weaponized shape), stalling the chain with empty blocks and blocking all transactions. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_6313ced2672530b6d23554ed60529ad4` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T14:50:00Z`
- Post-fix, every sparse/shallow graph tested is bounded; the 1024-ancestor wide tree (mainnet shape) mines as a harmless out-of-gas (OOG) transaction, and the chain maintains full throughput. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_6313ced2672530b6d23554ed60529ad4` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T14:50:00Z`
- The dense worst-case graph (16-parent / 1024-ancestor DAG, ~33k StateDB reads/call, ~2.5M reads/tx) is buildable and still DoSes the chain post-fix, stably reproduced in 4/4 rounds (sustained empty blocks). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_6313ced2672530b6d23554ed60529ad4` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T14:50:00Z`
- The fix caps the cost of a single call but not a single transaction (75–88 calls still fit). Full mitigation additionally requires an ancestors cap ≤16; gas fix alone is insufficient, and the geth watchdog (PR #37) should be kept. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_6313ced2672530b6d23554ed60529ad4` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T14:50:00Z`
- The partial‑mitigation finding was confirmed on both exposed royalty reads—LAP and LRP (LRP packs ~102 calls/tx vs LAP ~75, making it slightly worse). `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_6313ced2672530b6d23554ed60529ad4` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T14:50:00Z`
- Theoretical worst‑case traversal: V+E = 17,408 per call. Pre‑fix 575 calls/tx → ~10.0M traversals/tx (33–67× over DoS threshold). Post‑fix 75 calls/tx → ~1.31M traversals/tx (4.4–8.7× over threshold if the graph were buildable). `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_c4e19e835a14f180d48b79c81716f203` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T14:50:00Z`
- The DoS threshold is estimated at 150k–300k traversal units per 16M‑gas transaction, derived from empirical results (tip-64 74k mined, apex-975 post-fix 146k mined, lattice-183 299k DoS). `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_c4e19e835a14f180d48b79c81716f203` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T14:50:00Z`
- To fully cover the theoretical worst case by gas alone, getPolicyRoyalty would need ≥1.0M gas per call; the fix achieved 222k gas/call (75 calls/tx), leaving a gap of 4–9×. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_c4e19e835a14f180d48b79c81716f203` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T14:50:00Z`
- Practical safety: the per‑transaction 16M gas cap on registerDerivative makes the V+E=17,408 graph unbuildable; the densest buildable graph is ~2,000–3,000 V+E, giving post‑fix ~190k traversals/tx (near the threshold, but bounded). `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_c4e19e835a14f180d48b79c81716f203` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T14:50:00Z`
- Building the worst‑case dense graph via the protocol costs ~5.26 billion gas (461 transactions, ~30–45 min), ~131 IP on devnet at 25 gwei, and is a one‑time reusable investment for the attacker. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_c4e19e835a14f180d48b79c81716f203` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T14:50:00Z`
- The previously estimated $0.05 build cost was incorrect; an attacker must use registerDerivative (~5M gas per node), not the ACL‑gated addParentIp. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_c4e19e835a14f180d48b79c81716f203` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T14:50:00Z`
- The v1.3.3 fix was deployed via the actual 2-of-3 Safe → ProtocolAccessManager schedule‑execute governance path on devnet0. `claim:claim_1_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_6313ced2672530b6d23554ed60529ad4` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T14:50:00Z`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-3) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_96341417d57cb7651f7b87dbcdf7a045` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-3` `source_timestamp=2026-06-25T14:50:00Z`
- Evidence of pre‑fix DoS includes empty blocks (e.g., block 31150 with 0 transactions) during attack execution. `claim:claim_1_14` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-3) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_e30b7280293e21f8cd8429d7aba813ee` `chunk_id=srcchunk_96341417d57cb7651f7b87dbcdf7a045` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-3` `source_timestamp=2026-06-25T14:50:00Z`

## Open Questions

- Can the geth watchdog (PR #37) be merged and activated before mainnet deployment to provide defense‑in‑depth?
- Does the ancestor cap ≤16 introduce any protocol‑level compatibility issues or affect legitimate use‑cases?
- Is there a feasible gas‑pricing mechanism that fully covers the dense worst case without breaking other functionality?

## Related Pages

- `geth-watchdog-pr-37`
- `ipgraph-royalty-dos-initial-report`

## Sources

- `source_document_id`: `srcdoc_27fd4d8eda660512c3b65ba1dc9937a0`
- `source_revision_id`: `srcrev_e30b7280293e21f8cd8429d7aba813ee`
- `source_url`: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255)
