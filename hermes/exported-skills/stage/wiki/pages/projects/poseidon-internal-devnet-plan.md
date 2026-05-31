---
title: "Poseidon Internal Devnet Plan"
type: "project"
slug: "projects/poseidon-internal-devnet-plan"
freshness: "2025-10-31T00:53:00Z"
tags:
  - "devnet"
  - "poseidon"
  - "subnet"
  - "video-processing"
owners:
  - "Andy"
  - "Jdub"
  - "Jin"
  - "Kingter"
  - "Raul"
  - "Romain"
  - "Royce"
  - "Yao"
source_revision_ids:
  - "srcrev_820a27b1e06a1c3062f162e9e29a6539"
conflict_state: "none"
---

# Poseidon Internal Devnet Plan

## Summary

Plan for the Poseidon internal devnet to run the video processing pipeline end-to-end on the subnet, expose engineering issues, and experiment with new tech stacks like Rust, reth, and gRPC.

## Claims

- Goal 1: Run the video processing pipeline on the subnet to make sure the design works end-to-end. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_cad0f767dc3d96b8de5058f09c177058` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1` `source_timestamp=2025-10-31T00:53:00Z`
- Goal 2: Video files uploaded to subnet API and R2, and the processed file uploaded back to R2. Reward distributed to worker depending on validator votes. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_cad0f767dc3d96b8de5058f09c177058` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1` `source_timestamp=2025-10-31T00:53:00Z`
- Goal 3: Exposing any potential engineering issues through this POC. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_cad0f767dc3d96b8de5058f09c177058` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1` `source_timestamp=2025-10-31T00:53:00Z`
- Goal 4: Experiment with new tech stack like Rust for API, reth and GRPC. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_cad0f767dc3d96b8de5058f09c177058` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1` `source_timestamp=2025-10-31T00:53:00Z`
- Stretch Goal: Load test the pipeline by scaling the different components of the subnet. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_cad0f767dc3d96b8de5058f09c177058` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-1` `source_timestamp=2025-10-31T00:53:00Z`
- Infra checklist includes 3 subnet API instances (Rust gRPC server), 1 loadbalancer, 1 subnet DB (Postgres/RDS), subnet L2 with 1 sequencer node and 2 L2 RPC nodes, native bridge, stable L1 network, blockscout explorer, 2 worker for processing, 3 worker for validation, and Prometheus/Grafana dashboard. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_8dc8545653a688048e0eab80966d7f23` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2` `source_timestamp=2025-10-31T00:53:00Z`
- Team assignments: Jdub for overall subnet, L2 and worker software; Andy/Jin for infra (Tony for offchain infra, Jin for subnet L2 infra); Romain/Kingter for Subnet Contract; Royce/Yao for Testing; Raul for security check/audit preparation. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_8dc8545653a688048e0eab80966d7f23` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2` `source_timestamp=2025-10-31T00:53:00Z`
- Timeline Week of July 28: API Design and Subnet contract design/implementation. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_8dc8545653a688048e0eab80966d7f23` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2` `source_timestamp=2025-10-31T00:53:00Z`
- Timeline Week of August 4 Demo goals: Deploy sequencer and L2 RPC nodes on cloud, connect L2 sequencer with L1 devnet, subnet API instances + subnet DB; run 1 workflow through L2 contract and 1 worker software with 2 simple tasks; subnet API server with file upload to R2; prepare test plan and testing scripts. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_8dc8545653a688048e0eab80966d7f23` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2` `source_timestamp=2025-10-31T00:53:00Z`
- Timeline Week of August 11 Demo goals: Block Explorer on L2, Prometheus/Grafana metrics, load balancer for RPC and subnet API nodes, Cloudflare rate limiting; run video file workflow by uploading through subnet API with final output stored back to R2 using single workflow contract and all components on cloud; subnet API signup/login and get file status; start functional testing on subnet API. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_8dc8545653a688048e0eab80966d7f23` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2` `source_timestamp=2025-10-31T00:53:00Z`
- Post crunch time goals (til Sept 6): End-to-end integration for worker registration and running workflow, load testing for starting workflows, MVP Contracts to support epoch and rewards, finalize binary and contract versions and deploy. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2) `source_document_id=srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa` `source_revision_id=srcrev_820a27b1e06a1c3062f162e9e29a6539` `chunk_id=srcchunk_8dc8545653a688048e0eab80966d7f23` `native_locator=https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287#chunk-2` `source_timestamp=2025-10-31T00:53:00Z`

## Sources

- `source_document_id`: `srcdoc_1fa38e0fc1fc46d837deeec415ebe9fa`
- `source_revision_id`: `srcrev_820a27b1e06a1c3062f162e9e29a6539`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-internal-devnet-plan-23e051299a54807d9791e0ca305f2287)
