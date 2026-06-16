---
title: "Numo Expert Task Types"
type: "concept"
slug: "concepts/numo-expert-task-types"
freshness: "2026-06-16T03:13:00Z"
tags:
  - "annotation"
  - "schema"
  - "task-types"
  - "workflow"
owners:
  - "Core Dev (26cd872b)"
  - "Seb (Data Methodology)"
source_revision_ids:
  - "srcrev_4c67f92f7646472d7cf1af6b6218b03d"
conflict_state: "none"
---

# Numo Expert Task Types

## Summary

Defines the task types, annotation workflow, and schema‑driven approach used by Numo Expert to support transcript correction, audio‑match, and future tasks.

## Claims

- The TaskType enum currently includes TRANSCRIPT_CORRECTION, TRANSCRIPT_AUDIO_MATCH, CORRECTION_VALIDATION, and CUSTOM. `claim:claim_2_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_94a83b071f26b4fa0d475747efc1467c` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4` `source_timestamp=2026-06-16T03:13:00Z`
- Pre-submission transcript sanity check will be represented using CUSTOM in Phase 1, with specific behavior defined in Task.schema. `claim:claim_2_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_94a83b071f26b4fa0d475747efc1467c` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4` `source_timestamp=2026-06-16T03:13:00Z`
- TRANSCRIPT_CORRECTION allows a user to correct a transcript without listening to the original audio; input contains transcript text and output includes a corrected version and list of issues. `claim:claim_2_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_94a83b071f26b4fa0d475747efc1467c` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4` `source_timestamp=2026-06-16T03:13:00Z`
- TRANSCRIPT_AUDIO_MATCH tasks validate whether audio matches the transcript, requiring an audio URL and transcript input, and produce a match result with word‑level errors. `claim:claim_2_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_94a83b071f26b4fa0d475747efc1467c` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4` `source_timestamp=2026-06-16T03:13:00Z`
- The generic annotation flow is: Task.schema defines workflow → Item.input provides payload → Annotator submits → Submission.output stored → Consensus/QA → final result or escalation. `claim:claim_2_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_94a83b071f26b4fa0d475747efc1467c` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-4` `source_timestamp=2026-06-16T03:13:00Z`
- The platform uses JSONB fields (Task.schema, Item.input, Item.metadata, Submission.output) to support dynamic task schemas without database migrations, enabling future task types like deepfake detection or RLHF. `claim:claim_2_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-5) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_793af77b4c0c8064237c76567da3fbd6` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-5` `source_timestamp=2026-06-16T03:13:00Z`

## Open Questions

- Should pre-submission sanity check remain CUSTOM or get its own enum later? (Owner: @Sasi)
- What is the correct unique error threshold for TRANSCRIPT_AUDIO_MATCH? (Owner: @Seb)

## Related Pages

- `numo-expert`
- `numo-expert-validation`

## Sources

- `source_document_id`: `srcdoc_2e170372bdff094145bb549910241d88`
- `source_revision_id`: `srcrev_4c67f92f7646472d7cf1af6b6218b03d`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92)
