---
title: "Runbook: Provision a Production Environment"
type: "runbook"
slug: "runbooks/provision-prod-environment"
freshness: "2024-10-28T00:48:00Z"
tags:
  - "argocd"
  - "aws"
  - "eks"
  - "infrastructure"
  - "production"
  - "terraform"
owners:
  - "Andy Wu (andy@storyprotocol.xyz)"
source_revision_ids:
  - "srcrev_dcc30a4a1e25d004e73959724c2dd198"
conflict_state: "none"
---

# Runbook: Provision a Production Environment

## Summary

Step-by-step guide for provisioning a production environment in AWS us-east-1 using Terraform and EKS blueprints, including EKS cluster setup, add-on deployment, and ArgoCD configuration.

## Claims

- The production environment is provisioned in the AWS region us-east-1 using account 243963068353. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-1) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_72562e0007faac6af6cc1eecb8b4d7f7` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-1` `source_timestamp=2024-10-28T00:48:00Z`
- The infrastructure is managed using Terraform and AWS EKS blueprints, with the IaC repository at https://github.com/storyprotocol/iac-pro. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-1) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_72562e0007faac6af6cc1eecb8b4d7f7` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-1` `source_timestamp=2024-10-28T00:48:00Z`
- A complete functional environment includes an EKS K8S cluster, an AWS RDS (Postgres DB), an AWS ECR (two repos), and supporting AWS resources (VPC, EC2, load balancer, Route 53, API Gateway, Amazon Certificate Manager). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-1) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_72562e0007faac6af6cc1eecb8b4d7f7` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-1` `source_timestamp=2024-10-28T00:48:00Z`
- The EKS cluster must be provisioned first before running Terraform scripts for EKS Kubernetes add-ons. The add-ons configuration file `eks_blueprints_kubernetes_addons.tf` should be commented out during the initial cluster provisioning. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-6) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_be29f80fb663bbff1e621d42f6b05262` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-6` `source_timestamp=2024-10-28T00:48:00Z`
- After the EKS cluster is created, update the local kubeconfig using the command: `aws eks --region us-east-1 update-kubeconfig --name prod-story-eks-fPKR0qZ1`. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-6) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_be29f80fb663bbff1e621d42f6b05262` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-6` `source_timestamp=2024-10-28T00:48:00Z`
- To enable the ArgoCD Web UI, patch the service to use a LoadBalancer: `kubectl patch svc argo-cd-argocd-server -n argocd -p '{"spec": {"type": "LoadBalancer"}}'`. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-7) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_66d52d5171aae99b0009c2a523dae18b` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-7` `source_timestamp=2024-10-28T00:48:00Z`
- Two secrets must be manually added before executing Terraform scripts: `argocd-ssh` for fetching changes from the private project-nova-cd repo, and `argocd-admin2` for the preset admin password for ArgoCD access. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-9) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_5cf856750730123e9b8e558e1c10f9ab` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-9` `source_timestamp=2024-10-28T00:48:00Z`
- An additional secret `api` is required for API application logic. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-9) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_5cf856750730123e9b8e558e1c10f9ab` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-9` `source_timestamp=2024-10-28T00:48:00Z`
- The RDS daily backup policy should be automated. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-9) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_5cf856750730123e9b8e558e1c10f9ab` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-9` `source_timestamp=2024-10-28T00:48:00Z`
- The document was authored by Andy Wu (andy@storyprotocol.xyz) on February 27, 2023. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-1) `source_document_id=srcdoc_b10ae949d6777d8b799d074fb2cfc11e` `source_revision_id=srcrev_dcc30a4a1e25d004e73959724c2dd198` `chunk_id=srcchunk_72562e0007faac6af6cc1eecb8b4d7f7` `native_locator=https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff#chunk-1` `source_timestamp=2024-10-28T00:48:00Z`

## Related Pages

- `system/argocd-deployment`
- `system/eks-cluster-prod`
- `system/iac-pro-repository`

## Sources

- `source_document_id`: `srcdoc_b10ae949d6777d8b799d074fb2cfc11e`
- `source_revision_id`: `srcrev_dcc30a4a1e25d004e73959724c2dd198`
- `source_url`: [Notion source](https://www.notion.so/KB-Provision-a-Prod-Env-97f9c63780744dfab8207e4ba8345fff)
