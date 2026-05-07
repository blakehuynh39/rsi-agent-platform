---
title: "Poseidon Subnet as a Rollup-based Layer2"
type: "system"
slug: "systems/poseidon-subnet-rollup-layer2"
freshness: "2025-10-31T20:35:00Z"
tags:
  - "layer2"
  - "op-stack"
  - "poseidon"
  - "smart-contracts"
  - "subnet"
  - "workflow-engine"
owners: []
source_revision_ids:
  - "srcrev_f558338bebabf01ad2675536cae71c1c"
conflict_state: "none"
---

# Poseidon Subnet as a Rollup-based Layer2

## Summary

High-level design for a scalable, programmable Subnet infrastructure where the Subnet Server is a Layer 2 chain built on OP Stack, supporting workflow execution as smart contracts, decentralized task queues, and off-chain workers.

## Claims

- The Subnet Server is implemented as a Layer 2 chain built on OP Stack. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`
- The Workflow Engine is a smart contract that implements orchestration logic. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`
- Activities are stateless execution units executed by workers off-chain, with results updated on-chain via the Workflow Engine contract; they are designed to be idempotent and retry-safe. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`
- The Task Queue is a decentralized, smart-contract-based queue that stores pending activity metadata; workers poll tasks by querying the Task Queue contract via Layer 2 RPC. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`
- The Subnet Control Plane smart contract handles epoch-based subnet membership, worker staking/slashing, worker selection for tasks, reward distribution, and reputation tracking. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`
- The Subnet Miner is an off-chain worker that polls tasks from the workflow task queue, executes activities (e.g., data processing, AI inference), and submits results on-chain via RPC to Layer 2. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`
- The Submit Validator is a special off-chain worker that polls a validation_task_queue, executes validation logic, and submits a vote/attestation to the Control Plane. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`
- The Subnet Platform Service is an off-chain API gateway for DePin and other applications, handling subnet registration/discovery, workflow request orchestration, and communication with the Subnet Server (Layer 2 chain). `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`
- The design is inspired by Temporal’s workflow model, mapping similar principles into an on-chain system using smart contracts. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d) `source_document_id=srcdoc_d32ebab57027083d934781cbf3472d30` `source_revision_id=srcrev_f558338bebabf01ad2675536cae71c1c` `chunk_id=srcchunk_49c8f1cf5327ae2ac42ebd90a2a7050e` `native_locator=https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d` `source_timestamp=2025-10-31T20:35:00Z`

## Open Questions

- What is the complete end-to-end flow for the Poseidon Subnet? The source chunk is truncated.

## Sources

- `source_document_id`: `srcdoc_d32ebab57027083d934781cbf3472d30`
- `source_revision_id`: `srcrev_f558338bebabf01ad2675536cae71c1c`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-Subnet-as-a-Rollup-based-Layer2-High-Level-Design-239051299a54805aa741d73d3e0e3e3d)
