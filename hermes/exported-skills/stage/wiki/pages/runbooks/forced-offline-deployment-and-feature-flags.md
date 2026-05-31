---
title: "Forced Offline Deployment \u0026 Feature Flags"
type: "runbook"
slug: "runbooks/forced-offline-deployment-and-feature-flags"
freshness: "2024-12-19T02:36:00Z"
tags:
  - "deployment"
  - "feature-flags"
  - "forced-offline"
  - "vercel"
owners: []
source_revision_ids:
  - "srcrev_634deaee81463b4296fa241958397f62"
conflict_state: "none"
---

# Forced Offline Deployment & Feature Flags

## Summary

Runbook for deploying the Forced Offline project, including environment overview, PR process, feature flag implementation, and rollback procedures.

## Claims

- The production environment for Forced Offline is deployed from the main branch and accessible at https://forcedoffline.xyz/ with API at api.forcedoffline.xyz. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-1) `source_document_id=srcdoc_2153e9ac54dce828be221ce82dc73fff` `source_revision_id=srcrev_634deaee81463b4296fa241958397f62` `chunk_id=srcchunk_77ca5e528caeb4811a45dc6a351d2ed5` `native_locator=https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-1` `source_timestamp=2024-12-19T02:36:00Z`
- The staging environment is deployed from the staging branch and accessible at https://forcedoffline.vercel.app/ with API at stag.api.forcedoffline.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-1) `source_document_id=srcdoc_2153e9ac54dce828be221ce82dc73fff` `source_revision_id=srcrev_634deaee81463b4296fa241958397f62` `chunk_id=srcchunk_77ca5e528caeb4811a45dc6a351d2ed5` `native_locator=https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-1` `source_timestamp=2024-12-19T02:36:00Z`
- New feature branches should be created off the staging branch, and pull requests should target staging with WIP tags for in-progress work. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-1) `source_document_id=srcdoc_2153e9ac54dce828be221ce82dc73fff` `source_revision_id=srcrev_634deaee81463b4296fa241958397f62` `chunk_id=srcchunk_77ca5e528caeb4811a45dc6a351d2ed5` `native_locator=https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-1` `source_timestamp=2024-12-19T02:36:00Z`
- Feature flags can be implemented using environment variables prefixed with NEXT_PUBLIC_FEATURE_NAME_HERE, with optional timestamp-based activation using NEXT_PUBLIC_FEATURE_NAME_HERE_TIMESTAMP. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-2) `source_document_id=srcdoc_2153e9ac54dce828be221ce82dc73fff` `source_revision_id=srcrev_634deaee81463b4296fa241958397f62` `chunk_id=srcchunk_f4b475794c5f975e8a24fcca004b164b` `native_locator=https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-2` `source_timestamp=2024-12-19T02:36:00Z`
- Helper functions for feature flags should be placed in /helper/index.js, checking both the boolean flag and optional timestamp condition. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-2) `source_document_id=srcdoc_2153e9ac54dce828be221ce82dc73fff` `source_revision_id=srcrev_634deaee81463b4296fa241958397f62` `chunk_id=srcchunk_f4b475794c5f975e8a24fcca004b164b` `native_locator=https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-2` `source_timestamp=2024-12-19T02:36:00Z`
- Production rollbacks should follow Vercel's instant rollback documentation. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-2) `source_document_id=srcdoc_2153e9ac54dce828be221ce82dc73fff` `source_revision_id=srcrev_634deaee81463b4296fa241958397f62` `chunk_id=srcchunk_f4b475794c5f975e8a24fcca004b164b` `native_locator=https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523#chunk-2` `source_timestamp=2024-12-19T02:36:00Z`

## Sources

- `source_document_id`: `srcdoc_2153e9ac54dce828be221ce82dc73fff`
- `source_revision_id`: `srcrev_634deaee81463b4296fa241958397f62`
- `source_url`: [Notion source](https://www.notion.so/Testnet-Deployment-Testing-Under-Construction-6097d89959bf452fb1dfa8b66674b523)
