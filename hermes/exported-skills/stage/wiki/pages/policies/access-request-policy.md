---
title: "Access Request Policy"
type: "policy"
slug: "policies/access-request-policy"
freshness: "2026-03-30T17:04:54Z"
tags:
  - "access-control"
  - "claude-platform"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_167febc4f99caccde461be5e2c03155a"
  - "srcrev_97cb648bedcfbd14d9d27cc50e2ee6df"
  - "srcrev_b9628796690f67c5daa46db7cb44abe6"
conflict_state: "none"
---

# Access Request Policy

## Summary

Policy for requesting access to Claude Platform and Google Groups, using SecBot self-service with admin oversight. Discussed best practices for API management.

## Claims

- Regular users must submit their own access requests through SecBot (self-service) to ensure proper KYC verification and audit trails. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_69a1214f7112e8975cc8e79b9101dcd0` `source_revision_id=srcrev_97cb648bedcfbd14d9d27cc50e2ee6df` `chunk_id=srcchunk_7656ea4401ee6d368339b74221b2ea89` `native_locator=slack:C0547N89JUB:1774836787.943339:1774890294.188289` `source_timestamp=2026-03-30T17:04:54Z`
- Admins can request actions on behalf of users when needed. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_69a1214f7112e8975cc8e79b9101dcd0` `source_revision_id=srcrev_97cb648bedcfbd14d9d27cc50e2ee6df` `chunk_id=srcchunk_7656ea4401ee6d368339b74221b2ea89` `native_locator=slack:C0547N89JUB:1774836787.943339:1774890294.188289` `source_timestamp=2026-03-30T17:04:54Z`
- SecurityBot takes requests directly from the user; only admin can request on behalf of the user. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_69a1214f7112e8975cc8e79b9101dcd0` `source_revision_id=srcrev_b9628796690f67c5daa46db7cb44abe6` `chunk_id=srcchunk_d65ef85dd6c3540dace56c0420f3fdcc` `native_locator=slack:C0547N89JUB:1774836787.943339:1774890287.611259` `source_timestamp=2026-03-30T17:04:47Z`
- The IT admin can perform Google Group membership actions, but may have limited availability (e.g., offline on weekends). `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_69a1214f7112e8975cc8e79b9101dcd0` `source_revision_id=srcrev_167febc4f99caccde461be5e2c03155a` `chunk_id=srcchunk_bd202f1c9d6f4c3e1dac5576369f081c` `native_locator=slack:C0547N89JUB:1774836787.943339:1774837952.359009` `source_timestamp=2026-03-30T02:32:32Z`

## Open Questions

- What is the best practice to manage Claude APIs: invite all users to the platform or have admin spin up APIs for users?

## Sources

- `source_document_id`: `srcdoc_69a1214f7112e8975cc8e79b9101dcd0`
- `source_revision_id`: `srcrev_97cb648bedcfbd14d9d27cc50e2ee6df`
