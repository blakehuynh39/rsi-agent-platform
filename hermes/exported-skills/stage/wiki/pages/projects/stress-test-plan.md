---
title: "Stress Test Plan"
type: "project"
slug: "projects/stress-test-plan"
freshness: "2024-06-07T03:56:00Z"
tags:
  - "blockchain"
  - "performance"
  - "stress-test"
owners: []
source_revision_ids:
  - "srcrev_959a9c4800011d88dcdc877c0e757687"
conflict_state: "none"
---

# Stress Test Plan

## Summary

Plan to stress test the blockchain network using real data to find baseline performance and recommended node specs. Setup includes 2s block time, 50 c5.large nodes across regions, 10k validators, 1M delegators, parallel transactions, and various load tests. Metrics cover TPS, block time, resource usage, and a 3-hour full load run.

## Claims

- Stress test the blockchain network using real data. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Find the baseline of the performance. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Figure out the recommended node spec. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Block time set to 2 seconds. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- 50 nodes using c5.large instances. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Nodes distributed to different regions, or using half GCP half AWS. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Gas limit per block is configured (value unspecified). `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- 10,000 validators and 1,000,000 delegators. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Regular transactions use parallel execution. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Staking includes a withdraw stress test. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- IP graph consists of 10,000 nodes. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Swap load test measures number of swaps per second. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- RPC load test is included. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Metrics include transactions per second and swaps per second. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Block time is a tracked metric. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- CPU, memory, and disk load of c5.large nodes are measured. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Disk consumption rate is monitored. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- Bandwidth consumption is measured. `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`
- The test runs for 3 hours under full load. `claim:claim_1_19` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4) `source_document_id=srcdoc_60c564685e4c0ff3874b214d6b71f863` `source_revision_id=srcrev_959a9c4800011d88dcdc877c0e757687` `chunk_id=srcchunk_a52b78c2a050c2848e64caf2ae300045` `native_locator=https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4` `source_timestamp=2024-06-07T03:56:00Z`

## Sources

- `source_document_id`: `srcdoc_60c564685e4c0ff3874b214d6b71f863`
- `source_revision_id`: `srcrev_959a9c4800011d88dcdc877c0e757687`
- `source_url`: [Notion source](https://www.notion.so/Stress-test-plan-draft-5eee2530301e4e4aa8e12b9f89717bb4)
