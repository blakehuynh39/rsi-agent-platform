---
title: "GitHub PAT Rotation for story-private-fork (January 2026)"
type: "decision"
slug: "decisions/github-pat-story-private-fork-rotation-jan-2026"
freshness: "2026-01-15T04:10:08Z"
tags:
  - "ci-cd"
  - "github"
  - "incident"
  - "pat"
  - "security"
owners:
  - "S083BDZ4FTM"
source_revision_ids:
  - "srcrev_04bd9d16842f5db94634a51b6c23761b"
  - "srcrev_6f3c4e695593cfe075d53e6f9c082d43"
  - "srcrev_8d7ef6f56dde02148dc3b16f01dcb355"
  - "srcrev_de1c546378be17f3345876891a7e35e5"
conflict_state: "none"
---

# GitHub PAT Rotation for story-private-fork (January 2026)

## Summary

Investigation and resolution of an unknown GitHub personal access token (PAT) used in the story-private-fork GitHub Actions workflow. The token was urgently replaced by @U07TNT9N4JC after the issuer could not be determined.

## Claims

- A GitHub PAT was present in the story-private-fork repository's GitHub Actions workflow at `.github/workflows/build-release-artifacts.yml` line 46. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_6f3c4e695593cfe075d53e6f9c082d43` `chunk_id=srcchunk_13c9c2c4aec75770ce899943739ee930` `native_locator=slack:C0547N89JUB:1768440554.902359:1768440554.902359` `source_timestamp=2026-01-15T01:29:14Z`
- The PAT issuer was unknown; it used an alias and might have been created at the GitHub organization level by user U08332YRB7W. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_de1c546378be17f3345876891a7e35e5` `chunk_id=srcchunk_a2e84528119477633ef9471816570c91` `native_locator=slack:C0547N89JUB:1768440554.902359:1768442686.011239` `source_timestamp=2026-01-15T02:04:46Z`
- Due to urgency, user U07TNT9N4JC issued a new PAT and the existing one was replaced. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_8d7ef6f56dde02148dc3b16f01dcb355` `chunk_id=srcchunk_182d4903f509c2b1c37c7438f2a12198` `native_locator=slack:C0547N89JUB:1768440554.902359:1768443002.648879` `source_timestamp=2026-01-15T02:10:02Z`
- The issue was resolved. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d05b3a3fc2274b7c80534a1bb379745` `source_revision_id=srcrev_04bd9d16842f5db94634a51b6c23761b` `chunk_id=srcchunk_7ea6990728000d64cb5c1f347932e4dd` `native_locator=slack:C0547N89JUB:1768440554.902359:1768450208.009269` `source_timestamp=2026-01-15T04:10:08Z`

## Open Questions

- Who originally issued the PAT? The issuer remains uncertain; it may have been created at the org level.

## Sources

- `source_document_id`: `srcdoc_4d05b3a3fc2274b7c80534a1bb379745`
- `source_revision_id`: `srcrev_04bd9d16842f5db94634a51b6c23761b`
