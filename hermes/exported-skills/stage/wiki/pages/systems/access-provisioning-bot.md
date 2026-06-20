---
title: "Access Provisioning Bot"
type: "system"
slug: "systems/access-provisioning-bot"
freshness: "2026-03-24T17:57:49Z"
tags:
  - "access-provisioning"
  - "automation"
  - "bot"
owners:
  - "Platform Team"
source_revision_ids:
  - "srcrev_0be65039fcf2f5ee94d90d1b761bb687"
  - "srcrev_273b806a5eeb6e4b7ddd6217bda47d38"
  - "srcrev_4ab0a9a53a79444bb3abc94ddb7b5701"
  - "srcrev_6f4d68919280533659c2219cfe168b75"
  - "srcrev_f1d8145a136d9d8a0aa4ee4ef04470b1"
conflict_state: "none"
---

# Access Provisioning Bot

## Summary

A bot that automates access requests by creating tasks and PRs for various applications. Currently supports GitHub, AWS, Google Workspace, 1Password, and Notion, with a goal to add Vercel and other apps.

## Claims

- The bot can create tasks or PRs to handle access requests. `claim:claim_3_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_273b806a5eeb6e4b7ddd6217bda47d38` `chunk_id=srcchunk_66190ad38fbd4afd410f63ad6fd76ef7` `native_locator=slack:C0547N89JUB:1774374755.299529:1774375033.286429` `source_timestamp=2026-03-24T17:57:13Z`
- Currently supported applications for automated provisioning: GitHub (repo/team access), AWS (via GSuite group + Terraform PR), Google Workspace (SSO groups, user management), 1Password (vault access), Notion (page access). `claim:claim_3_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_6f4d68919280533659c2219cfe168b75` `chunk_id=srcchunk_d6e5a4801b601ad735eaf99ae984a8fe` `native_locator=slack:C0547N89JUB:1774374755.299529:1774375069.552419` `source_timestamp=2026-03-24T17:57:49Z`
- Vercel is not yet a supported application; it requires manual steps for provisioning. `claim:claim_3_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_6f4d68919280533659c2219cfe168b75` `chunk_id=srcchunk_d6e5a4801b601ad735eaf99ae984a8fe` `native_locator=slack:C0547N89JUB:1774374755.299529:1774375069.552419` `source_timestamp=2026-03-24T17:57:49Z`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_0be65039fcf2f5ee94d90d1b761bb687` `chunk_id=srcchunk_0a25a49033a21adad4327a8aef4ebd59` `native_locator=slack:C0547N89JUB:1774374755.299529:1774375061.036659` `source_timestamp=2026-03-24T17:57:41Z`
- A feature request task was created to add Vercel support to the bot. `claim:claim_3_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_6f4d68919280533659c2219cfe168b75` `chunk_id=srcchunk_d6e5a4801b601ad735eaf99ae984a8fe` `native_locator=slack:C0547N89JUB:1774374755.299529:1774375069.552419` `source_timestamp=2026-03-24T17:57:49Z`
- The bot experienced a temporary permission error when creating a task (task #294) but recovered and successfully created it. `claim:claim_3_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_f1d8145a136d9d8a0aa4ee4ef04470b1` `chunk_id=srcchunk_50bc89f5973884065f67ea5b6dd40f34` `native_locator=slack:C0547N89JUB:1774374755.299529:1774374827.884399` `source_timestamp=2026-03-24T17:53:47Z`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_4ab0a9a53a79444bb3abc94ddb7b5701` `chunk_id=srcchunk_90779136c6b6274c4f14943553533604` `native_locator=slack:C0547N89JUB:1774374755.299529:1774374843.660229` `source_timestamp=2026-03-24T17:54:03Z`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_273b806a5eeb6e4b7ddd6217bda47d38` `chunk_id=srcchunk_66190ad38fbd4afd410f63ad6fd76ef7` `native_locator=slack:C0547N89JUB:1774374755.299529:1774375033.286429` `source_timestamp=2026-03-24T17:57:13Z`

## Open Questions

- What caused the temporary permission error for task creation?
- What is the timeline for adding Vercel provisioning support?

## Related Pages

- `leo-chen`
- `vercel-access-request-process`

## Sources

- `source_document_id`: `srcdoc_f51862e05f2a179dae5973e09a6656b8`
- `source_revision_id`: `srcrev_c8b8ce7a7f80b6795906702aba04e0c5`
