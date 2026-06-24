---
title: "PIP-Labs-RE GitHub Org Setup"
type: "project"
slug: "projects/pip-labs-re-github-org-setup"
freshness: "2026-02-01T02:39:47Z"
tags:
  - "access-control"
  - "branch-protection"
  - "collaborators"
  - "docs"
  - "github"
  - "organization"
owners: []
source_revision_ids:
  - "srcrev_209c303befe3d89b0f601b1f6d270a1a"
  - "srcrev_41c1178cc55b6bd9b409f7d4da0ab4e9"
  - "srcrev_69eecd3a7fabaa934c44927b42e0942b"
  - "srcrev_83f7a49227247d4986c37ec2a8fb2fa1"
  - "srcrev_aeba5bed738fee5702a6aaf90ad1f7df"
  - "srcrev_b7875800c045b6eb17a5510aa32f0a03"
  - "srcrev_c78ddce4368bb1667c16a7be3184ea85"
  - "srcrev_d30b1d9d46379fa2550093046376c04a"
  - "srcrev_e8b30516e765ac8fb663f988de7f3623"
conflict_state: "none"
---

# PIP-Labs-RE GitHub Org Setup

## Summary

Setup of the PIP-Labs-RE GitHub organization to host the docs repository independently, driven by enterprise branch protection rules that blocked the sole maintainer from pushing directly to main. This page tracks the configuration of the org, repository access, and reviewer addition.

## Claims

- Jack was unable to add reviewers to the docs repo in PIP-Labs-RE. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_69eecd3a7fabaa934c44927b42e0942b` `chunk_id=srcchunk_34c592830554458ed6fdea9b89e8ebd8` `native_locator=slack:C0547N89JUB:1769627331.711949:1769627331.711949` `source_timestamp=2026-01-28T19:08:51Z`
- Jack requested to add himself and Meng as reviewers to PR #48 in the PIP-Labs-RE/docs repo. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_209c303befe3d89b0f601b1f6d270a1a` `chunk_id=srcchunk_fb7349a86245b84a4d2c94af91f7bdf0` `native_locator=slack:C0547N89JUB:1769627331.711949:1769627353.014979` `source_timestamp=2026-01-28T19:09:13Z`
- An invite was sent to Jack to join the PIP-Labs-RE org. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_41c1178cc55b6bd9b409f7d4da0ab4e9` `chunk_id=srcchunk_31280963f956d42bfb452385a3db9162` `native_locator=slack:C0547N89JUB:1769627331.711949:1769628680.698659` `source_timestamp=2026-01-28T19:31:20Z`
- Jack joined the org but still needed to be added to the repo separately. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_b7875800c045b6eb17a5510aa32f0a03` `chunk_id=srcchunk_234ddabe1fab16b7c0532814bbb78007` `native_locator=slack:C0547N89JUB:1769627331.711949:1769628747.002299` `source_timestamp=2026-01-28T19:32:27Z`
- Meng was added as an outside collaborator to the repo and sent an org invite, requiring both acceptances. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_83f7a49227247d4986c37ec2a8fb2fa1` `chunk_id=srcchunk_09f17120a54fb5efe45440564c58a947` `native_locator=slack:C0547N89JUB:1769627331.711949:1769629699.483029` `source_timestamp=2026-01-28T19:48:19Z`
- The PIP-Labs-RE org was created to move the docs repo there. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- Enterprise-level branch protection rules blocked Jacob (the sole maintainer) from pushing to main without approval. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- The repo remains public with lenient branch rules in the new org to allow direct pushes to main. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_d30b1d9d46379fa2550093046376c04a` `chunk_id=srcchunk_86a5002dc07947f9fdfbb4b4c0ffd777` `native_locator=slack:C0547N89JUB:1769627331.711949:1769736144.699799` `source_timestamp=2026-01-30T01:22:24Z`
- Jacob, as sole maintainer, wanted the ability to push directly to main, which motivated the separation. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_e8b30516e765ac8fb663f988de7f3623` `chunk_id=srcchunk_f4e821e3bed97dc456d299b0643dd8de` `native_locator=slack:C0547N89JUB:1769627331.711949:1769743802.301819` `source_timestamp=2026-01-30T03:30:02Z`
- Some participants questioned whether GitHub Enterprise could allow repo-level exceptions for direct pushes without approval. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_c78ddce4368bb1667c16a7be3184ea85` `chunk_id=srcchunk_a494ccf1d449141d022adc72af1d12fe` `native_locator=slack:C0547N89JUB:1769627331.711949:1769737611.364149` `source_timestamp=2026-01-30T01:46:51Z`
- Despite the lenient setup, participants agreed that having at least one approver is better for a public-facing repo. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e3f1cd8eff335b482cd951577e5b8bcb` `source_revision_id=srcrev_aeba5bed738fee5702a6aaf90ad1f7df` `chunk_id=srcchunk_121106e207f74db6cec189a73f09c174` `native_locator=slack:C0547N89JUB:1769627331.711949:1769913587.130999` `source_timestamp=2026-02-01T02:39:47Z`

## Open Questions

- Can GitHub Enterprise exceptions be configured at the repository level to allow direct pushes to main without approval?

## Sources

- `source_document_id`: `srcdoc_e3f1cd8eff335b482cd951577e5b8bcb`
- `source_revision_id`: `srcrev_d38d992b6a9fd92cae3443ad8c802c7b`
