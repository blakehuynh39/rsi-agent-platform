---
title: "Docs Repo Migration to PIP-Labs-RE"
type: "decision"
slug: "decisions/docs-repo-migration-to-pip-labs-re"
freshness: "2026-02-02T18:52:28Z"
tags:
  - "branch-protection"
  - "docs-repo"
  - "github"
  - "pip-labs-re"
owners:
  - "U08332YRB7W"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_209c303befe3d89b0f601b1f6d270a1a"
  - "srcrev_69eecd3a7fabaa934c44927b42e0942b"
  - "srcrev_730b43c6859ff32bf9cd31d699e30e4c"
  - "srcrev_83f7a49227247d4986c37ec2a8fb2fa1"
  - "srcrev_aeba5bed738fee5702a6aaf90ad1f7df"
  - "srcrev_d30b1d9d46379fa2550093046376c04a"
  - "srcrev_d38d992b6a9fd92cae3443ad8c802c7b"
  - "srcrev_e8b30516e765ac8fb663f988de7f3623"
conflict_state: "none"
---

# Docs Repo Migration to PIP-Labs-RE

## Summary

The docs repository was moved to a new GitHub organization PIP-Labs-RE to bypass enterprise branch protection rules, allowing the maintainer to push directly to main. Consensus remains on having at least one approver for public-facing changes.

## Claims

- The docs repository in PIP-Labs-RE had an enterprise-level branch protection rule that prevented Jacob from merging changes without approval. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- A new GitHub organization, PIP-Labs-RE, was created to host the docs repo with lenient branch protection to resolve the issue. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- Jacob, as the sole maintainer, desired the ability to directly push to main, which motivated the separation. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_e8b30516e765ac8fb663f988de7f3623` `chunk_id=srcchunk_f4e821e3bed97dc456d299b0643dd8de` `native_locator=slack:C0547N89JUB:1769627331.711949:1769743802.301819` `source_timestamp=2026-01-30T03:30:02Z`
- Despite the lenient rule, the team agreed that having at least one approver is preferable for public-facing repositories. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_aeba5bed738fee5702a6aaf90ad1f7df` `chunk_id=srcchunk_121106e207f74db6cec189a73f09c174` `native_locator=slack:C0547N89JUB:1769627331.711949:1769913587.130999` `source_timestamp=2026-02-01T02:39:47Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d38d992b6a9fd92cae3443ad8c802c7b` `chunk_id=srcchunk_95e8653ec59cbe51876a809f77b19f86` `native_locator=slack:C0547N89JUB:1769627331.711949:1770058348.033779` `source_timestamp=2026-02-02T18:52:28Z`
- Jack attempted to add himself and Meng as reviewers to a pull request in the new org but initially faced permission issues, which were resolved by sending org invitations and adding Meng as an outside collaborator. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_69eecd3a7fabaa934c44927b42e0942b` `chunk_id=srcchunk_34c592830554458ed6fdea9b89e8ebd8` `native_locator=slack:C0547N89JUB:1769627331.711949:1769627331.711949` `source_timestamp=2026-01-28T19:08:51Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_209c303befe3d89b0f601b1f6d270a1a` `chunk_id=srcchunk_fb7349a86245b84a4d2c94af91f7bdf0` `native_locator=slack:C0547N89JUB:1769627331.711949:1769627353.014979` `source_timestamp=2026-01-28T19:09:13Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_730b43c6859ff32bf9cd31d699e30e4c` `chunk_id=srcchunk_2c94a171fe9995e409596aa7725b20c9` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629509.085009` `source_timestamp=2026-01-28T19:45:09Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_83f7a49227247d4986c37ec2a8fb2fa1` `chunk_id=srcchunk_09f17120a54fb5efe45440564c58a947` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629699.483029` `source_timestamp=2026-01-28T19:48:19Z`

## Sources

- `source_document_id`: `srcdoc_e3f1cd8eff335b482cd951577e5b8bcb`
- `source_revision_id`: `srcrev_e8b30516e765ac8fb663f988de7f3623`
