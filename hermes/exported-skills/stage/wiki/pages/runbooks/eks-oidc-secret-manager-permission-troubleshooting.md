---
title: "EKS OIDC Secret Manager Permission Troubleshooting"
type: "runbook"
slug: "runbooks/eks-oidc-secret-manager-permission-troubleshooting"
freshness: "2023-11-30T05:06:00Z"
tags:
  - "eks"
  - "iam"
  - "oidc"
  - "secret-manager"
  - "troubleshooting"
owners: []
source_revision_ids:
  - "srcrev_e4a9f29d090d9a36d5d2afa3df0f8353"
conflict_state: "none"
---

# EKS OIDC Secret Manager Permission Troubleshooting

## Summary

Steps to diagnose and fix secret manager permission issues in EKS using OIDC, including verifying OIDC provider, service account role, trust relationship, and testing with a pod.

## Claims

- EKS uses OIDC to bind policies to KSAs. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- The OIDC provider URL for the cluster is https://oidc.eks.us-east-2.amazonaws.com/id/1A80314FFA2F9B3919D2F6938D75E61E. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- You can check the listed providers by running `aws iam list-open-id-connect-providers`. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- You can find the OIDC issuer configured for the cluster using `aws eks describe-cluster --name stag-story-eks-LzD98NGd --region us-east-2 --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///"`. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- Ensure the OIDC provider from AWS console matches the one from the cluster description. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- Check the role assumed by the service account using `kubectl describe serviceaccount default -n edge` and look at the annotation `eks.amazonaws.com/role-arn`. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- The role ARN for the edge default service account is arn:aws:iam::243963068353:role/stag-story-eks-LzD98NGd-edge-default. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- Define the needed policies (e.g., secret manager, S3 access) on the IAM role found in the service account annotation. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- The trust relationship must include the OIDC provider with audience sts.amazonaws.com and subject system:serviceaccount:edge:default. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- To test permissions, create a test pod with the service account and awscli, then run `kubectl exec -it awscli -n edge -- aws sts get-caller-identity` and `kubectl exec -it awscli -n edge -- aws secretsmanager get-secret-value --region us-east-2 --secret-id api`. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`
- Example test pod YAML: apiVersion: v1, kind: Pod, metadata: name: awscli, labels: app: awscli, spec: serviceAccountName: default, containers: - image: amazon/aws-cli, command: ["sleep", "604800"], imagePullPolicy: IfNotPresent, name: awscli, restartPolicy: Always. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c) `source_document_id=srcdoc_a30a7aa892d9c5d2f3b36353a88a255b` `source_revision_id=srcrev_e4a9f29d090d9a36d5d2afa3df0f8353` `chunk_id=srcchunk_1e3505b00878c5bdb1256cf417f59458` `native_locator=https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c` `source_timestamp=2023-11-30T05:06:00Z`

## Sources

- `source_document_id`: `srcdoc_a30a7aa892d9c5d2f3b36353a88a255b`
- `source_revision_id`: `srcrev_e4a9f29d090d9a36d5d2afa3df0f8353`
- `source_url`: [Notion source](https://www.notion.so/Nov-28-secret-manager-permission-issue-19ade2ecda5c45c088432e1cb658a37c)
