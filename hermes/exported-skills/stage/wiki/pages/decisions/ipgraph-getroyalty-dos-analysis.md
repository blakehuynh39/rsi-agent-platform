---
title: "ipgraph getRoyalty DoS — Worst-Case Analysis"
type: "decision"
slug: "decisions/ipgraph-getroyalty-dos-analysis"
freshness: "2026-06-25T10:58:00Z"
tags:
  - "analysis"
  - "DoS"
  - "gas"
  - "ipgraph"
  - "security"
owners:
  - "security-team"
source_revision_ids:
  - "srcrev_4c0db62d69767cf622a55117d4d5e52f"
conflict_state: "none"
---

# ipgraph getRoyalty DoS — Worst-Case Analysis

## Summary

The ipgraph getRoyalty function is vulnerable to a Denial-of-Service (DoS) attack due to fixed gas cost while performing unbounded StateDB reads. Under current contract limits (≤1024 ancestors, ≤16 parents), a single transaction can perform ~4.3 million reads, taking ~13 seconds on mainnet, far exceeding the 2-second block time. This prevents fillTransactions from completing, leading to sustained empty blocks. The graph costs ~$0.05 to build and the attack tx pays nothing. Recommendations include capping ancestors to 16 and forcing external callers to use Ext variants.

## Claims

- The attack is still feasible under the current contract limits (≤1024 ancestors, ≤16 parents). `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- A single getRoyaltyExt transaction over a dense 1024-ancestor DAG performs ~4.3M StateDB reads (LRP), taking ~13s on mainnet, which keeps fillTransactions from completing and produces sustained empty blocks. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- The graph costs ~$0.05 to build once (reusable), and the trigger tx pays nothing (it never lands). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- The current limits cap a single call but not a single transaction (88–128 calls fit), so they do not prevent the DoS. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- Recommendation: cap ancestors ≤16 (parent may stay 16), and force external callers onto the Ext variants. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- getRoyalty/getRoyaltyExt charge a fixed gas (RequiredGas, derived from a hardcoded average ancestor count), while the actual work — topologicalSort over the queried node's ancestor closure — is unbounded and reads StateDB per node and per edge. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- Reads per call: 1 + V + 2E (1 ACL check, V node parent-counts, E parent pointers in topologicalSort, E edge-royalties read in distribution loop). `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- Per-call reads are identical between LAP and LRP, but LRP's fixed price is lower, allowing more calls per transaction. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- Worst-case DAG: V=1024, E = Σ min(16, r-1) = 16,248, resulting in 33,520 reads/call. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- Per transaction: LAP allows 88 calls and 2,949,760 reads; LRP allows 128 calls and 4,290,560 reads. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`
- Localnet measurement: full attack tx performed 2,815,680 reads in 1.452s. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8) `source_document_id=srcdoc_7d3e5745ad658305b0fff3e461c774c3` `source_revision_id=srcrev_4c0db62d69767cf622a55117d4d5e52f` `chunk_id=srcchunk_0ab2ceb12a0a409eef7b70fd2f1cea13` `native_locator=https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8` `source_timestamp=2026-06-25T10:58:00Z`

## Open Questions

- Has the recommendation (cap ancestors ≤16, force Ext variants) been implemented?

## Sources

- `source_document_id`: `srcdoc_7d3e5745ad658305b0fff3e461c774c3`
- `source_revision_id`: `srcrev_4c0db62d69767cf622a55117d4d5e52f`
- `source_url`: [source](https://app.notion.com/p/ipgraph-getRoyalty-DoS-Worst-Case-Analysis-38a051299a548024a9e9df7393f702c8)
