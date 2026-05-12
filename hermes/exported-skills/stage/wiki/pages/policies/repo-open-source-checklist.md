---
title: "Open Source Repo Checklist"
type: "policy"
slug: "policies/repo-open-source-checklist"
freshness: "2024-08-25T15:21:00Z"
tags:
  - "checklist"
  - "open-source"
  - "repo-readiness"
owners:
  - "Andy"
  - "iliad"
  - "Leeren"
  - "Meng"
  - "Raul"
  - "Ze"
source_revision_ids:
  - "srcrev_9fbfe284665df18b21f9e977b219bd58"
conflict_state: "none"
---

# Open Source Repo Checklist

## Summary

A checklist for preparing the Omni and geth fork repositories for public open sourcing. It tracks completed items such as README, security policy, license, code of conduct, and branch protection, as well as outstanding work like contributing guide, CI fixes, unit tests, release process, and code refactoring.

## Claims

- The README.md must be present, look nice (referencing Omni Network's README as example), include an intro, a local setup guide, and a folder structure explanation. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- A security.md file must be created (assigned to Raul). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- The license must be GPL. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- An issue template must be added (assigned to Andy). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- Branch protection rules must be configured (assigned to Andy). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- A code of conduct must be present. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- PR workflows must be fixed. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- A contributing guide must be created, covering code conventions, PR conventions, comment conventions, and branch rules. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- Code refactoring is needed: remove unused logs, add missing comments, remove outdated comments, and remove unused folders (assigned to Ze). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- Unit tests must be added for the evmstaking module and other new code (assigned to Meng). `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- A release process must be defined for both the main project and the geth fork, including version rules, release notes, and release tags (assigned to Leeren). `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`
- Dependabot configuration must be fixed (assigned to Andy). `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9) `source_document_id=srcdoc_bc123e3fc541d178e0a32a5b0de42d8c` `source_revision_id=srcrev_9fbfe284665df18b21f9e977b219bd58` `chunk_id=srcchunk_a3e569c175cd2323fd0aca8c41499084` `native_locator=https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9` `source_timestamp=2024-08-25T15:21:00Z`

## Open Questions

- Has the geth fork release process been treated as a separate lifecycle or aligned with the main project?
- What is the target date for completing all open items?
- Who is responsible for the 'document links' and 'acknowledgement' sections?

## Related Pages

- `ci-pipeline-config`
- `repo-contributing-guide`
- `repo-geth-fork-release-process`

## Sources

- `source_document_id`: `srcdoc_bc123e3fc541d178e0a32a5b0de42d8c`
- `source_revision_id`: `srcrev_9fbfe284665df18b21f9e977b219bd58`
- `source_url`: [Notion source](https://www.notion.so/Open-source-repo-checklist-a641ad3b16964686920060f2a43b5cf9)
