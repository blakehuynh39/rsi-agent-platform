---
title: "100M Gas Data Registration Benchmarks"
type: "concept"
slug: "concepts/100m-gas-data-registration-benchmarks"
freshness: "2026-05-22T19:06:00Z"
tags:
  - "benchmark"
  - "data-registry"
  - "gas-limit"
  - "registration-contract"
  - "sha-256"
  - "smart-contract"
  - "throughput"
owners: []
source_revision_ids:
  - "srcrev_2a81c79f5a6c27602e968ba9f5c9fe16"
conflict_state: "none"
---

# 100M Gas Data Registration Benchmarks

## Summary

Performance benchmarks for the data registration contract (single and batch10 workloads) under a 100M EL gas limit, measuring transaction throughput, storage writes per second, and block-level resource usage.

## Claims

- For single workload (register), max tx/s is 230.6 with 200 wallets and target_cps=600, achieving 99.41% delivery. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- For batch10 workload (registerBatch), max tx/s is 73.0 with 200 wallets and target_cps=800, achieving 99.13% delivery; max records/s is 730.2. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- Maximum SSTORE writes per second is 730.2, achieved with the batch10 workload (10 SSTORE per tx). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- For batch10 at peak (target_cps=800), mean block time is 3.27s, mean block size is 701.6 KB, max block size is 829.6 KB, and max gas used is 87.83% of 100M. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- With the single workload, maximum block gas usage reaches 99.78% at all wallet counts tested (50–300 wallets). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- For the single workload with 300 wallets, the mean gasUsed per block is 295.9 KB, while the max gasUsed per block is 611.9 KB (99.78% of 100M). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- In the batch10 workload, block time remains steady at 3.0s up to target_cps=400 (45% gas utilization), rising to 3.27s at target_cps=800 (88% gas utilization), indicating chain pushback only at high load. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- The workload is gas-bound, not size-bound: the max observed block size is 830 KB, which is only 4% of the 20 MB CL cap. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- Test setup used 3 validators + 1 RPC + 1 submitter (m6i.4xlarge), all in us-east-1a, running Story v1.7.0 and Story-Geth v1.2.0. EL gas limit was set to 100M via genesis.gasLimit and miner.gaslimit. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_35350dc332f82e20f4a2108f523a27b9` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-1` `source_timestamp=2026-05-22T19:06:00Z`
- Planned follow-up tests include 10-minute sustained validation, WAN cross-region re-test, contract sharding (K=4), and a reproducible report PR. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-2) `source_document_id=srcdoc_f137db0ea908d30b088c350bdd7f6aa4` `source_revision_id=srcrev_2a81c79f5a6c27602e968ba9f5c9fe16` `chunk_id=srcchunk_1682858d86867f5d1c32d3fd2b119b29` `native_locator=https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741#chunk-2` `source_timestamp=2026-05-22T19:06:00Z`

## Open Questions

- Are there further improvements possible in batch10 throughput beyond 73.0 tx/s?
- How does contract sharding (K=4) affect the observed gas-bound behavior?
- How does the chain behave under sustained 10-minute peak loads?
- What are the throughput implications of moving to a WAN cross-region topology?

## Sources

- `source_document_id`: `srcdoc_f137db0ea908d30b088c350bdd7f6aa4`
- `source_revision_id`: `srcrev_2a81c79f5a6c27602e968ba9f5c9fe16`
- `source_url`: [Notion source](https://www.notion.so/100M-Gas-Data-Registration-Testing-368051299a5480169e6dca7f5df4d741)
