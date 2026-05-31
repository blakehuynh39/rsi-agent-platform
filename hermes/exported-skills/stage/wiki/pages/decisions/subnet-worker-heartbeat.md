---
title: "Worker Heartbeat Signal Design"
type: "decision"
slug: "decisions/subnet-worker-heartbeat"
freshness: "2025-08-27T22:58:00Z"
tags:
  - "heartbeat"
  - "liveness"
  - "subnet"
  - "worker"
owners: []
source_revision_ids:
  - "srcrev_1fc1b382d19ed23fe13d7ee37d5541e8"
conflict_state: "none"
---

# Worker Heartbeat Signal Design

## Summary

Decision that worker heartbeat should be a global liveness signal rather than per-activity, since it only indicates whether a worker is alive.

## Claims

- Worker heartbeat should be a global signal instead of per activity, since it only tells if a worker is alive. `claim:claim_3_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429) `source_document_id=srcdoc_3110f64bf66eeee994c86d3bed55469e` `source_revision_id=srcrev_1fc1b382d19ed23fe13d7ee37d5541e8` `chunk_id=srcchunk_2fb5602212ad4e56c203cdfee2c410c1` `native_locator=https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429` `source_timestamp=2025-08-27T22:58:00Z`

## Related Pages

- `subnet-stale-task-detection`
- `subnet-worker-concurrency`

## Sources

- `source_document_id`: `srcdoc_3110f64bf66eeee994c86d3bed55469e`
- `source_revision_id`: `srcrev_1fc1b382d19ed23fe13d7ee37d5541e8`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429)
