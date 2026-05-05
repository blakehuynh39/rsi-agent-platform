---
title: "GitHub Permissions Policy"
type: "policy"
slug: "policies/github-permissions-policy"
freshness: "2026-05-05T06:26:04Z"
tags:
  - "access-control"
  - "github"
  - "permissions"
owners:
  - "user://6e49a49b-0756-434a-b0ff-5e6c7e7bfe20"
  - "user://aa6d3d9b-7b57-4cd2-be12-b7691c15100f"
  - "user://d5afbb5c-e4fa-4d48-a970-a1716d0c2a6b"
source_revision_ids:
  - "srcrev_5d046ea803f37136679a33410839b2ce"
conflict_state: "none"
---

# GitHub Permissions Policy

## Summary

Defines role-based access control, team assignments, and repository permissions for the organization.

## Claims

- The organization follows role-based access control (RBAC) and the principle of least privilege. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The main branch is protected. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- External contributors are limited. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- Read permission allows View and Clone, but not Push, Issue, PR, or Settings. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- Triage role (struck through, possibly deprecated) allowed View, Clone, Issue, and PR, but not Push or Settings. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- Write permission allows View, Clone, Push, Issue, and PR, but not Settings. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- Maintain permission allows all capabilities: View, Clone, Push, Issue, PR, and Settings. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The App-dev team (Application/SDK) includes Ze, Allen, Don, and Brent. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The Protocol-dev team includes Raul, Kingter, and Leeren. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The Devops team includes Leeren and Bruce. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The PDEE team includes Jason, Susan, and Weilei. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The Partners team includes a16z, 57blocks, tomo, storyverse, and advisor. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The protocol-contract repository is private, maintained by Raul, with read access for Partner and App teams, and write access for the Protocol team. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The typescript-sdk repository is public, maintained by Ze, with write access for the App team. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The explorer-app repository is public, maintained by Allen, with write access for the App team. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The kbw-demo repository is private, maintained by Allen, with read access for the Eng team and write access for the App team. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- The project-nova repository is public, maintained by Ze, with write access for the App team. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- Action item: review all repos next week (target date 2023-10-04). `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- Action item: implement the setup (target date 2023-10-03). `claim:claim_1_19` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`
- Organization owners are three users identified by internal user IDs. `claim:claim_1_20` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd) `source_document_id=srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3` `source_revision_id=srcrev_5d046ea803f37136679a33410839b2ce` `chunk_id=srcchunk_ea58404cc13e47f49d4ed742e9ae07b9` `native_locator=https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd` `source_timestamp=2026-05-05T06:26:04Z`

## Open Questions

- Are the action items from 2023 still pending, or have they been completed/superseded?

## Sources

- `source_document_id`: `srcdoc_f61bacbdcc2ebfe9db7c89c188f46be3`
- `source_revision_id`: `srcrev_5d046ea803f37136679a33410839b2ce`
- `source_url`: [Notion source](https://www.notion.so/Github-Permissions-8c6e83fb5a404cf2b67e1d83278df7fd)
