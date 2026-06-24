---
title: "Move docs repo to PIP-Labs-RE for branch protection flexibility"
type: "decision"
slug: "decisions/docs-repo-move-pip-labs-re"
freshness: "2026-02-02T18:52:28Z"
tags:
  - "branch-protection"
  - "docs"
  - "github"
owners:
  - "Jack"
  - "Jacob"
source_revision_ids:
  - "srcrev_83f7a49227247d4986c37ec2a8fb2fa1"
  - "srcrev_aeba5bed738fee5702a6aaf90ad1f7df"
  - "srcrev_b7875800c045b6eb17a5510aa32f0a03"
  - "srcrev_d30b1d9d46379fa2550093046376c04a"
  - "srcrev_d38d992b6a9fd92cae3443ad8c802c7b"
  - "srcrev_e8b30516e765ac8fb663f988de7f3623"
conflict_state: "none"
---

# Move docs repo to PIP-Labs-RE for branch protection flexibility

## Summary

The docs repo was moved to a new GitHub organization PIP-Labs-RE to allow the sole maintainer to push directly to main, bypassing enterprise branch protection rules that required approval. There is ongoing discussion about maintaining at least one approver for public-facing repos.

## Claims

- The docs repository was moved to a new GitHub organization called PIP-Labs-RE. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- Enterprise branch protection rules were blocking the sole maintainer from merging without approval. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- The new organization has lenient rules to allow direct pushes to main. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_e8b30516e765ac8fb663f988de7f3623` `chunk_id=srcchunk_f4e821e3bed97dc456d299b0643dd8de` `native_locator=slack:C0547N89JUB:1769627331.711949:1769743802.301819` `source_timestamp=2026-01-30T03:30:02Z`
- It is recommended to have at least one approver for public-facing repositories. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_aeba5bed738fee5702a6aaf90ad1f7df` `chunk_id=srcchunk_121106e207f74db6cec189a73f09c174` `native_locator=slack:C0547N89JUB:1769627331.711949:1769913587.130999` `source_timestamp=2026-02-01T02:39:47Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d38d992b6a9fd92cae3443ad8c802c7b` `chunk_id=srcchunk_95e8653ec59cbe51876a809f77b19f86` `native_locator=slack:C0547N89JUB:1769627331.711949:1770058348.033779` `source_timestamp=2026-02-02T18:52:28Z`
- Inviting a user to the org and adding as outside collaborator requires accepting two separate invitations. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_83f7a49227247d4986c37ec2a8fb2fa1` `chunk_id=srcchunk_09f17120a54fb5efe45440564c58a947` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629699.483029` `source_timestamp=2026-01-28T19:48:19Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_b7875800c045b6eb17a5510aa32f0a03` `chunk_id=srcchunk_234ddabe1fab16b7c0532814bbb78007` `native_locator=slack:C0547N89JUB:1769627331.711949:1769628747.002299` `source_timestamp=2026-01-28T19:32:27Z`

## Sources

- `source_document_id`: `srcdoc_e3f1cd8eff335b482cd951577e5b8bcb`
- `source_revision_id`: `srcrev_7f0e778c581973bc85622797ca01b711`
