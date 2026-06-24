---
title: "Security Audit: Sha256DataRegistry Trace Contracts"
type: "project"
slug: "projects/security-audit-sha256-data-registry-trace-contracts"
freshness: "2026-06-24T04:16:00Z"
tags:
  - "audit"
  - "blockchain"
  - "security"
  - "sha256-data-registry"
  - "trace-contracts"
owners:
  - "Raúl"
source_revision_ids:
  - "srcrev_5d8ac6e715ce4595280ef7ed6bd2a651"
conflict_state: "none"
---

# Security Audit: Sha256DataRegistry Trace Contracts

## Summary

AI security audit of storyprotocol/story-api trace contracts (commit ba415d4). Remediation PR #1281 fixed 5 of 8 findings (Low/Info), 3 acknowledged. No Critical/High vulnerabilities.

## Claims

- The remediation PR storyprotocol/story-api#1281 fixed 5 of 8 findings; 3 acknowledged as non-issues or accepted by design. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- Test suite: 58 passing, 1 skipped (gas bench); forge fmt clean. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- Verdict: No Critical/High findings; the on-chain design is solid with no external-attacker path. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- F-1 (Low) — Fixed: `redact_cmd()` masks `--private-key` in all logs/exceptions; residual ps exposure documented w/ keystore mitigation. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- F-2 (Low) — Acknowledged: Whole-batch revert is intentional atomicity; off-chain submitter pre-validates. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- F-3 (Low) — Fixed: `adminRebindIndex` (DEFAULT_ADMIN_ROLE) clears/rebinds a burned coordinate; `IndexBucket.rebind` added; mismatch-guarded. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- F-4 (Low) — Fixed: `VerifyBuckets` checks every bucket by default (`VERIFY_STRIDE=1`); sampling is explicit non-authoritative opt-in. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- F-5 (Info) — Acknowledged: Duplicate-hash index skip is the correct invariant-preserving behavior. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- F-6 (Info) — Fixed: `registered()`/`get()`/`eventAt()` now revert `BucketNotDeployed` for undeployed shards. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- F-7 (Info) — Acknowledged: Granting a role to `address(0)` is inert; input is admin-supplied at deploy time. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`
- F-8 (Info) — Fixed: `AccessControlDefaultAdminRulesUpgradeable` — 2-step admin transfer + 1-day delay; no direct grant/revoke. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1) `source_document_id=srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed` `source_revision_id=srcrev_5d8ac6e715ce4595280ef7ed6bd2a651` `chunk_id=srcchunk_26dbf3542365517c900c2b09507c54ed` `native_locator=https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f#chunk-1` `source_timestamp=2026-06-24T04:16:00Z`

## Sources

- `source_document_id`: `srcdoc_9d4d5b6a8e85c1f6b9e23703df8c38ed`
- `source_revision_id`: `srcrev_5d8ac6e715ce4595280ef7ed6bd2a651`
- `source_url`: [source](https://app.notion.com/p/trace-contracts-Sha256DataRegistry-AI-Security-Audit-Report-389051299a54814da2c6dbb8a5ca639f)
