---
title: "Infra Proposal for a Multi-tenant System"
type: "decision"
slug: "decisions/infra-multi-tenant-aws-org-proposal"
freshness: "2023-03-30T02:36:00Z"
tags:
  - "aws"
  - "infrastructure"
  - "multi-tenant"
  - "organization"
  - "sso"
owners:
  - "Andy Wu"
source_revision_ids:
  - "srcrev_9ba3d5abe055eef850faf3b78acbe950"
conflict_state: "none"
---

# Infra Proposal for a Multi-tenant System

## Summary

Proposal to restructure AWS accounts under a centralized AWS Organization with dedicated OUs for STAG, PROD, and ECR, enabling environment isolation and unified identity management via AWS SSO with Google Workspace.

## Claims

- Currently, staging (us-east-2) and production (us-east-1) environments are managed under a single AWS account. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`
- The proposal aims to create a centralized AWS Organization to manage all AWS accounts. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`
- Three recommended Organization Units (OUs) are STAG, PROD, and ECR, each potentially containing one or more AWS accounts. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`
- The existing AWS account (243963068353) should be moved to the PROD OU to avoid migration risks. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`
- A new AWS account should be created under the STAG OU for the staging environment. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`
- A dedicated AWS account under the ECR OU should be created to store all container images (dev/stag/prod) for future releases. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`
- Unified identity management should be implemented using AWS SSO with an external identity provider, likely Google Workspace via SAML 2.0. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`
- Kubernetes namespaces are not strictly necessary for different franchises but may be a good practice for grouping dedicated resources; the decision is delegated to a specific reviewer. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`
- A single RDS instance is considered adequate for the multi-tenant design, as no sensitive information is stored. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471) `source_document_id=srcdoc_04e871e6cbf6a83166767e56a0a504fb` `source_revision_id=srcrev_9ba3d5abe055eef850faf3b78acbe950` `chunk_id=srcchunk_f61b17816b64be408e94500f48035a88` `native_locator=https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471` `source_timestamp=2023-03-30T02:36:00Z`

## Open Questions

- Ensure all services/roles in a cluster can pull images from different accounts.
- How to manage service roles (e.g., GitHub Action role): manually or via Infrastructure as Code (IaC)?
- Infrastructure provision and deployment required for the newly created STAG environment.

## Sources

- `source_document_id`: `srcdoc_04e871e6cbf6a83166767e56a0a504fb`
- `source_revision_id`: `srcrev_9ba3d5abe055eef850faf3b78acbe950`
- `source_url`: [Notion source](https://www.notion.so/Infra-Proposal-for-a-Multi-tenant-System-627f722f907a4a22ae122a625e179471)
