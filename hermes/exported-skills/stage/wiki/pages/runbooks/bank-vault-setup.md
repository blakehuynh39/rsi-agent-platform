---
title: "Bank Vault Setup and Operations"
type: "runbook"
slug: "runbooks/bank-vault-setup"
freshness: "2026-01-28T23:41:44Z"
tags:
  - "deployment"
  - "kubernetes"
  - "vault"
owners: []
source_revision_ids:
  - "srcrev_25accc771fc5889bdb207b448a25549f"
conflict_state: "none"
---

# Bank Vault Setup and Operations

## Summary

The Bank Vault was set up manually. Its manifest and readme are located in the story-helm repository. To update the vault, edit the Kubernetes custom resource 'vault' in the default namespace (e.g., bump image tag).

## Claims

- Bank Vault manifest and readme are located at https://github.com/storyprotocol/story-helm/tree/bypass/bank-vault `claim:claim_3_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- The readme covers all operations done for vault setup. `claim:claim_3_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`
- Previous vault version issue was resolved by updating vault image version via 'kubectl edit vault'. `claim:claim_3_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_25accc771fc5889bdb207b448a25549f` `chunk_id=srcchunk_bcfdc3d27708c80720f09426fd832568` `native_locator=slack:C0547N89JUB:1769643694.436449:1769643694.436449` `source_timestamp=2026-01-28T23:41:44Z`

## Sources

- `source_document_id`: `srcdoc_b853172082170c2d6c8b40804fa03731`
- `source_revision_id`: `srcrev_c13b4c031f6ea986e666b0390abae8d2`
