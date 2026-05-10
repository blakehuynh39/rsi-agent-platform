---
title: "Odyssey Devnet Runbook"
type: "runbook"
slug: "runbooks/odyssey-devnet-runbook"
freshness: "2025-01-23T18:08:00Z"
tags:
  - "deprecated"
  - "infrastructure"
  - "odyssey-devnet"
  - "runbook"
owners: []
source_revision_ids:
  - "srcrev_fabd18f907e87935ee67e4801b4e5b8f"
conflict_state: "none"
---

# Odyssey Devnet Runbook

## Summary

Runbook for the deprecated Odyssey devnet, a pre-launch test network for the Odyssey network. Covers context, services, infrastructure details including node distribution and SSH access, port configurations, and future improvements.

## Claims

- The Odyssey devnet was created before the official launch of the Odyssey network and is infrastructure-wise extremely similar to the finalized design. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1) `source_document_id=srcdoc_ff8e29d9ccbd348cae838e175a4ea41d` `source_revision_id=srcrev_fabd18f907e87935ee67e4801b4e5b8f` `chunk_id=srcchunk_773c6bdfc43ea6e8b73cd15263064911` `native_locator=https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1` `source_timestamp=2025-01-23T18:08:00Z`
- The ChainID for Odyssey devnet is 1315. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1) `source_document_id=srcdoc_ff8e29d9ccbd348cae838e175a4ea41d` `source_revision_id=srcrev_fabd18f907e87935ee67e4801b4e5b8f` `chunk_id=srcchunk_773c6bdfc43ea6e8b73cd15263064911` `native_locator=https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1` `source_timestamp=2025-01-23T18:08:00Z`
- The Odyssey devnet services include an Explorer at https://devnet.storyscan.xyz/, an RPC Endpoint at https://devnet.storyrpc.io, Grafana monitoring at https://monitoring.devnet.storyprotocol.net, and Pagerduty at https://storyprotocol.pagerduty.com/. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1) `source_document_id=srcdoc_ff8e29d9ccbd348cae838e175a4ea41d` `source_revision_id=srcrev_fabd18f907e87935ee67e4801b4e5b8f` `chunk_id=srcchunk_773c6bdfc43ea6e8b73cd15263064911` `native_locator=https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1` `source_timestamp=2025-01-23T18:08:00Z`
- All nodes for odyssey-devnet are provisioned using shell scripts and tagged with Network: odyssey-devnet and a Role such as validator, bootnode, explorer, or rpc. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1) `source_document_id=srcdoc_ff8e29d9ccbd348cae838e175a4ea41d` `source_revision_id=srcrev_fabd18f907e87935ee67e4801b4e5b8f` `chunk_id=srcchunk_773c6bdfc43ea6e8b73cd15263064911` `native_locator=https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1` `source_timestamp=2025-01-23T18:08:00Z`
- The Odyssey devnet infrastructure includes 2 bootnodes, 2 RPC nodes, and 8 validators distributed across us-east-1, us-west-1, eu-west-1, and ap-southeast-1. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1) `source_document_id=srcdoc_ff8e29d9ccbd348cae838e175a4ea41d` `source_revision_id=srcrev_fabd18f907e87935ee67e4801b4e5b8f` `chunk_id=srcchunk_773c6bdfc43ea6e8b73cd15263064911` `native_locator=https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1` `source_timestamp=2025-01-23T18:08:00Z`
- SSH access to nodes uses specific .pem key files per region, such as 'odyssey-devnet-us-east-1.pem' for us-east-1 nodes. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1) `source_document_id=srcdoc_ff8e29d9ccbd348cae838e175a4ea41d` `source_revision_id=srcrev_fabd18f907e87935ee67e4801b4e5b8f` `chunk_id=srcchunk_773c6bdfc43ea6e8b73cd15263064911` `native_locator=https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-1` `source_timestamp=2025-01-23T18:08:00Z`
- Port 8546 (el-websocket) is open on validators, port 26656 (cl-p2p) is open on bootnodes, validators, and RPC nodes, and ports 6060 (el-metrics), 26660 (cl-metrics), and 9100 (node exporter) are open on all node types. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-2) `source_document_id=srcdoc_ff8e29d9ccbd348cae838e175a4ea41d` `source_revision_id=srcrev_fabd18f907e87935ee67e4801b4e5b8f` `chunk_id=srcchunk_7a728ff94c0f63095267394a7197d7b8` `native_locator=https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-2` `source_timestamp=2025-01-23T18:08:00Z`
- Future improvements include provisioning a bastion host for secure SSH access and tightening the security group for the RPC ALB. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-2) `source_document_id=srcdoc_ff8e29d9ccbd348cae838e175a4ea41d` `source_revision_id=srcrev_fabd18f907e87935ee67e4801b4e5b8f` `chunk_id=srcchunk_7a728ff94c0f63095267394a7197d7b8` `native_locator=https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87#chunk-2` `source_timestamp=2025-01-23T18:08:00Z`

## Sources

- `source_document_id`: `srcdoc_ff8e29d9ccbd348cae838e175a4ea41d`
- `source_revision_id`: `srcrev_fabd18f907e87935ee67e4801b4e5b8f`
- `source_url`: [Notion source](https://www.notion.so/deprecated-Odyssey-devnet-Runbook-31dfd9f90d0d4a0faf974d6b63481c87)
