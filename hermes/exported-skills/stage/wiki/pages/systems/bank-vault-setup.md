---
title: "Bank Vault Setup and Maintenance"
type: "system"
slug: "systems/bank-vault-setup"
freshness: "2026-01-28T23:41:44Z"
tags:
  - "kubernetes"
  - "manual-setup"
  - "story-helm"
  - "vault"
owners:
  - "@U07TNT9N4JC"
  - "@U08332YRB7W"
source_revision_ids:
  - "srcrev_25accc771fc5889bdb207b448a25549f"
conflict_state: "none"
---

# Bank Vault Setup and Maintenance

## Summary

The bank vault was set up manually; manifest and readme are stored in story-helm repository under bypass/bank-vault. Image version can be updated via kubectl edit on the vault CR.

## Claims

- The vault manifest and readme are located at https://github.com/storyprotocol/story-helm/tree/bypass/bank-vault, which documents the manual setup operations. `claim:claim_5_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- To update the vault image version, one can use `kubectl get vault` to fetch the manifest and `kubectl edit` to modify the image tag for the vault CR in the default namespace. `claim:claim_5_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`

## Sources

- `source_document_id`: `srcdoc_b853172082170c2d6c8b40804fa03731`
- `source_revision_id`: `srcrev_97c339aa56c30de445f0c4ee3c0d58ce`
