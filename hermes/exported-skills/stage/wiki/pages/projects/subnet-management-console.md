---
title: "Subnet Management Console"
type: "project"
slug: "projects/subnet-management-console"
freshness: "2026-01-14T16:44:00Z"
tags:
  - "ethereum-l2"
  - "management-console"
  - "subnet"
  - "workflow"
owners: []
source_revision_ids:
  - "srcrev_a05faecd3a21324deeb2eb6b7a125b83"
conflict_state: "none"
---

# Subnet Management Console

## Summary

A management console for subnet operators to monitor and manage the state of workflows and activities on an Ethereum L2-based scheduling system. The console focuses on visualization and operational actions rather than workflow authoring.

## Claims

- The subnet team built a functional subnet using Ethereum L2 as the scheduling stack and control plane in August. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-1) `source_document_id=srcdoc_6d90945f54beda7ceaef3388fe9d1f86` `source_revision_id=srcrev_a05faecd3a21324deeb2eb6b7a125b83` `chunk_id=srcchunk_8cf8f456a2895a7f0ec8c4a40b8cf431` `native_locator=https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-1` `source_timestamp=2026-01-14T16:44:00Z`
- The management console is focused on managing rather than creating, with workflow authoring and pre-workflow creation state monitoring out of scope. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-1) `source_document_id=srcdoc_6d90945f54beda7ceaef3388fe9d1f86` `source_revision_id=srcrev_a05faecd3a21324deeb2eb6b7a125b83` `chunk_id=srcchunk_8cf8f456a2895a7f0ec8c4a40b8cf431` `native_locator=https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-1` `source_timestamp=2026-01-14T16:44:00Z`
- P0 features include workflow visualization (graph/timeline), activity monitoring, and task queue monitoring. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-1) `source_document_id=srcdoc_6d90945f54beda7ceaef3388fe9d1f86` `source_revision_id=srcrev_a05faecd3a21324deeb2eb6b7a125b83` `chunk_id=srcchunk_8cf8f456a2895a7f0ec8c4a40b8cf431` `native_locator=https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-1` `source_timestamp=2026-01-14T16:44:00Z`
- All data sources for the console should come from L2 and smart contracts, requiring an off-chain common indexing layer to collect, cache, and aggregate on-chain data. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-3) `source_document_id=srcdoc_6d90945f54beda7ceaef3388fe9d1f86` `source_revision_id=srcrev_a05faecd3a21324deeb2eb6b7a125b83` `chunk_id=srcchunk_2ef4a9e5b230d577eb87a59f926d4585` `native_locator=https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-3` `source_timestamp=2026-01-14T16:44:00Z`
- The indexer is intended to be open-sourced so users can run it separately to support other tools requiring on-chain workflow state data. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-3) `source_document_id=srcdoc_6d90945f54beda7ceaef3388fe9d1f86` `source_revision_id=srcrev_a05faecd3a21324deeb2eb6b7a125b83` `chunk_id=srcchunk_2ef4a9e5b230d577eb87a59f926d4585` `native_locator=https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-3` `source_timestamp=2026-01-14T16:44:00Z`
- The infrastructure setup includes VPC with public/private subnets, security groups for ALB/ECS/RDS, ECR repository, secrets management, RDS instance, and ECS service with FireLens log driver configured for Loki. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-3) `source_document_id=srcdoc_6d90945f54beda7ceaef3388fe9d1f86` `source_revision_id=srcrev_a05faecd3a21324deeb2eb6b7a125b83` `chunk_id=srcchunk_2ef4a9e5b230d577eb87a59f926d4585` `native_locator=https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-3` `source_timestamp=2026-01-14T16:44:00Z`
- The ECS task definition uses FireLens to ship logs to Loki, with support for basic auth and bearer token authentication. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-6) `source_document_id=srcdoc_6d90945f54beda7ceaef3388fe9d1f86` `source_revision_id=srcrev_a05faecd3a21324deeb2eb6b7a125b83` `chunk_id=srcchunk_c61f8f95bf0331468e05e9ab8c881c4a` `native_locator=https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-6` `source_timestamp=2026-01-14T16:44:00Z`
- The project uses a phased infrastructure script approach (01-vpc.sh through 06-ecs-service.sh) to provision AWS resources. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-4) `source_document_id=srcdoc_6d90945f54beda7ceaef3388fe9d1f86` `source_revision_id=srcrev_a05faecd3a21324deeb2eb6b7a125b83` `chunk_id=srcchunk_98000abc6aa6fc1c100d227adf4f3dab` `native_locator=https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba#chunk-4` `source_timestamp=2026-01-14T16:44:00Z`

## Open Questions

- What are the exact Loki host, port, and TLS settings?
- What is the database engine and version?
- What is the specific API_PORT value?
- Who are the assigned team members for Product, Frontend, Backend, Infra, and Test/Release?

## Related Pages

- `meeting-notes-subnet`

## Sources

- `source_document_id`: `srcdoc_6d90945f54beda7ceaef3388fe9d1f86`
- `source_revision_id`: `srcrev_a05faecd3a21324deeb2eb6b7a125b83`
- `source_url`: [Notion source](https://www.notion.so/PRD-Subnet-Management-Console-26e051299a5480f6a1a0df8dd4adfcba)
