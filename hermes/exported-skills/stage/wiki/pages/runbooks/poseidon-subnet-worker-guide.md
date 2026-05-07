---
title: "Poseidon Subnet Worker Guide"
type: "runbook"
slug: "runbooks/poseidon-subnet-worker-guide"
freshness: "2025-09-05T00:50:00Z"
tags:
  - "guide"
  - "poseidon"
  - "subnet"
  - "worker"
owners: []
source_revision_ids:
  - "srcrev_599aafd1c547e848814661a836c161e9"
conflict_state: "none"
---

# Poseidon Subnet Worker Guide

## Summary

Guide for workers in the Poseidon Subnet workflow engine, covering registration, activity processing, staking, and rewards.

## Claims

- Workers need at least 100 PSDN tokens to register. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- The SubnetTreasury needs approval to transfer the worker's stake. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- Registration is performed by calling registerWorker(uint256) on the SubnetControlPlane contract with a minimum stake of 100 PSDN. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- Workers can verify registration by calling getWorkerInfo(address) and isWorkerActive(address) on the SubnetControlPlane. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- Workers poll for activities in queues such as default_queue, validation_queue, video/1.0.0/processing, and video/1.0.0/validation. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- Polling an activity returns an activity struct and a boolean indicating availability. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- Activities are claimed by calling claimActivity(bytes32) on the TaskQueue contract. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- Claimed activities expire after a timeout that varies by activity type; workers must complete or fail the activity before expiration. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- Multiple workers may claim validation activities, which are consensus-based. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- After claiming, workers retrieve activity details via getActivity(bytes32) on the TaskQueue, decode inputData, perform off-chain work, and submit results. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_4e3051d9bd9fd5f242cc9d45594e9b75` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-1` `source_timestamp=2025-09-05T00:50:00Z`
- Workers can unregister and withdraw stake after a delay period by calling withdrawStake(address) on the SubnetControlPlane. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-2) `source_document_id=srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219` `source_revision_id=srcrev_599aafd1c547e848814661a836c161e9` `chunk_id=srcchunk_16990a75b9bc2637ea3bba0729495c35` `native_locator=https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5#chunk-2` `source_timestamp=2025-09-05T00:50:00Z`

## Sources

- `source_document_id`: `srcdoc_9c9153df3cf07e0ec3b11dd3cdff9219`
- `source_revision_id`: `srcrev_599aafd1c547e848814661a836c161e9`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Worker-Guide-263051299a5480fbb642e1f83e38d3c5)
