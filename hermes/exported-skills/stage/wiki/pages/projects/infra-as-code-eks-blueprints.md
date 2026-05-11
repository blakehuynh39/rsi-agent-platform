---
title: "Infrastructure as Code with AWS EKS Blueprints"
type: "project"
slug: "projects/infra-as-code-eks-blueprints"
freshness: "2024-10-28T02:00:00Z"
tags:
  - "aws"
  - "eks"
  - "infrastructure-as-code"
  - "kubernetes"
  - "terraform"
owners:
  - "user://2dfb900d-99bf-405c-b66a-f957d2e568d0"
source_revision_ids:
  - "srcrev_9f33e0d402b219198473c626f76eba44"
conflict_state: "none"
---

# Infrastructure as Code with AWS EKS Blueprints

## Summary

Project to provision and manage infrastructure using Terraform and AWS EKS Blueprints, covering EKS, ECR, and RDS services.

## Claims

- A functional development environment was provisioned in December 2022, allowing developers to deploy and update applications on an actual cluster. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-1) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_8c1b998af5663d40d8200262c0eb9be9` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-1` `source_timestamp=2024-10-28T02:00:00Z`
- User authentication and AWS load balancer deployment are labor-intensive and error-prone if repeated manually. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-1) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_8c1b998af5663d40d8200262c0eb9be9` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-1` `source_timestamp=2024-10-28T02:00:00Z`
- AWS EKS Blueprints is a relatively new open-source framework owned by AWS, designed to be more DevOps/SRE friendly for provisioning EKS clusters and critical add-ons. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-1) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_8c1b998af5663d40d8200262c0eb9be9` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-1` `source_timestamp=2024-10-28T02:00:00Z`
- The three major services to provision are Amazon Elastic Kubernetes Service (EKS), Elastic Container Registry (ECR), and Relational Database Service (RDS). Simple Storage Service (S3) may be provisioned later if needed. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-1) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_8c1b998af5663d40d8200262c0eb9be9` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-1` `source_timestamp=2024-10-28T02:00:00Z`
- The EKS architecture includes a master node (control plane) and worker nodes. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-2) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_205336322f70b5f8e1313f6edcd77519` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-2` `source_timestamp=2024-10-28T02:00:00Z`
- ECR is configured as a private repository with a lifecycle policy. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-2) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_205336322f70b5f8e1313f6edcd77519` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-2` `source_timestamp=2024-10-28T02:00:00Z`
- RDS configuration includes subnet groups (private subnet group for hosting the RDS instance), parameter groups, option groups, and the database itself. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-2) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_205336322f70b5f8e1313f6edcd77519` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-2` `source_timestamp=2024-10-28T02:00:00Z`
- The infrastructure code repository is located at https://github.com/storyprotocol/iac-pro. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-2) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_205336322f70b5f8e1313f6edcd77519` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-2` `source_timestamp=2024-10-28T02:00:00Z`
- The demo workflow involves destroying an existing staging environment, selecting the Terraform workspace, planning with a variable file, and applying the plan. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-3) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_60966643059ef39c649c8d901cfa28f5` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-3` `source_timestamp=2024-10-28T02:00:00Z`
- A known issue (GitHub issue #13) broke the deployment due to a previous commit to deploy ArgoCD in the cluster, and a fix is in progress. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-4) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_101d70cbe1d8e12a93e4d3868de4d2a2` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-4` `source_timestamp=2024-10-28T02:00:00Z`
- GitHub issue #19 references checking ArgoCD pod status using 'kubectl get po --namespace argocd'. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-4) `source_document_id=srcdoc_955519f6562cc75ed1df6eb936c0ec81` `source_revision_id=srcrev_9f33e0d402b219198473c626f76eba44` `chunk_id=srcchunk_101d70cbe1d8e12a93e4d3868de4d2a2` `native_locator=https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1#chunk-4` `source_timestamp=2024-10-28T02:00:00Z`

## Sources

- `source_document_id`: `srcdoc_955519f6562cc75ed1df6eb936c0ec81`
- `source_revision_id`: `srcrev_9f33e0d402b219198473c626f76eba44`
- `source_url`: [Notion source](https://www.notion.so/Manage-an-Infra-Using-Infra-As-Code-Design-Principles-a59a583df513406b837376fb89429fb1)
