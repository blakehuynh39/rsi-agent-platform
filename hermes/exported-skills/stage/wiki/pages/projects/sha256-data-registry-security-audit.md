---
title: "Sha256DataRegistry Security Audit"
type: "project"
slug: "projects/sha256-data-registry-security-audit"
freshness: "2026-06-24T02:29:00Z"
tags:
  - "security-audit"
  - "smart-contract"
  - "storyprotocol"
owners: []
source_revision_ids:
  - "srcrev_f89c2be9462319ce70d6f71ac554b162"
conflict_state: "none"
---

# Sha256DataRegistry Security Audit

## Summary

Security audit of the trace-contracts (Sha256DataRegistry) smart contract system by Story Protocol, revealing low and informational severity findings with no critical or high issues.

## Claims

- The on-chain design is solid with no critical or high severity findings, and no external-attacker path; the core invariant holds under all orderings. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_324b42216ff71c0254914fd31b021629` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T02:29:00Z`
- Deployer private keys are passed via command-line arguments and interpolated into error logs, risking disclosure (F-1). `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_324b42216ff71c0254914fd31b021629` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T02:29:00Z`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_29ce10a9179f44dc60dd31c789d475e5` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2` `source_timestamp=2026-06-24T02:29:00Z`
- A single bad item in registerBatch reverts the entire batch, discarding valid items (F-2). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_29ce10a9179f44dc60dd31c789d475e5` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2` `source_timestamp=2026-06-24T02:29:00Z`
- IndexConflict permanently burns a coordinate; no clear/overwrite/admin-repair path exists (F-3). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_29ce10a9179f44dc60dd31c789d475e5` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2` `source_timestamp=2026-06-24T02:29:00Z`
- Read paths return default values for undeployed shards, indistinguishable from unregistered hashes (F-6). `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_3f9ca8aa473afa49da4aa2671c451771` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3` `source_timestamp=2026-06-24T02:29:00Z`
- EXTRA_REGISTRARS are not zero-filtered while REGISTRY_ADMIN is, leading to asymmetry (F-7). `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_3f9ca8aa473afa49da4aa2671c451771` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3` `source_timestamp=2026-06-24T02:29:00Z`
- DEFAULT_ADMIN_ROLE has a single-step renounce/transfer footgun, risking permanent bricking of upgradeability and registrar management (F-8). `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_3f9ca8aa473afa49da4aa2671c451771` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3` `source_timestamp=2026-06-24T02:29:00Z`
- The audit covered Solidity 0.8.35 contracts with UUPS upgradeability, OpenZeppelin Contracts-Upgradeable v5.6.1, and Solady CREATE3. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_f89c2be9462319ce70d6f71ac554b162` `chunk_id=srcchunk_324b42216ff71c0254914fd31b021629` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T02:29:00Z`

## Sources

- `source_document_id`: `srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed`
- `source_revision_id`: `srcrev_f89c2be9462319ce70d6f71ac554b162`
- `source_url`: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f)
