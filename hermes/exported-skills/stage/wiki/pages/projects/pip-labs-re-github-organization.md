---
title: "PIP-Labs-RE GitHub Organization"
type: "project"
slug: "projects/pip-labs-re-github-organization"
freshness: "2026-02-02T18:52:28Z"
tags:
  - "access-control"
  - "docs"
  - "github"
  - "organization"
owners: []
source_revision_ids:
  - "srcrev_209c303befe3d89b0f601b1f6d270a1a"
  - "srcrev_41c1178cc55b6bd9b409f7d4da0ab4e9"
  - "srcrev_69eecd3a7fabaa934c44927b42e0942b"
  - "srcrev_730b43c6859ff32bf9cd31d699e30e4c"
  - "srcrev_7f0e778c581973bc85622797ca01b711"
  - "srcrev_83f7a49227247d4986c37ec2a8fb2fa1"
  - "srcrev_aeba5bed738fee5702a6aaf90ad1f7df"
  - "srcrev_b7875800c045b6eb17a5510aa32f0a03"
  - "srcrev_d30b1d9d46379fa2550093046376c04a"
  - "srcrev_d38d992b6a9fd92cae3443ad8c802c7b"
  - "srcrev_e8b30516e765ac8fb663f988de7f3623"
conflict_state: "none"
---

# PIP-Labs-RE GitHub Organization

## Summary

The PIP-Labs-RE organization was created to host the docs repository, allowing the sole maintainer to directly push to main while permitting reviewer access. It was established to circumvent an enterprise-level branch protection rule that required approval for merges, which was blocking the maintainer. The org and repo access invitations were sent to reviewers, and there was agreement that having at least one approver would be preferable for a public-facing repo.

## Claims

- A new GitHub organization, PIP-Labs-RE, was created to move the docs repository there. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- The organization was created because an enterprise-level branch protection rule blocked Jacob (sole maintainer) from merging changes without approval. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- The repo under the new org is public with lenient rules, allowing direct push to main. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_e8b30516e765ac8fb663f988de7f3623` `chunk_id=srcchunk_f4e821e3bed97dc456d299b0643dd8de` `native_locator=slack:C0547N89JUB:1769627331.711949:1769743802.301819` `source_timestamp=2026-01-30T03:30:02Z`
- There is a preference to have at least one approver for the public-facing repo, but direct push is currently allowed. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_aeba5bed738fee5702a6aaf90ad1f7df` `chunk_id=srcchunk_121106e207f74db6cec189a73f09c174` `native_locator=slack:C0547N89JUB:1769627331.711949:1769913587.130999` `source_timestamp=2026-02-01T02:39:47Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d38d992b6a9fd92cae3443ad8c802c7b` `chunk_id=srcchunk_95e8653ec59cbe51876a809f77b19f86` `native_locator=slack:C0547N89JUB:1769627331.711949:1770058348.033779` `source_timestamp=2026-02-02T18:52:28Z`
- Initial difficulty adding reviewers to the docs repo in PIP-Labs-RE was resolved by sending organization invites and adding collaborators. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_69eecd3a7fabaa934c44927b42e0942b` `chunk_id=srcchunk_34c592830554458ed6fdea9b89e8ebd8` `native_locator=slack:C0547N89JUB:1769627331.711949:1769627331.711949` `source_timestamp=2026-01-28T19:08:51Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_209c303befe3d89b0f601b1f6d270a1a` `chunk_id=srcchunk_fb7349a86245b84a4d2c94af91f7bdf0` `native_locator=slack:C0547N89JUB:1769627331.711949:1769627353.014979` `source_timestamp=2026-01-28T19:09:13Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_41c1178cc55b6bd9b409f7d4da0ab4e9` `chunk_id=srcchunk_31280963f956d42bfb452385a3db9162` `native_locator=slack:C0547N89JUB:1769627331.711949:1769628680.698659` `source_timestamp=2026-01-28T19:31:20Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_b7875800c045b6eb17a5510aa32f0a03` `chunk_id=srcchunk_234ddabe1fab16b7c0532814bbb78007` `native_locator=slack:C0547N89JUB:1769627331.711949:1769628747.002299` `source_timestamp=2026-01-28T19:32:27Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_730b43c6859ff32bf9cd31d699e30e4c` `chunk_id=srcchunk_2c94a171fe9995e409596aa7725b20c9` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629509.085009` `source_timestamp=2026-01-28T19:45:09Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_7f0e778c581973bc85622797ca01b711` `chunk_id=srcchunk_db9bead697b1fed797d4b45e16968810` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629592.195109` `source_timestamp=2026-01-28T19:46:32Z`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_83f7a49227247d4986c37ec2a8fb2fa1` `chunk_id=srcchunk_09f17120a54fb5efe45440564c58a947` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629699.483029` `source_timestamp=2026-01-28T19:48:19Z`

## Open Questions

- Is there a plan to enforce at least one approver for the docs repo in the future, given the preference for public-facing repos?

## Sources

- `source_document_id`: `srcdoc_e3f1cd8eff335b482cd951577e5b8bcb`
- `source_revision_id`: `srcrev_b7875800c045b6eb17a5510aa32f0a03`
