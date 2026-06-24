---
title: "Bank Vault Management"
type: "runbook"
slug: "runbooks/bank-vault-management"
freshness: "2026-01-28T23:41:44Z"
tags:
  - "devops"
  - "kubernetes"
  - "vault"
owners:
  - "U07TNT9N4JC"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_25accc771fc5889bdb207b448a25549f"
conflict_state: "none"
---

# Bank Vault Management

## Summary

How to manage the bank vault deployed in Kubernetes.

## Claims

- The vault manifest and readme are located at https://github.com/storyprotocol/story-helm/tree/bypass/bank-vault. The initial setup was manual. `claim:claim_3_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- To update the vault image version, use `kubectl get vault` to fetch its manifest and `kubectl edit` to modify the image tag. `claim:claim_3_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`

## Sources

- `source_document_id`: `srcdoc_b853172082170c2d6c8b40804fa03731`
- `source_revision_id`: `srcrev_e108af2fe4f8cd79a31c55c17dca3e86`
