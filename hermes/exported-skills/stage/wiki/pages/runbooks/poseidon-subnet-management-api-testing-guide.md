---
title: "Poseidon Subnet Management API Testing Guide"
type: "runbook"
slug: "runbooks/poseidon-subnet-management-api-testing-guide"
freshness: "2025-10-15T20:01:00Z"
tags:
  - "api"
  - "poseidon"
  - "subnet-management"
  - "testing"
owners: []
source_revision_ids:
  - "srcrev_b5820f5ba41350c7dc24443490700d41"
conflict_state: "none"
---

# Poseidon Subnet Management API Testing Guide

## Summary

A runbook for testing the Poseidon Subnet Management API, covering quick start, endpoints, and example curl commands.

## Claims

- The Poseidon Subnet Management API base URL is http://localhost:8080/api/v1. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c) `source_document_id=srcdoc_efe7648b9d4a006570a2348e92489b11` `source_revision_id=srcrev_b5820f5ba41350c7dc24443490700d41` `chunk_id=srcchunk_eb57e7a5ff33e060180b4dd989837e5a` `native_locator=https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c` `source_timestamp=2025-10-15T20:01:00Z`
- The server can be started with `make run` and all tests run with `./test-api.sh`. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c) `source_document_id=srcdoc_efe7648b9d4a006570a2348e92489b11` `source_revision_id=srcrev_b5820f5ba41350c7dc24443490700d41` `chunk_id=srcchunk_eb57e7a5ff33e060180b4dd989837e5a` `native_locator=https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c` `source_timestamp=2025-10-15T20:01:00Z`
- The dashboard metrics endpoint returns active workers, total staked, workflow counts, success rates, and duration percentiles. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c) `source_document_id=srcdoc_efe7648b9d4a006570a2348e92489b11` `source_revision_id=srcrev_b5820f5ba41350c7dc24443490700d41` `chunk_id=srcchunk_eb57e7a5ff33e060180b4dd989837e5a` `native_locator=https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c` `source_timestamp=2025-10-15T20:01:00Z`
- Workflows can be listed with pagination, filtering by type, status, and time range, and sorted by startTime. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c) `source_document_id=srcdoc_efe7648b9d4a006570a2348e92489b11` `source_revision_id=srcrev_b5820f5ba41350c7dc24443490700d41` `chunk_id=srcchunk_eb57e7a5ff33e060180b4dd989837e5a` `native_locator=https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c` `source_timestamp=2025-10-15T20:01:00Z`
- Workflow detail includes full metadata, state transition history, activity summaries, and worker information. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c) `source_document_id=srcdoc_efe7648b9d4a006570a2348e92489b11` `source_revision_id=srcrev_b5820f5ba41350c7dc24443490700d41` `chunk_id=srcchunk_eb57e7a5ff33e060180b4dd989837e5a` `native_locator=https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c` `source_timestamp=2025-10-15T20:01:00Z`
- Activity detail returns activity ID, workflow ID, type, step index, status, input/output, worker, queue, timestamps, duration, and attempt number. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c) `source_document_id=srcdoc_efe7648b9d4a006570a2348e92489b11` `source_revision_id=srcrev_b5820f5ba41350c7dc24443490700d41` `chunk_id=srcchunk_eb57e7a5ff33e060180b4dd989837e5a` `native_locator=https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c` `source_timestamp=2025-10-15T20:01:00Z`
- Worker listing supports pagination, and worker detail includes status, stake, active tasks, recent workflows, heartbeat, and jailed status. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c) `source_document_id=srcdoc_efe7648b9d4a006570a2348e92489b11` `source_revision_id=srcrev_b5820f5ba41350c7dc24443490700d41` `chunk_id=srcchunk_eb57e7a5ff33e060180b4dd989837e5a` `native_locator=https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c` `source_timestamp=2025-10-15T20:01:00Z`
- Task queues can be listed and optionally include activities. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c) `source_document_id=srcdoc_efe7648b9d4a006570a2348e92489b11` `source_revision_id=srcrev_b5820f5ba41350c7dc24443490700d41` `chunk_id=srcchunk_eb57e7a5ff33e060180b4dd989837e5a` `native_locator=https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c` `source_timestamp=2025-10-15T20:01:00Z`

## Sources

- `source_document_id`: `srcdoc_efe7648b9d4a006570a2348e92489b11`
- `source_revision_id`: `srcrev_b5820f5ba41350c7dc24443490700d41`
- `source_url`: [Notion source](https://www.notion.so/Test-cmd-for-subnet-management-api-28d051299a54806a85c3dc3bc29bb76c)
