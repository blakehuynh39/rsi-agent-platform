---
title: "IPGraph Royalty DoS v1.3.3 Fix Verification Test (devnet0)"
type: "runbook"
slug: "runbooks/ipgraph-royalty-dos-fix-verification-test"
freshness: "2026-06-25T12:19:00Z"
tags:
  - "devnet0"
  - "DoS"
  - "IPGraph"
  - "royalty"
  - "security"
owners:
  - "Barry (QA)"
source_revision_ids:
  - "srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849"
conflict_state: "none"
---

# IPGraph Royalty DoS v1.3.3 Fix Verification Test (devnet0)

## Summary

Verification test on devnet0 of the v1.3.3 protocol fix for the IPGraph royalty-read gas underpricing DoS. The fix mitigates the attack for all buildable graph shapes, but a theoretical dense worst-case graph remains a DoS vector. The geth wall-clock watchdog (PR #37) is recommended as defense-in-depth.

## Claims

- The test was performed by Barry on 2026-06-25 on devnet0 (chainId 1512) against the IPGraph royalty-read gas underpricing DoS issue (Hans, 2026-06-21). `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849` `chunk_id=srcchunk_deb0c7824d2a54b5fda4181b18383d96` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:19:00Z`
- Before the fix, the DoS attack stalls the chain (empty blocks, all txs blocked) on 4/5 graph shapes including the mainnet weaponized shape (wide tree with up to 1024 ancestors). After the fix, the attack is bounded on all tested sparse/shallow graphs and mines as a harmless OOG. The fix covers all three attacker-reachable royalty reads and was applied via the real 2-of-3 Safe → ProtocolAccessManager governance path. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849` `chunk_id=srcchunk_deb0c7824d2a54b5fda4181b18383d96` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:19:00Z`
- The dense worst-case graph (16 parents, 1024 ancestors per node, DAG with ~33k StateDB reads per call) is buildable on devnet0 and still DoSes the chain post-fix (stably reproduced 4/4 rounds). Full mitigation requires a contract-level cap of ancestors ≤ 16, as the gas fix alone is insufficient. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849` `chunk_id=srcchunk_deb0c7824d2a54b5fda4181b18383d96` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-1` `source_timestamp=2026-06-25T12:19:00Z`
- Theoretical traversal per call is V+E = 17,408 (1024 ancestors + 16*1024 parents). Pre-fix total traversal per tx ~10M, post-fix ~1.31M. DoS threshold estimated at 150k–300k units. Post-fix traversal is 4.4–8.7× over threshold, meaning a theoretical worst-case graph would still DoS. However, the per-tx cap on registerDerivative makes such a graph unbuildable; the densest buildable graph has V+E ~2,000–3,000, which post-fix yields ~190k traversal (near threshold), bounded. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849` `chunk_id=srcchunk_61a44f15c29dfb4577abae7b35df504a` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:19:00Z`
- The geth wall-clock watchdog (PR #37) should be kept as defense-in-depth. The contract fix alone is necessary and effective for real (buildable) attacks but does not cover the theoretical worst case. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849` `chunk_id=srcchunk_61a44f15c29dfb4577abae7b35df504a` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:19:00Z`
- Recovery in tests used replacement of the staller transaction with a higher-gas tx because the tester controlled the key. A real attacker's staller cannot be replaced by validators; real mitigation requires the fix + geth watchdog. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849` `chunk_id=srcchunk_61a44f15c29dfb4577abae7b35df504a` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:19:00Z`
- devnet0's geth reproduced the stall without the watchdog; the contract fix is the mitigation validated here. The DoS threshold and buildable-max V+E are estimates, not exact, and the borderline post-fix case was not empirically forced. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2) `source_document_id=srcdoc_27fd4d8eda660512c3b65ba1dc9937a0` `source_revision_id=srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849` `chunk_id=srcchunk_61a44f15c29dfb4577abae7b35df504a` `native_locator=https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255#chunk-2` `source_timestamp=2026-06-25T12:19:00Z`

## Open Questions

- Is the buildable-max V+E exactly 2,000–3,000, and can a more dense graph be constructed within the 16M gas limit?
- What is the exact DoS traversal threshold? Currently estimated as 150k–300k units based on empirical results, but needs precise calibration.
- Will the geth watchdog PR #37 be deployed on all nodes before mainnet exposure?

## Related Pages

- `geth-watchdog-pr-37`
- `ipgraph-royalty-dos-incident`
- `protocol-v1-3-3-upgrade`

## Sources

- `source_document_id`: `srcdoc_27fd4d8eda660512c3b65ba1dc9937a0`
- `source_revision_id`: `srcrev_09ceaa7cd0e3b3b2b2926a5b76f93849`
- `source_url`: [source](https://app.notion.com/p/Test-Report-IPGraph-royalty-DoS-v1-3-3-fix-verification-devnet0-38a051299a548112b38ce686f3c51255)
