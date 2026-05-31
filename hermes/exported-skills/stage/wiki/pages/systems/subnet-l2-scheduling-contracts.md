---
title: "Subnet L2 \u0026 Scheduling Contracts"
type: "system"
slug: "systems/subnet-l2-scheduling-contracts"
freshness: "2025-08-11T19:46:00Z"
tags:
  - "epoch"
  - "rollup"
  - "scheduling"
  - "smart-contracts"
  - "subnet"
owners: []
source_revision_ids:
  - "srcrev_330965e2d8d8fe14d7c3ef25328e5d88"
conflict_state: "none"
---

# Subnet L2 & Scheduling Contracts

## Summary

On-chain scheduling system for a rollup-based subnet that manages job queues, worker registration, staking, and reward distribution using an epoch-driven push model.

## Claims

- The system uses a push-based model where the subnet owner pushes jobs to on-chain queues. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- The subnet owner can assign a job to a specific registered worker or make it first-come, first-served. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- The Scheduler contract (JobScheduler.sol) stores queues of jobs for workers to take. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- A Job struct includes fields such as id, jobType (0 for processing, 1 for validation), jobVersion, assignee, status, initialPoints, currentPoints, createdBlock, createdEpoch, takenEpoch, metadata, and dataUri. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- Job metadata (off-chain) includes MediaPath, UserID, MediaType, and UploadDestination. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- The Scheduler uses OpenZeppelin's DoubleEndedQueue for processing and validation queues, and EnumerableSet for tracking taken, completed, and failed jobs. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- The take_job function requires the caller to be an active worker (checked via worker_manager.is_active) before assigning the job. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- The post_job_result function requires the caller to be an active worker assigned to the job; on success it rewards the worker, on failure it punishes the worker. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- Epoch advancement checks that the current time is past the epoch duration, then calls worker_manager.advance_epoch, checks worker points, distributes rewards for the last epoch, sets the target for the current epoch, and increments the epoch counter. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- The Configuration contract (JobConfig.sol) stores a mapping from subnetId and version to metadata bytes, and only the subnet owner can add a new version. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- The WorkerManager contract (WorkerManager.sol) manages worker stakes, slashing, and rewarding. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`
- The rollup proof of concept is located at https://github.com/PSDN-AI/subnet-poc/tree/main/rollup-based. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85) `source_document_id=srcdoc_d9db9461f01d9b274263d6c0506e508f` `source_revision_id=srcrev_330965e2d8d8fe14d7c3ef25328e5d88` `chunk_id=srcchunk_f5a683484fdfab7479ddf6333d6ddf7f` `native_locator=https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85` `source_timestamp=2025-08-11T19:46:00Z`

## Sources

- `source_document_id`: `srcdoc_d9db9461f01d9b274263d6c0506e508f`
- `source_revision_id`: `srcrev_330965e2d8d8fe14d7c3ef25328e5d88`
- `source_url`: [Notion source](https://www.notion.so/Subnet-L2-Scheduling-Contracts-23f051299a5480d58e35e087c000bb85)
