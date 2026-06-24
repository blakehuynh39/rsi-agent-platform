---
title: "Vault Infrastructure Setup"
type: "system"
slug: "systems/vault-infrastructure-setup"
freshness: "2026-01-28T23:41:44Z"
tags:
  - "kubernetes"
  - "story-helm"
  - "vault"
owners: []
source_revision_ids:
  - "srcrev_25accc771fc5889bdb207b448a25549f"
conflict_state: "none"
---

# Vault Infrastructure Setup

## Summary

Vault is deployed manually using manifests in story-helm repo. To update, edit the Vault custom resource image tag in default namespace.

## Claims

- The vault manifest and readme are located in the story-helm GitHub repository under bypass/bank-vault. `claim:claim_4_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- The vault setup was manual, and the readme covers all operations. `claim:claim_4_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- To update the vault image version, edit the Kubernetes custom resource 'vault' in the default namespace using kubectl get/edit. `claim:claim_4_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`

## Sources

- `source_document_id`: `srcdoc_b853172082170c2d6c8b40804fa03731`
- `source_revision_id`: `srcrev_04ca2c026aec72e98dc7cbb47e5c6637`
