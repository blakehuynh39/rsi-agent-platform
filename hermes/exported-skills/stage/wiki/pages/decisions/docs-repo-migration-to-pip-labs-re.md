---
title: "Docs Repo Migration to PIP-Labs-RE GitHub Organization"
type: "decision"
slug: "decisions/docs-repo-migration-to-pip-labs-re"
freshness: "2026-02-02T18:52:28Z"
tags:
  - "branch-protection"
  - "docs"
  - "GitHub"
  - "PIP-Labs-RE"
owners:
  - "Jack"
  - "Jacob"
  - "Meng"
source_revision_ids:
  - "srcrev_209c303befe3d89b0f601b1f6d270a1a"
  - "srcrev_41c1178cc55b6bd9b409f7d4da0ab4e9"
  - "srcrev_69eecd3a7fabaa934c44927b42e0942b"
  - "srcrev_83f7a49227247d4986c37ec2a8fb2fa1"
  - "srcrev_aeba5bed738fee5702a6aaf90ad1f7df"
  - "srcrev_d30b1d9d46379fa2550093046376c04a"
  - "srcrev_d38d992b6a9fd92cae3443ad8c802c7b"
  - "srcrev_e8b30516e765ac8fb663f988de7f3623"
conflict_state: "none"
---

# Docs Repo Migration to PIP-Labs-RE GitHub Organization

## Summary

The public-facing documentation repository was moved to a new GitHub organization (PIP-Labs-RE) to allow the sole maintainer, Jacob, to directly push to the main branch without requiring additional approvals, bypassing enterprise-level branch protection rules. It was agreed that having at least one reviewer approval is preferable for public-facing changes.

## Claims

- A GitHub organization named PIP-Labs-RE hosts the docs repository. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_69eecd3a7fabaa934c44927b42e0942b` `chunk_id=srcchunk_34c592830554458ed6fdea9b89e8ebd8` `native_locator=slack:C0547N89JUB:1769627331.711949:1769627331.711949` `source_timestamp=2026-01-28T19:08:51Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- Enterprise-level branch protection rules previously blocked the sole maintainer from directly pushing to main. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- The maintainer Jacob wanted the ability to directly push to main. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_e8b30516e765ac8fb663f988de7f3623` `chunk_id=srcchunk_f4e821e3bed97dc456d299b0643dd8de` `native_locator=slack:C0547N89JUB:1769627331.711949:1769743802.301819` `source_timestamp=2026-01-30T03:30:02Z`
- The new PIP-Labs-RE org has a public repo with lenient branch protection rules. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- It is agreed that at least one reviewer approval should be required for public-facing documentation changes. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_aeba5bed738fee5702a6aaf90ad1f7df` `chunk_id=srcchunk_121106e207f74db6cec189a73f09c174` `native_locator=slack:C0547N89JUB:1769627331.711949:1769913587.130999` `source_timestamp=2026-02-01T02:39:47Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d38d992b6a9fd92cae3443ad8c802c7b` `chunk_id=srcchunk_95e8653ec59cbe51876a809f77b19f86` `native_locator=slack:C0547N89JUB:1769627331.711949:1770058348.033779` `source_timestamp=2026-02-02T18:52:28Z`
- Adding reviewers to the repo requires sending GitHub org invites and, for outside collaborators, adding them to the repo directly. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_41c1178cc55b6bd9b409f7d4da0ab4e9` `chunk_id=srcchunk_31280963f956d42bfb452385a3db9162` `native_locator=slack:C0547N89JUB:1769627331.711949:1769628680.698659` `source_timestamp=2026-01-28T19:31:20Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_83f7a49227247d4986c37ec2a8fb2fa1` `chunk_id=srcchunk_09f17120a54fb5efe45440564c58a947` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629699.483029` `source_timestamp=2026-01-28T19:48:19Z`
- Meng was added as an outside collaborator and invited to join the PIP-Labs-RE org. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_83f7a49227247d4986c37ec2a8fb2fa1` `chunk_id=srcchunk_09f17120a54fb5efe45440564c58a947` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629699.483029` `source_timestamp=2026-01-28T19:48:19Z`
- The pull request in question is #48. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_209c303befe3d89b0f601b1f6d270a1a` `chunk_id=srcchunk_fb7349a86245b84a4d2c94af91f7bdf0` `native_locator=slack:C0547N89JUB:1769627331.711949:1769627353.014979` `source_timestamp=2026-01-28T19:09:13Z`

## Sources

- `source_document_id`: `srcdoc_e3f1cd8eff335b482cd951577e5b8bcb`
- `source_revision_id`: `srcrev_693b6df6b3ee51055078a290190c8c73`
