---
title: "IPGraph Royalty DoS Vulnerability and v1.3.3 Fix Verification"
type: "decision"
slug: "decisions/ipgraph-royalty-dos-v1-3-3-fix-verification"
freshness: "2026-06-25T12:32:00Z"
tags:
  - "fix-verification"
  - "ipgraph"
  - "royalty-dos"
  - "security"
  - "v1.3.3"
owners:
  - "Barry (QA)"
source_revision_ids:
  - "srcrev_24e472ba523d561a3d99967caa3d6387"
conflict_state: "none"
---

# IPGraph Royalty DoS Vulnerability and v1.3.3 Fix Verification

## Summary

Test report verifying the v1.3.3 fix for IPGraph royalty-related DoS vulnerability on devnet0. The fix bounds sparse/shallow graphs but does not fully cover dense worst-case graphs; full mitigation requires ancestor cap ≤ 16 and geth watchdog PR #37.

## Claims

- Pre-fix, the DoS attack stalled the chain on all non-trivial graph shapes. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_261047c73190e9e0997f22a99bcdab03` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:32:00Z`
- Post-fix, every sparse/shallow graph tested is bounded, including the 1024-ancestor wide tree (mainnet shape); the attack mines as a harmless OOG. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_261047c73190e9e0997f22a99bcdab03` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:32:00Z`
- The dense worst-case graph (16-parent / 1024-ancestor DAG) is buildable and still DoSes the chain post-fix, stably reproduced 4/4 rounds. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_261047c73190e9e0997f22a99bcdab03` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:32:00Z`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_de24c52e47b3e3b5656b145297ce719c` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:32:00Z`
- Full mitigation needs the contract cap on ancestors ≤ 16 in addition to the gas fix; the geth watchdog PR #37 should be kept. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_261047c73190e9e0997f22a99bcdab03` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:32:00Z`
- The v1.3.3 fix was applied via the real 2-of-3 Safe → ProtocolAccessManager governance path on devnet0. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_261047c73190e9e0997f22a99bcdab03` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:32:00Z`
- The fix routed all 3 attacker-reachable royalty reads to the expensive precompile selector getRoyaltyExt, increasing per-call gas by 5.5–7.6×. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_261047c73190e9e0997f22a99bcdab03` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:32:00Z`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_de24c52e47b3e3b5656b145297ce719c` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:32:00Z`
- Theoretical maximum traversals per transaction: pre‑fix ~10.0M; post‑fix ~1.31M. The DoS threshold is estimated at 150k–300k, so post‑fix remains 4.4–8.7× over if the graph is buildable. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_de24c52e47b3e3b5656b145297ce719c` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:32:00Z`
- Building the dense worst-case graph costs ~5.26 billion gas (~131 IP on devnet0), making it a viable one-time expenditure for an attacker. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_de24c52e47b3e3b5656b145297ce719c` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:32:00Z`
- Test recovery used nonce replacement from a controlled staller key; a real attacker’s staller cannot be replaced by validators, so actual mitigation requires the fix plus the geth watchdog. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_de24c52e47b3e3b5656b145297ce719c` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:32:00Z`
- Testing was performed on devnet0 (chainId 1512, RPC https://devnet0.storyrpc.io, explorer https://devnet0.storyscan.io). `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_24e472ba523d561a3d99967caa3d6387` `chunk_id=srcchunk_261047c73190e9e0997f22a99bcdab03` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:32:00Z`

## Open Questions

- Whether the borderline post‑fix case (densest buildable graph) can be empirically forced to confirm the exact threshold.

## Sources

- `source_document_id`: `srcdoc_27fd4d8eda660512c3b65ba1dc9937a0`
- `source_revision_id`: `srcrev_24e472ba523d561a3d99967caa3d6387`
- `source_url`: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255)
