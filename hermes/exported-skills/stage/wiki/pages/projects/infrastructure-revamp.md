---
title: "Infrastructure Revamp"
type: "project"
slug: "projects/infrastructure-revamp"
freshness: "2023-10-30T15:17:00Z"
tags:
  - "aws"
  - "ci-cd"
  - "iac"
  - "infrastructure"
  - "kubernetes"
  - "monitoring"
  - "terraform"
owners: []
source_revision_ids:
  - "srcrev_ccf4550b1d1a0712fc01c8726c1f8959"
conflict_state: "none"
---

# Infrastructure Revamp

## Summary

Project to revamp Story Protocol's core infrastructure, migrating from a single AWS account with two single-ingress EKS clusters to a multi-account setup with improved IaC, CI/CD, monitoring, and API management.

## Claims

- As of October 2023, Story Protocol's infrastructure was mainly contained under a single AWS account, comprised of two single-ingress EKS clusters utilizing a rudimentary CI/CD process composed of Github Actions and ArgoCD. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-1) `source_document_id=srcdoc_8d87513469f76f255bf161b3476b08a7` `source_revision_id=srcrev_ccf4550b1d1a0712fc01c8726c1f8959` `chunk_id=srcchunk_407f3fe951fb71392cb8f83fbf1dc0c2` `native_locator=https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-1` `source_timestamp=2023-10-30T15:17:00Z`
- The infrastructure revamp started in October 2023 with the goal of building a more robust platform for app & infra deployments, API routing, monitoring & alerting, network optimization, IAM, and fully packaging infra and app codebase under a cohesive IaC repository. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-1) `source_document_id=srcdoc_8d87513469f76f255bf161b3476b08a7` `source_revision_id=srcrev_ccf4550b1d1a0712fc01c8726c1f8959` `chunk_id=srcchunk_407f3fe951fb71392cb8f83fbf1dc0c2` `native_locator=https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-1` `source_timestamp=2023-10-30T15:17:00Z`
- High-level goals for the initial infrastructure revamp include: IaC migration (P0), documentation (P0), monitoring & alerting (P2), cluster management (P1), and API layer creation (P0). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-1) `source_document_id=srcdoc_8d87513469f76f255bf161b3476b08a7` `source_revision_id=srcrev_ccf4550b1d1a0712fc01c8726c1f8959` `chunk_id=srcchunk_407f3fe951fb71392cb8f83fbf1dc0c2` `native_locator=https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-1` `source_timestamp=2023-10-30T15:17:00Z`
- Phase 1 (current as of Oct 23, 2023) includes provisioning VPC (completed), setting up S3, RDS, and other storage, and planning for API rate-limiting and L7 proxying tools. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-2) `source_document_id=srcdoc_8d87513469f76f255bf161b3476b08a7` `source_revision_id=srcrev_ccf4550b1d1a0712fc01c8726c1f8959` `chunk_id=srcchunk_6fe2e0416bd2ec19e17fe861247c8c90` `native_locator=https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-2` `source_timestamp=2023-10-30T15:17:00Z`
- Phase 2 (TBD) will set up ArgoCD in a ci-cd account in single-manager, multi-cluster mode, with ECR in the ci-cd account, and GH actions integration for staging CI/CD. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-2) `source_document_id=srcdoc_8d87513469f76f255bf161b3476b08a7` `source_revision_id=srcrev_ccf4550b1d1a0712fc01c8726c1f8959` `chunk_id=srcchunk_6fe2e0416bd2ec19e17fe861247c8c90` `native_locator=https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-2` `source_timestamp=2023-10-30T15:17:00Z`
- The API infrastructure layer will support multi-cluster ingress, IP rate-limiting, internal API routing, API policy management, and API versioning for the launch of Emergence and Alpha protocol APIs. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-1) `source_document_id=srcdoc_8d87513469f76f255bf161b3476b08a7` `source_revision_id=srcrev_ccf4550b1d1a0712fc01c8726c1f8959` `chunk_id=srcchunk_407f3fe951fb71392cb8f83fbf1dc0c2` `native_locator=https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf#chunk-1` `source_timestamp=2023-10-30T15:17:00Z`

## Sources

- `source_document_id`: `srcdoc_8d87513469f76f255bf161b3476b08a7`
- `source_revision_id`: `srcrev_ccf4550b1d1a0712fc01c8726c1f8959`
- `source_url`: [Notion source](https://www.notion.so/Infrastructure-Setup-Doc-0d2224e33c9149f699eb46a7123ebedf)
