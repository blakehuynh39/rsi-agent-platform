---
title: "Incident: Prod Cluster Down due to ArgoCD Misconfig (Mar 16, 2023)"
type: "runbook"
slug: "runbooks/incident-2023-03-16-prod-cluster-down-argocd-misconfig"
freshness: "2026-05-05T06:38:44Z"
tags:
  - "argocd"
  - "incident"
  - "kubernetes"
  - "production"
  - "terraform"
owners: []
source_revision_ids:
  - "srcrev_533216a3c994c6145867347d63515fb8"
conflict_state: "none"
---

# Incident: Prod Cluster Down due to ArgoCD Misconfig (Mar 16, 2023)

## Summary

On March 16, 2023, the production API cluster went down due to an ArgoCD misconfiguration. A Terraform change intended for staging was applied while the local kubectl context was pointing to the production cluster, causing ArgoCD to pull staging images into production. The issue was detected by UptimeRobot at 11:00 PM and resolved by 11:16 PM by correcting the ArgoCD deployment path.

## Claims

- A Terraform change was applied to the staging environment and failed in the middle. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1) `source_document_id=srcdoc_b41ea634949f533868011debdcd6b92f` `source_revision_id=srcrev_533216a3c994c6145867347d63515fb8` `chunk_id=srcchunk_a50c93a4200273ba92a60d7d572ac9cf` `native_locator=https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1` `source_timestamp=2026-05-05T06:38:44Z`
- UptimeRobot alerted that the production API was down at 11:00 PM via Slack and PagerDuty. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1) `source_document_id=srcdoc_b41ea634949f533868011debdcd6b92f` `source_revision_id=srcrev_533216a3c994c6145867347d63515fb8` `chunk_id=srcchunk_a50c93a4200273ba92a60d7d572ac9cf` `native_locator=https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1` `source_timestamp=2026-05-05T06:38:44Z`
- Investigation in the ArgoCD UI revealed that production servers were using staging images. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1) `source_document_id=srcdoc_b41ea634949f533868011debdcd6b92f` `source_revision_id=srcrev_533216a3c994c6145867347d63515fb8` `chunk_id=srcchunk_a50c93a4200273ba92a60d7d572ac9cf` `native_locator=https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1` `source_timestamp=2026-05-05T06:38:44Z`
- The production ArgoCD was configured to pull images from the GitHub staging folder instead of the production folder. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1) `source_document_id=srcdoc_b41ea634949f533868011debdcd6b92f` `source_revision_id=srcrev_533216a3c994c6145867347d63515fb8` `chunk_id=srcchunk_a50c93a4200273ba92a60d7d572ac9cf` `native_locator=https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1` `source_timestamp=2026-05-05T06:38:44Z`
- At 11:16 PM, the ArgoCD configuration was changed to pull from the production folder and synced, restoring the cluster to a healthy state. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1) `source_document_id=srcdoc_b41ea634949f533868011debdcd6b92f` `source_revision_id=srcrev_533216a3c994c6145867347d63515fb8` `chunk_id=srcchunk_a50c93a4200273ba92a60d7d572ac9cf` `native_locator=https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-1` `source_timestamp=2026-05-05T06:38:44Z`
- The root cause was that the Terraform workspace was switched from 'stag' to 'prod' while the local kubectl context still pointed to the production cluster, causing the ArgoCD deployment path to be changed unexpectedly. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-2) `source_document_id=srcdoc_b41ea634949f533868011debdcd6b92f` `source_revision_id=srcrev_533216a3c994c6145867347d63515fb8` `chunk_id=srcchunk_10ee87a8ed6824829643347d242c116b` `native_locator=https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea#chunk-2` `source_timestamp=2026-05-05T06:38:44Z`

## Open Questions

- What safeguards can prevent kubectl context mismatch in the future?
- What specific Terraform change was being applied?
- Why did the Terraform apply fail in the middle?

## Sources

- `source_document_id`: `srcdoc_b41ea634949f533868011debdcd6b92f`
- `source_revision_id`: `srcrev_533216a3c994c6145867347d63515fb8`
- `source_url`: [Notion source](https://www.notion.so/Mar-16-Prod-cluster-down-due-to-ArgoCD-misconfig-1d7f6bf669c846ada9e70f87ef73f3ea)
