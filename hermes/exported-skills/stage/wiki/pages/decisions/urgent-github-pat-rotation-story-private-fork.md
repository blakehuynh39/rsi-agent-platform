---
title: "Urgent GitHub PAT Rotation for story-private-fork"
type: "decision"
slug: "decisions/urgent-github-pat-rotation-story-private-fork"
freshness: "2026-01-15T04:10:44Z"
tags:
  - "github"
  - "incident"
  - "pat"
  - "rotation"
  - "security"
  - "story-private-fork"
owners:
  - "subteam:S083BDZ4FTM"
  - "U07TNT9N4JC"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_04bd9d16842f5db94634a51b6c23761b"
  - "srcrev_262ad47e7adb91807e7600b582ba3eb4"
  - "srcrev_6f3c4e695593cfe075d53e6f9c082d43"
  - "srcrev_8d7ef6f56dde02148dc3b16f01dcb355"
  - "srcrev_de1c546378be17f3345876891a7e35e5"
conflict_state: "none"
---

# Urgent GitHub PAT Rotation for story-private-fork

## Summary

An urgent rotation of the GitHub Personal Access Token (PAT) used in the CI/CD pipeline for the story-private-fork repository was performed. The original PAT was possibly created at the org level by user U08332YRB7W. Due to urgency, a new PAT was issued by U07TNT9N4JC and replaced the existing one. The rotation was confirmed resolved.

## Claims

- A GitHub Personal Access Token (PAT) was used in the GitHub Actions workflow of the story-private-fork repository at `.github/workflows/build-release-artifacts.yml` line 46. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_6f3c4e695593cfe075d53e6f9c082d43` `chunk_id=srcchunk_13c9c2c4aec75770ce899943739ee930` `native_locator=slack:C0547N89JUB:1768440554.902359:1768440554.902359` `source_timestamp=2026-01-15T01:29:14Z`
- The PAT might have been created at the org level, possibly by user U08332YRB7W. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_de1c546378be17f3345876891a7e35e5` `chunk_id=srcchunk_a2e84528119477633ef9471816570c91` `native_locator=slack:C0547N89JUB:1768440554.902359:1768442686.011239` `source_timestamp=2026-01-15T02:04:46Z`
- Due to urgency, a new PAT was issued by U07TNT9N4JC and the existing PAT was replaced with the new one. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_8d7ef6f56dde02148dc3b16f01dcb355` `chunk_id=srcchunk_182d4903f509c2b1c37c7438f2a12198` `native_locator=slack:C0547N89JUB:1768440554.902359:1768443002.648879` `source_timestamp=2026-01-15T02:10:02Z`
- The PAT rotation issue was resolved. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_04bd9d16842f5db94634a51b6c23761b` `chunk_id=srcchunk_7ea6990728000d64cb5c1f347932e4dd` `native_locator=slack:C0547N89JUB:1768440554.902359:1768450208.009269` `source_timestamp=2026-01-15T04:10:08Z`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_262ad47e7adb91807e7600b582ba3eb4` `chunk_id=srcchunk_1c71b3e9ebd0e324a1ce5f6acb669ff1` `native_locator=slack:C0547N89JUB:1768440554.902359:1768450244.335849` `source_timestamp=2026-01-15T04:10:44Z`

## Open Questions

- Who are the Slack users U08332YRB7W, U07TNT9N4JC, and U04KTUN5WFQ? They are mentioned but not identified by name.
- Who originally created the PAT? It is unclear whether it was U08332YRB7W or someone else.

## Related Pages

- `story-private-fork`

## Sources

- `source_document_id`: `srcdoc_4d05b3a3fc2274b7c80534a1bb379745`
- `source_revision_id`: `srcrev_6f3c4e695593cfe075d53e6f9c082d43`
