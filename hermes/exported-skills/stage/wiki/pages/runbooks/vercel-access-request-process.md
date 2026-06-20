---
title: "Vercel Access Request Process"
type: "runbook"
slug: "runbooks/vercel-access-request-process"
freshness: "2026-03-25T18:18:52Z"
tags:
  - "access"
  - "process"
  - "vercel"
owners:
  - "Vinod"
source_revision_ids:
  - "srcrev_17f05ccb2a8ce976fe86b78cca8669fa"
  - "srcrev_273b806a5eeb6e4b7ddd6217bda47d38"
  - "srcrev_44e76bbd1a08d9862665df6168928332"
  - "srcrev_4ab0a9a53a79444bb3abc94ddb7b5701"
  - "srcrev_6f4d68919280533659c2219cfe168b75"
  - "srcrev_c8b8ce7a7f80b6795906702aba04e0c5"
  - "srcrev_f1d8145a136d9d8a0aa4ee4ef04470b1"
conflict_state: "none"
---

# Vercel Access Request Process

## Summary

Process for requesting access to the company Vercel account. Involves a bot creating a tracking task and a human admin (Vinod) manually sending an invitation.

## Claims

- Access to Vercel requires a manual invitation sent by a human admin (Vinod) after the request is filed. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_44e76bbd1a08d9862665df6168928332` `chunk_id=srcchunk_7cf4cb2be2de34842239c8940cdc2a0d` `native_locator=slack:C0547N89JUB:1774374755.299529:1774462724.640789` `source_timestamp=2026-03-25T18:18:44Z`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_c8b8ce7a7f80b6795906702aba04e0c5` `chunk_id=srcchunk_d599953185dd47a616a5c08d6f2d98c1` `native_locator=slack:C0547N89JUB:1774374755.299529:1774462732.014949` `source_timestamp=2026-03-25T18:18:52Z`
- The access provisioning bot can create a task on the IT project board to track Vercel access requests (e.g., issue #294). `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_4ab0a9a53a79444bb3abc94ddb7b5701` `chunk_id=srcchunk_90779136c6b6274c4f14943553533604` `native_locator=slack:C0547N89JUB:1774374755.299529:1774374843.660229` `source_timestamp=2026-03-24T17:54:03Z`
- The bot initially encountered a permission error when creating the task, but later succeeded in creating it. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_f1d8145a136d9d8a0aa4ee4ef04470b1` `chunk_id=srcchunk_50bc89f5973884065f67ea5b6dd40f34` `native_locator=slack:C0547N89JUB:1774374755.299529:1774374827.884399` `source_timestamp=2026-03-24T17:53:47Z`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_4ab0a9a53a79444bb3abc94ddb7b5701` `chunk_id=srcchunk_90779136c6b6274c4f14943553533604` `native_locator=slack:C0547N89JUB:1774374755.299529:1774374843.660229` `source_timestamp=2026-03-24T17:54:03Z`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_273b806a5eeb6e4b7ddd6217bda47d38` `chunk_id=srcchunk_66190ad38fbd4afd410f63ad6fd76ef7` `native_locator=slack:C0547N89JUB:1774374755.299529:1774375033.286429` `source_timestamp=2026-03-24T17:57:13Z`
- The bot supports access request creation for GitHub, AWS (via Terraform PR), Google Workspace, 1Password, and Notion, but Vercel is not yet a supported application requiring manual intervention. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_6f4d68919280533659c2219cfe168b75` `chunk_id=srcchunk_d6e5a4801b601ad735eaf99ae984a8fe` `native_locator=slack:C0547N89JUB:1774374755.299529:1774375069.552419` `source_timestamp=2026-03-24T17:57:49Z`
- The Vercel access request was prioritized as P1 (high urgency). `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_17f05ccb2a8ce976fe86b78cca8669fa` `chunk_id=srcchunk_6deefcf26d18e4f2b06747ce09f08030` `native_locator=slack:C0547N89JUB:1774374755.299529:1774374810.229309` `source_timestamp=2026-03-24T17:53:30Z`
- The human admin who sent the invite is Vinod (Slack user U08D32C1EF3). `claim:claim_2_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f51862e05f2a179dae5973e09a6656b8` `source_revision_id=srcrev_44e76bbd1a08d9862665df6168928332` `chunk_id=srcchunk_7cf4cb2be2de34842239c8940cdc2a0d` `native_locator=slack:C0547N89JUB:1774374755.299529:1774462724.640789` `source_timestamp=2026-03-25T18:18:44Z`

## Open Questions

- When will Vercel be added to the bot's supported apps for automatic provisioning?

## Related Pages

- `access-provisioning-bot`
- `leo-chen`

## Sources

- `source_document_id`: `srcdoc_f51862e05f2a179dae5973e09a6656b8`
- `source_revision_id`: `srcrev_c8b8ce7a7f80b6795906702aba04e0c5`
