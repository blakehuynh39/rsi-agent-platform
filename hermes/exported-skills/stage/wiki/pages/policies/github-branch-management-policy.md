---
title: "Github Branch Management Policy"
type: "policy"
slug: "policies/github-branch-management-policy"
freshness: "2026-05-05T06:26:06Z"
tags:
  - "branch-management"
  - "development-process"
  - "github"
owners: []
source_revision_ids:
  - "srcrev_33bc0731129f0ab7dbe5ed0cb1f86549"
conflict_state: "none"
---

# Github Branch Management Policy

## Summary

Proposed branch management strategy for RSI company repositories, covering naming conventions, main branch protection, feature branch workflow, pull request reviews, CI, and release branches.

## Claims

- Limit the number of branches on the main repository. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c) `source_document_id=srcdoc_a70fe5ef63f037059739639fc5214aad` `source_revision_id=srcrev_33bc0731129f0ab7dbe5ed0cb1f86549` `chunk_id=srcchunk_77f1f7cb3f617f98523a2228af190c96` `native_locator=https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c` `source_timestamp=2026-05-05T06:26:06Z`
- Establish a consistent branch naming convention using prefixes such as feature/, bugfix/, or hotfix/ followed by a descriptive name. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c) `source_document_id=srcdoc_a70fe5ef63f037059739639fc5214aad` `source_revision_id=srcrev_33bc0731129f0ab7dbe5ed0cb1f86549` `chunk_id=srcchunk_77f1f7cb3f617f98523a2228af190c96` `native_locator=https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c` `source_timestamp=2026-05-05T06:26:06Z`
- Enable branch protection rules on the main branch to prevent forced pushes, deletion, and to require code reviews and passing tests before merging; only core team members can merge. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c) `source_document_id=srcdoc_a70fe5ef63f037059739639fc5214aad` `source_revision_id=srcrev_33bc0731129f0ab7dbe5ed0cb1f86549` `chunk_id=srcchunk_77f1f7cb3f617f98523a2228af190c96` `native_locator=https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c` `source_timestamp=2026-05-05T06:26:06Z`
- Use a feature branch workflow where feature branches are created from main and merged back via pull request; feature branches are for coordination, and if mostly one person works on a feature, a private fork can be used instead. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c) `source_document_id=srcdoc_a70fe5ef63f037059739639fc5214aad` `source_revision_id=srcrev_33bc0731129f0ab7dbe5ed0cb1f86549` `chunk_id=srcchunk_77f1f7cb3f617f98523a2228af190c96` `native_locator=https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c` `source_timestamp=2026-05-05T06:26:06Z`
- Require peer code reviews for all pull requests; use GitHub review features for feedback and adherence to coding standards; work-in-progress should be done on personal forks and submitted to the main repo for PR. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c) `source_document_id=srcdoc_a70fe5ef63f037059739639fc5214aad` `source_revision_id=srcrev_33bc0731129f0ab7dbe5ed0cb1f86549` `chunk_id=srcchunk_77f1f7cb3f617f98523a2228af190c96` `native_locator=https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c` `source_timestamp=2026-05-05T06:26:06Z`
- Integrate continuous integration (CI) tools like GitHub Actions or Jenkins to automate build, testing, coding style checks, commit message checks, linter, automated testing, gas estimate, and test coverage report for all PRs on the main branch. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c) `source_document_id=srcdoc_a70fe5ef63f037059739639fc5214aad` `source_revision_id=srcrev_33bc0731129f0ab7dbe5ed0cb1f86549` `chunk_id=srcchunk_77f1f7cb3f617f98523a2228af190c96` `native_locator=https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c` `source_timestamp=2026-05-05T06:26:06Z`
- Adopt a fixed release schedule, such as monthly releases, using release branches as snapshots of the main branch for bug fixes and maintenance while development continues on feature branches. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c) `source_document_id=srcdoc_a70fe5ef63f037059739639fc5214aad` `source_revision_id=srcrev_33bc0731129f0ab7dbe5ed0cb1f86549` `chunk_id=srcchunk_77f1f7cb3f617f98523a2228af190c96` `native_locator=https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c` `source_timestamp=2026-05-05T06:26:06Z`

## Open Questions

- Can PR review check boxes be automated using a GitHub bot?
- Should the main branch history be linear (rebase only) or allow merge commits?

## Sources

- `source_document_id`: `srcdoc_a70fe5ef63f037059739639fc5214aad`
- `source_revision_id`: `srcrev_33bc0731129f0ab7dbe5ed0cb1f86549`
- `source_url`: [Notion source](https://www.notion.so/Github-Branch-Management-1042b3ba64464722b6b00d4f765d0a7c)
