---
title: "Vault Kubernetes Auth Namespace Whitelist Drift – May 2026"
type: "runbook"
slug: "runbooks/vault-auth-namespace-whitelist-drift-2026-05"
freshness: "2026-05-11T01:40:38Z"
tags:
  - "drift"
  - "incident"
  - "kubernetes"
  - "staking-api"
  - "vault"
owners:
  - "Blake (U0772SH7BRA)"
  - "Jinn (U0A7JJMU5T2)"
  - "Seong (U0AKJV8710S)"
source_revision_ids:
  - "srcrev_09ffca5f0db35abde026f614c42722ca"
  - "srcrev_a5dc9d9665e0641125822d4909846513"
  - "srcrev_b91142b414763f83d9656533a65e6606"
  - "srcrev_fb5ac7893623301bef2b23b031946baf"
conflict_state: "none"
---

# Vault Kubernetes Auth Namespace Whitelist Drift – May 2026

## Summary

Investigation of staking-api deployments crashing with 'namespace not authorized' Vault auth error. Root cause: new Kubernetes namespaces not allowed in Vault role default, compounded by Terraform reconcile job resetting the whitelist.

## Claims

- Staging staking-api deployments for mainnet and aeneid were in CrashLoopBackOff due to Vault Kubernetes auth error: 'namespace not authorized'. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cff3406825f6b8717a08bb31cf9dee3b` `source_revision_id=srcrev_09ffca5f0db35abde026f614c42722ca` `chunk_id=srcchunk_4715642da3f8711df43200064b63a5e4` `native_locator=slack:C0547N89JUB:1778303844.935079:1778304047.963549` `source_timestamp=2026-05-09T05:20:47Z`
- The Vault Kubernetes auth role 'default' has a bound_service_account_namespaces setting that only includes specific namespaces, and staking-mainnet and staking-aeneid were not in that list. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cff3406825f6b8717a08bb31cf9dee3b` `source_revision_id=srcrev_09ffca5f0db35abde026f614c42722ca` `chunk_id=srcchunk_4715642da3f8711df43200064b63a5e4` `native_locator=slack:C0547N89JUB:1778303844.935079:1778304047.963549` `source_timestamp=2026-05-09T05:20:47Z`
- Terraform-managed reconcile job in story-infra-aws resets the Vault role's allowed namespaces using variable vault_default_role_bound_service_account_namespaces, causing manual additions to be overwritten. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cff3406825f6b8717a08bb31cf9dee3b` `source_revision_id=srcrev_a5dc9d9665e0641125822d4909846513` `chunk_id=srcchunk_22fedb53c4627e674df4537674d46150` `native_locator=slack:C0547N89JUB:1778303844.935079:1778463425.751759` `source_timestamp=2026-05-11T01:37:05Z`
- Blake Huynh (U0772SH7BRA) was editing staging Vault configuration for RSI agent, which may have contributed to the whitelist being modified. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cff3406825f6b8717a08bb31cf9dee3b` `source_revision_id=srcrev_b91142b414763f83d9656533a65e6606` `chunk_id=srcchunk_801ca729b3238f1f7eb8cb05555c2f24` `native_locator=slack:C0547N89JUB:1778303844.935079:1778309512.535459` `source_timestamp=2026-05-09T06:51:52Z`
- Two potential fixes considered: (1) add staking-mainnet and staking-aeneid to the Terraform variable vault_default_role_bound_service_account_namespaces, or (2) remove the reconcile job to allow manual whitelist changes. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cff3406825f6b8717a08bb31cf9dee3b` `source_revision_id=srcrev_fb5ac7893623301bef2b23b031946baf` `chunk_id=srcchunk_04ff96c8e2a3912031a0aba2324e607b` `native_locator=slack:C0547N89JUB:1778303844.935079:1778463638.283729` `source_timestamp=2026-05-11T01:40:38Z`

## Open Questions

- Is the reconcile job causing other unintended drifts across environments?
- Should the Vault namespace whitelist be fully managed by Terraform IaC or allowed to drift manually? How to prevent future drift?
- Why were the new namespaces not added to the Terraform variable before deployment? What is the process for adding namespaces?

## Sources

- `source_document_id`: `srcdoc_cff3406825f6b8717a08bb31cf9dee3b`
- `source_revision_id`: `srcrev_3e2a4370eda10e30753e3879da42129f`
