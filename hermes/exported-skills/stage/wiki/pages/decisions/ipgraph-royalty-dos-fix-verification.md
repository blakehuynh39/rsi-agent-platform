---
title: "IPGraph Royalty DoS Fix Verification"
type: "decision"
slug: "decisions/ipgraph-royalty-dos-fix-verification"
freshness: "2026-06-25T09:54:00Z"
tags:
  - "devnet0"
  - "DoS"
  - "fix"
  - "IPGraph"
  - "royalty"
  - "verification"
owners:
  - "Barry"
source_revision_ids:
  - "srcrev_1837bbeccc16f5d2262d2af22df663b1"
conflict_state: "none"
---

# IPGraph Royalty DoS Fix Verification

## Summary

Validation of protocol v1.3.3 fix for IPGraph royalty-read gas underpricing DoS.

## Claims

- The v1.3.3 fix was validated end-to-end on devnet0 with result PASS and high confidence. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_1837bbeccc16f5d2262d2af22df663b1` `chunk_id=srcchunk_8004ea86c20836075260239aad1862b1` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T09:54:00Z`
- Pre-fix, the attack stalled the chain (empty blocks, all txs blocked) on non-trivial graph shapes. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_1837bbeccc16f5d2262d2af22df663b1` `chunk_id=srcchunk_8004ea86c20836075260239aad1862b1` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T09:54:00Z`
- Post-fix, the attack mines as a harmless out-of-gas and throughput is unaffected. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_1837bbeccc16f5d2262d2af22df663b1` `chunk_id=srcchunk_8004ea86c20836075260239aad1862b1` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T09:54:00Z`
- All three attacker-reachable royalty reads have been fixed with per-call gas increases of 5.5–7.6×. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_1837bbeccc16f5d2262d2af22df663b1` `chunk_id=srcchunk_8004ea86c20836075260239aad1862b1` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T09:54:00Z`
- The upgrade was applied via the real 2-of-3 Safe → ProtocolAccessManager governance path. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_1837bbeccc16f5d2262d2af22df663b1` `chunk_id=srcchunk_8004ea86c20836075260239aad1862b1` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T09:54:00Z`
- The post-fix gas cost for getPolicyRoyalty on RoyaltyPolicyLAP is approximately 222,665 (flat). `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_1837bbeccc16f5d2262d2af22df663b1` `chunk_id=srcchunk_8ede09bd984452322bf547391690ebd8` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T09:54:00Z`
- The fix does not cover the theoretical worst-case graph but is safe in practice because the per-tx build cap makes that graph unbuildable. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_1837bbeccc16f5d2262d2af22df663b1` `chunk_id=srcchunk_8004ea86c20836075260239aad1862b1` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T09:54:00Z`
- The geth wall-clock watchdog (PR #37) should be retained as defense-in-depth. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_1837bbeccc16f5d2262d2af22df663b1` `chunk_id=srcchunk_8004ea86c20836075260239aad1862b1` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T09:54:00Z`

## Open Questions

- Confirm geth watchdog implementation and integration.
- Precisely determine the DoS threshold band for different hardware configurations.
- Verify the fix on Aeneid and mainnet.

## Sources

- `source_document_id`: `srcdoc_27fd4d8eda660512c3b65ba1dc9937a0`
- `source_revision_id`: `srcrev_1837bbeccc16f5d2262d2af22df663b1`
- `source_url`: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255)
