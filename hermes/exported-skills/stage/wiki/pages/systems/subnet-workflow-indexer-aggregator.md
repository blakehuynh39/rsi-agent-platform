---
title: "Subnet Workflow Indexer \u0026 Aggregator"
type: "system"
slug: "systems/subnet-workflow-indexer-aggregator"
freshness: "2026-01-14T16:44:00Z"
tags:
  - "aggregation"
  - "blockchain"
  - "indexing"
  - "subnet"
  - "workflow"
owners: []
source_revision_ids:
  - "srcrev_c40b945d1874c878998142df0213c1d5"
conflict_state: "none"
---

# Subnet Workflow Indexer & Aggregator

## Summary

A scalable indexing and aggregation framework for workflow management systems on subnets. It tracks workflow execution, task processing, and activity lifecycle events from workflow engine contracts. The system consists of an Indexer that extracts raw events and stores them in per-job tables, and an Aggregator that processes those events into workflow metrics and analytics. It is configuration-driven, horizontally scalable via Redis locking, supports multiple subnets, and uses Rust-based handler functions for aggregation logic.

## Claims

- The Subnet Workflow Indexer & Aggregator is a scalable indexing and aggregation framework for workflow management systems on subnets, tracking workflow execution, task processing, and activity lifecycle events from workflow engine contracts. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_12b3f2a996870158b394d8d1bbdc417a` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1` `source_timestamp=2026-01-14T16:44:00Z`
- The system has two main components: an Indexer that extracts raw events from workflow contracts and stores them in per-job tables, and an Aggregator that processes raw events into workflow metrics and analytics. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_12b3f2a996870158b394d8d1bbdc417a` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1` `source_timestamp=2026-01-14T16:44:00Z`
- Key features include configuration-driven setup with YAML and database configs, horizontal scalability via Redis locking, multi-subnet support, auto-generated tables per index job, and Rust-based handler functions for aggregation logic. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_12b3f2a996870158b394d8d1bbdc417a` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1` `source_timestamp=2026-01-14T16:44:00Z`
- The architecture consists of Subnet Workflow Contracts emitting events, Index Workers that fetch logs via RPC, decode with ABI, auto-create tables, and store raw events in PostgreSQL, followed by Aggregator Workers that resolve source tables, invoke Rust handler functions, and update progress. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_12b3f2a996870158b394d8d1bbdc417a` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1` `source_timestamp=2026-01-14T16:44:00Z`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-3) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_3c67f5ed70fa09b3c20e139174976ce8` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-3` `source_timestamp=2026-01-14T16:44:00Z`
- Indexer workers acquire a Redis lock with key pattern "index:{config_id}" and a TTL of 300 seconds to ensure only one worker processes a given config at a time. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-2) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_fda22b33e727a7f2cd5fca3df77fd13b` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-2` `source_timestamp=2026-01-14T16:44:00Z`
- Aggregator workers acquire a Redis lock with key pattern "agg:{config_id}" and a TTL of 300 seconds. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-3) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_3c67f5ed70fa09b3c20e139174976ce8` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-3` `source_timestamp=2026-01-14T16:44:00Z`
- Raw event tables are auto-generated with the naming convention event_{subnet_name}_{config_name}_{event_name}, for example event_poseidon_workflow_engine_workflow_started. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-2) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_fda22b33e727a7f2cd5fca3df77fd13b` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-2` `source_timestamp=2026-01-14T16:44:00Z`
- The database schema includes tables index_configs, index_jobs, aggregation_configs, aggregation_jobs, and the auto-generated event tables. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-2) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_fda22b33e727a7f2cd5fca3df77fd13b` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-2` `source_timestamp=2026-01-14T16:44:00Z`
- Aggregator handler functions are Rust-based and are resolved by name from a registry; they receive database connection, source tables, block range, and configuration, and typically perform JOINs, GROUP BY, and UPSERT operations to produce metrics. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-3) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_3c67f5ed70fa09b3c20e139174976ce8` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-3` `source_timestamp=2026-01-14T16:44:00Z`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_12b3f2a996870158b394d8d1bbdc417a` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-1` `source_timestamp=2026-01-14T16:44:00Z`
- An example setup involves inserting index jobs for WorkflowStarted and WorkflowCompleted events on subnet_a, which causes the indexer to create tables like event_workflow_engine_created_workflowcreated and event_workflow_engine_completed_workflowcompleted, and an aggregation job with handler workflow_metrics that joins these tables to compute daily metrics. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-4) `source_document_id=srcdoc_d096029b732a1a749be248779737a33e` `source_revision_id=srcrev_c40b945d1874c878998142df0213c1d5` `chunk_id=srcchunk_2c294060ce58518b66185d2acf7e51de` `native_locator=https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8#chunk-4` `source_timestamp=2026-01-14T16:44:00Z`

## Sources

- `source_document_id`: `srcdoc_d096029b732a1a749be248779737a33e`
- `source_revision_id`: `srcrev_c40b945d1874c878998142df0213c1d5`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Workflow-Indexer-Aggregator-286051299a5480329d6af37a19d0edf8)
