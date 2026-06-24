---
title: "Decision: Moving docs Repository to PIP-Labs-RE GitHub Organization"
type: "decision"
slug: "decisions/decision-docs-repo-move-pip-labs-re"
freshness: "2026-02-02T18:52:28Z"
tags:
  - "branch-protection"
  - "docs"
  - "enterprise"
  - "github"
  - "pip-labs-re"
owners:
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_aeba5bed738fee5702a6aaf90ad1f7df"
  - "srcrev_d30b1d9d46379fa2550093046376c04a"
  - "srcrev_d38d992b6a9fd92cae3443ad8c802c7b"
  - "srcrev_e8b30516e765ac8fb663f988de7f3623"
conflict_state: "none"
---

# Decision: Moving docs Repository to PIP-Labs-RE GitHub Organization

## Summary

The docs repository was moved to a new GitHub organization (PIP-Labs-RE) to bypass enterprise-level branch protection and allow direct pushes to main by the sole maintainer, while retaining a public-facing requirement for at least one approver.

## Claims

- An enterprise-level branch protection rule blocked merging to the docs repository without approval. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- The docs repo was moved to a new GitHub organization, PIP-Labs-RE, with a lenient branch protection rule allowing direct pushes. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- Jacob, as the sole maintainer, wanted the ability to directly push to main. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_e8b30516e765ac8fb663f988de7f3623` `chunk_id=srcchunk_f4e821e3bed97dc456d299b0643dd8de` `native_locator=slack:C0547N89JUB:1769627331.711949:1769743802.301819` `source_timestamp=2026-01-30T03:30:02Z`
- GitHub Enterprise can make exceptions to branch protection at the repository visibility and branch level. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_e8b30516e765ac8fb663f988de7f3623` `chunk_id=srcchunk_f4e821e3bed97dc456d299b0643dd8de` `native_locator=slack:C0547N89JUB:1769627331.711949:1769743802.301819` `source_timestamp=2026-01-30T03:30:02Z`
- Despite the direct push capability, it is considered better to have at least one approver for the public-facing repository. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_aeba5bed738fee5702a6aaf90ad1f7df` `chunk_id=srcchunk_121106e207f74db6cec189a73f09c174` `native_locator=slack:C0547N89JUB:1769627331.711949:1769913587.130999` `source_timestamp=2026-02-01T02:39:47Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d38d992b6a9fd92cae3443ad8c802c7b` `chunk_id=srcchunk_95e8653ec59cbe51876a809f77b19f86` `native_locator=slack:C0547N89JUB:1769627331.711949:1770058348.033779` `source_timestamp=2026-02-02T18:52:28Z`

## Related Pages

- `runbook-adding-reviewers-pip-labs-re-docs`

## Sources

- `source_document_id`: `srcdoc_e3f1cd8eff335b482cd951577e5b8bcb`
- `source_revision_id`: `srcrev_09ec151eeddfe1ba2ecc4176865a91f1`
