---
title: "Stale Task Detection Mechanisms"
type: "decision"
slug: "decisions/subnet-stale-task-detection"
freshness: "2025-08-27T22:58:00Z"
tags:
  - "keepers"
  - "scheduling"
  - "stale-tasks"
  - "subnet"
  - "worker-management"
owners: []
source_revision_ids:
  - "srcrev_1fc1b382d19ed23fe13d7ee37d5541e8"
conflict_state: "none"
---

# Stale Task Detection Mechanisms

## Summary

Proposed mechanisms for detecting stale tasks in the subnet scheduling system, including round-robin jail detection and community keeper incentives.

## Claims

- To detect stale jobs, two mechanisms are needed: one for round-robin scheduling and one for non-round-robin or large-scale systems. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429) `source_document_id=srcdoc_3110f64bf66eeee994c86d3bed55469e` `source_revision_id=srcrev_1fc1b382d19ed23fe13d7ee37d5541e8` `chunk_id=srcchunk_2fb5602212ad4e56c203cdfee2c410c1` `native_locator=https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429` `source_timestamp=2025-08-27T22:58:00Z`
- If the scheduling algorithm is round-robin, the scheduler will jail the worker with a stale job and reassign the task when it next attempts to schedule to that worker. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429) `source_document_id=srcdoc_3110f64bf66eeee994c86d3bed55469e` `source_revision_id=srcrev_1fc1b382d19ed23fe13d7ee37d5541e8` `chunk_id=srcchunk_2fb5602212ad4e56c203cdfee2c410c1` `native_locator=https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429` `source_timestamp=2025-08-27T22:58:00Z`
- For non-round-robin scheduling or systems with many workers and tasks, communities (keepers) can be leveraged to detect stale tasks by providing small incentives. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429) `source_document_id=srcdoc_3110f64bf66eeee994c86d3bed55469e` `source_revision_id=srcrev_1fc1b382d19ed23fe13d7ee37d5541e8` `chunk_id=srcchunk_2fb5602212ad4e56c203cdfee2c410c1` `native_locator=https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429` `source_timestamp=2025-08-27T22:58:00Z`
- Using keepers for stale task detection introduces downsides including potential collusion between workers and keepers, and increased spamming risks from opening up the RPC. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429) `source_document_id=srcdoc_3110f64bf66eeee994c86d3bed55469e` `source_revision_id=srcrev_1fc1b382d19ed23fe13d7ee37d5541e8` `chunk_id=srcchunk_2fb5602212ad4e56c203cdfee2c410c1` `native_locator=https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429` `source_timestamp=2025-08-27T22:58:00Z`

## Open Questions

- How to mitigate collusion risks between workers and keepers?
- How to secure the RPC against spam when opened for keeper detection?

## Related Pages

- `subnet-worker-concurrency`
- `subnet-worker-heartbeat`

## Sources

- `source_document_id`: `srcdoc_3110f64bf66eeee994c86d3bed55469e`
- `source_revision_id`: `srcrev_1fc1b382d19ed23fe13d7ee37d5541e8`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429)
