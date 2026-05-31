---
title: "Worker Concurrent Job Limits"
type: "decision"
slug: "decisions/subnet-worker-concurrency"
freshness: "2025-08-27T22:58:00Z"
tags:
  - "concurrency"
  - "subnet"
  - "throughput"
  - "worker"
owners: []
source_revision_ids:
  - "srcrev_1fc1b382d19ed23fe13d7ee37d5541e8"
conflict_state: "none"
---

# Worker Concurrent Job Limits

## Summary

Proposal to allow workers to specify how many concurrent jobs they can handle, with an upper bound limit, to increase system throughput.

## Claims

- To increase system throughput, workers should be allowed to specify how many jobs they want to take concurrently, with an upper bound limit. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429) `source_document_id=srcdoc_3110f64bf66eeee994c86d3bed55469e` `source_revision_id=srcrev_1fc1b382d19ed23fe13d7ee37d5541e8` `chunk_id=srcchunk_2fb5602212ad4e56c203cdfee2c410c1` `native_locator=https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429` `source_timestamp=2025-08-27T22:58:00Z`

## Open Questions

- What should the upper bound limit be for concurrent jobs per worker?

## Related Pages

- `subnet-stale-task-detection`
- `subnet-worker-heartbeat`

## Sources

- `source_document_id`: `srcdoc_3110f64bf66eeee994c86d3bed55469e`
- `source_revision_id`: `srcrev_1fc1b382d19ed23fe13d7ee37d5541e8`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Weekly-25c051299a5480a3b5a4e6c9c42b6429)
