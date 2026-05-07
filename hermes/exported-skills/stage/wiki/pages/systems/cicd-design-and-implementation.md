---
title: "CICD Design and Implementation"
type: "system"
slug: "systems/cicd-design-and-implementation"
freshness: "2024-10-23T01:16:00Z"
tags:
  - "argocd"
  - "aws-eks"
  - "cicd"
  - "continuous-delivery"
  - "continuous-integration"
  - "github-actions"
owners:
  - "user://2dfb900d-99bf-405c-b66a-f957d2e568d0"
source_revision_ids:
  - "srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1"
conflict_state: "none"
---

# CICD Design and Implementation

## Summary

Overview of the CI/CD pipeline for the project, using GitHub Actions for CI and ArgoCD for CD on AWS EKS.

## Claims

- GitHub Actions (GHA) was chosen for CI because it is affordable, with a basic account adequate for workload demand and extra minutes over quota costing $0.008 for a 2-core Linux VM. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-1) `source_document_id=srcdoc_48582cac2b6b26a70e99882c9b640ebe` `source_revision_id=srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1` `chunk_id=srcchunk_4d834493892fd89738df27ffb48472ab` `native_locator=https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-1` `source_timestamp=2024-10-23T01:16:00Z`
- GitHub Actions was also chosen for its customizability and flexibility, with 17,952 plugins available in the GitHub Marketplace at the time of documentation. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-1) `source_document_id=srcdoc_48582cac2b6b26a70e99882c9b640ebe` `source_revision_id=srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1` `chunk_id=srcchunk_4d834493892fd89738df27ffb48472ab` `native_locator=https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-1` `source_timestamp=2024-10-23T01:16:00Z`
- The CI workflow on pull requests includes: formatting and linting Go code, linting builder and API Dockerfiles, installing Go environment (≥1.19.0), building and tagging builder and API Docker images, and scanning image vulnerabilities. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-2) `source_document_id=srcdoc_48582cac2b6b26a70e99882c9b640ebe` `source_revision_id=srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1` `chunk_id=srcchunk_6be6378d974ef74ef21546a065dc30b8` `native_locator=https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-2` `source_timestamp=2024-10-23T01:16:00Z`
- When a PR is merged into the main branch, a GHA workflow named handle_push_secure.yaml executes additional jobs. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-2) `source_document_id=srcdoc_48582cac2b6b26a70e99882c9b640ebe` `source_revision_id=srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1` `chunk_id=srcchunk_6be6378d974ef74ef21546a065dc30b8` `native_locator=https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-2` `source_timestamp=2024-10-23T01:16:00Z`
- The PAT token created by Andy should be replaced with a token owned by Leeren or Leo, used to allow the GHA workflow to checkout code from the project-nova-cd repo and update values using Kustomize. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-3) `source_document_id=srcdoc_48582cac2b6b26a70e99882c9b640ebe` `source_revision_id=srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1` `chunk_id=srcchunk_dbf66bc8c834f6a19f39c4cc7302fa57` `native_locator=https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-3` `source_timestamp=2024-10-23T01:16:00Z`
- Do not approve any suspicious PR that attempts to modify files in .github/workflows; alternatively, enforce GHA triggers only on selected files. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-3) `source_document_id=srcdoc_48582cac2b6b26a70e99882c9b640ebe` `source_revision_id=srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1` `chunk_id=srcchunk_dbf66bc8c834f6a19f39c4cc7302fa57` `native_locator=https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-3` `source_timestamp=2024-10-23T01:16:00Z`
- ArgoCD is used for continuous delivery and has been deployed on the AWS EKS cluster via the AWS blueprints project. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-3) `source_document_id=srcdoc_48582cac2b6b26a70e99882c9b640ebe` `source_revision_id=srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1` `chunk_id=srcchunk_dbf66bc8c834f6a19f39c4cc7302fa57` `native_locator=https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-3` `source_timestamp=2024-10-23T01:16:00Z`
- ArgoCD staging environment is accessible at https://argocd-stag.storyprotocol.net/ and production at https://argocd-prod.storyprotocol.net/. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-3) `source_document_id=srcdoc_48582cac2b6b26a70e99882c9b640ebe` `source_revision_id=srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1` `chunk_id=srcchunk_dbf66bc8c834f6a19f39c4cc7302fa57` `native_locator=https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f#chunk-3` `source_timestamp=2024-10-23T01:16:00Z`

## Sources

- `source_document_id`: `srcdoc_48582cac2b6b26a70e99882c9b640ebe`
- `source_revision_id`: `srcrev_fb9c9a1eb22cc9e5e1daa20520e245b1`
- `source_url`: [Notion source](https://www.notion.so/KB-CICD-Design-and-Implementation-924ef723cb9a4790bebfbc0286aded8f)
