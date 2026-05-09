---
title: "Kubernetes Setup and Operations"
type: "runbook"
slug: "runbooks/kubernetes-setup-and-operations"
freshness: "2024-10-23T01:15:00Z"
tags:
  - "authentication"
  - "aws"
  - "ebs-csi-driver"
  - "eks"
  - "kubernetes"
  - "macos"
owners: []
source_revision_ids:
  - "srcrev_1325d97588b2abe5c0e8009c5a4534bb"
conflict_state: "none"
---

# Kubernetes Setup and Operations

## Summary

Runbook for setting up kubectl access to EKS clusters on macOS, troubleshooting common authentication issues, and configuring EBS CSI driver for persistent volumes.

## Claims

- On macOS, install aws-iam-authenticator using Homebrew: brew install aws-iam-authenticator. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_9de3f278b550c657c554b90afa769c0c` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1` `source_timestamp=2024-10-23T01:15:00Z`
- Configure kubectl to use aws-iam-authenticator for EKS cluster authentication by running `ca $CLUSTER_NAME` (e.g., prod-story-eks-fPKR0qZ1). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_9de3f278b550c657c554b90afa769c0c` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1` `source_timestamp=2024-10-23T01:15:00Z`
- Verify successful setup by running `kubectl get nodes`. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_9de3f278b550c657c554b90afa769c0c` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1` `source_timestamp=2024-10-23T01:15:00Z`
- If you encounter 'Error when retrieving token from sso: Token has expired and refresh failed', refresh the SSO token by running `aws sso login --profile $PROFILE` (e.g., admin). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_9de3f278b550c657c554b90afa769c0c` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1` `source_timestamp=2024-10-23T01:15:00Z`
- The AWS SSO profile configuration for the admin profile uses sso_start_url https://storyprotocol.awsapps.com/start/#, sso_region us-east-1, sso_account_id 243963068353, and sso_role_name AdministratorAccess. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_9de3f278b550c657c554b90afa769c0c` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1` `source_timestamp=2024-10-23T01:15:00Z`
- If you get 'Error from server (Forbidden): nodes is forbidden' or similar forbidden errors, the impersonating IAM role lacks sufficient Kubernetes permissions; a cluster admin must create a ClusterRoleBinding granting cluster-admin to the IAM role ARN. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_9de3f278b550c657c554b90afa769c0c` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1` `source_timestamp=2024-10-23T01:15:00Z`
- After creating the EKS cluster, configure the node pool. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_9de3f278b550c657c554b90afa769c0c` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-1` `source_timestamp=2024-10-23T01:15:00Z`
- Install the EBS CSI Driver to enable creation of PersistentVolumeClaims (PVCs). `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-2) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_049a70749e194cd32684cc9b03ea0802` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-2` `source_timestamp=2024-10-23T01:15:00Z`
- Create an IAM role named AmazonEKS_EBS_CSI_DriverRole with a trust policy that allows the EBS CSI controller service account to assume the role via OIDC. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-2) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_049a70749e194cd32684cc9b03ea0802` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-2` `source_timestamp=2024-10-23T01:15:00Z`
- Attach the AWS managed policy AmazonEBSCSIDriverPolicy to the IAM role AmazonEKS_EBS_CSI_DriverRole. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-2) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_049a70749e194cd32684cc9b03ea0802` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-2` `source_timestamp=2024-10-23T01:15:00Z`
- Node groups should have the three policies attached (as shown in the referenced screenshot). `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-3) `source_document_id=srcdoc_13c48c093614e8587ffd28cac3d3973f` `source_revision_id=srcrev_1325d97588b2abe5c0e8009c5a4534bb` `chunk_id=srcchunk_aa25e9cad0a0611aee91db83fa40e37e` `native_locator=https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d#chunk-3` `source_timestamp=2024-10-23T01:15:00Z`

## Sources

- `source_document_id`: `srcdoc_13c48c093614e8587ffd28cac3d3973f`
- `source_revision_id`: `srcrev_1325d97588b2abe5c0e8009c5a4534bb`
- `source_url`: [Notion source](https://www.notion.so/k8s-Ops-7a146fb0b29a47938784c6aeaf1c089d)
