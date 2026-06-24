---
title: "Vault (HashiCorp Vault) on Story Protocol"
type: "system"
slug: "systems/bank-vault"
freshness: "2026-01-28T23:41:44Z"
tags:
  - "devops"
  - "helm"
  - "kubernetes"
  - "vault"
owners:
  - "blake.huynh"
  - "tony"
source_revision_ids:
  - "srcrev_25accc771fc5889bdb207b448a25549f"
conflict_state: "none"
---

# Vault (HashiCorp Vault) on Story Protocol

## Summary

Information on the vault deployed manually via Helm and managed via Kubernetes.

## Claims

- The vault manifest and readme are located at https://github.com/storyprotocol/story-helm/tree/bypass/bank-vault `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- The initial vault setup was performed manually; the readme covers all operations. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- To update the vault image version, use `kubectl edit` on the vault custom resource in the default namespace. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- A previous vault legacy version issue was resolved by bumping the vault image version. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`

## Sources

- `source_document_id`: `srcdoc_b853172082170c2d6c8b40804fa03731`
- `source_revision_id`: `srcrev_25accc771fc5889bdb207b448a25549f`
