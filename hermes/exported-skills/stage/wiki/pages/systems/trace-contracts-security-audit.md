---
title: "Trace Contracts (Sha256DataRegistry) Security Audit"
type: "system"
slug: "systems/trace-contracts-security-audit"
freshness: "2026-06-24T03:46:00Z"
tags:
  - "audit"
  - "security"
  - "sha256-data-registry"
  - "trace-contracts"
owners:
  - "Raúl"
source_revision_ids:
  - "srcrev_6bb60147fdd3a595f707442e8a87f1e3"
conflict_state: "none"
---

# Trace Contracts (Sha256DataRegistry) Security Audit

## Summary

A security audit of the trace-contracts (Sha256DataRegistry) performed by AI (Claude) with reviewer Raúl. The on-chain design is solid with no Critical/High findings. Five out of eight findings were fixed, and three were acknowledged.

## Claims

- The on-chain design is solid – no Critical/High findings and no external-attacker path. The core invariant holds under all orderings, and CREATE3 bucket addresses are provably bound to the proxy. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_85a33e4a19742cee2cfe6808e3896604` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T03:46:00Z`
- Remediation PR storyprotocol/story-api#1281 (base staging) fixed 5 of 8 findings; 3 were acknowledged as non-issues / accepted by design. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_85a33e4a19742cee2cfe6808e3896604` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T03:46:00Z`
- F-1 (Low, Fixed): Deployer private keys were exposed in the process table and error logs via `--private-key` passed as a `cast` argv token. `claim:finding-1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_6a3d0f1df966487b569ef303e1516e5f` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2` `source_timestamp=2026-06-24T03:46:00Z`
- F-2 (Low, Acknowledged): A single bad item in `registerBatch` reverts the entire batch; the whole-batch revert is intentional to preserve atomicity. `claim:finding-2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_6a3d0f1df966487b569ef303e1516e5f` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-2` `source_timestamp=2026-06-24T03:46:00Z`
- F-3 (Low, Fixed): `adminRebindIndex` (DEFAULT_ADMIN_ROLE) was added to clear/rebind a burned coordinate, with mismatch-guarded `IndexBucket.rebind`. `claim:finding-3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_85a33e4a19742cee2cfe6808e3896604` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T03:46:00Z`
- F-4 (Low, Fixed): `VerifyBuckets` originally sampled only ~0.4% of buckets by default; now it checks every bucket (VERIFY_STRIDE=1). `claim:finding-4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_274accfaca3a6d7fdc99095d0354f3fa` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3` `source_timestamp=2026-06-24T03:46:00Z`
- F-5 (Info, Acknowledged): Duplicate-hash registration silently skips the index entry, which is the correct invariant-preserving behavior. `claim:finding-5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_274accfaca3a6d7fdc99095d0354f3fa` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3` `source_timestamp=2026-06-24T03:46:00Z`
- F-6 (Info, Fixed): Read paths (`registered()`, `get()`, `eventAt()`) previously returned default/empty for undeployed shards; they now revert `BucketNotDeployed`. `claim:finding-6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_274accfaca3a6d7fdc99095d0354f3fa` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-3` `source_timestamp=2026-06-24T03:46:00Z`
- F-7 (Info, Acknowledged): Granting a role to `address(0)` is inert because the input is admin-supplied at deploy time. `claim:finding-7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_85a33e4a19742cee2cfe6808e3896604` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T03:46:00Z`
- F-8 (Info, Fixed): `AccessControlDefaultAdminRulesUpgradeable` was adopted to enforce a 2-step admin transfer with a 2-day delay. `claim:finding-8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_6bb60147fdd3a595f707442e8a87f1e3` `chunk_id=srcchunk_85a33e4a19742cee2cfe6808e3896604` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T03:46:00Z`

## Sources

- `source_document_id`: `srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed`
- `source_revision_id`: `srcrev_6bb60147fdd3a595f707442e8a87f1e3`
- `source_url`: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f)
