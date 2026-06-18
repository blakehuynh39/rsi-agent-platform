---
title: "Numo Expert Platform Implementation"
type: "project"
slug: "projects/numo-expert-platform"
freshness: "2026-06-18T16:59:00Z"
tags:
  - "annotation"
  - "hitl"
  - "implementation"
  - "mvp"
  - "quality-control"
owners: []
source_revision_ids:
  - "srcrev_35dce993a6ad7c6af743a623466372c6"
conflict_state: "none"
---

# Numo Expert Platform Implementation

## Summary

Bring the internal Poseidon annotation workflow into app.numolabs.ai, enabling qualified users to complete audio/transcript annotation tasks with quality controls, consensus, and payout tracking.

## Claims

- The project is broken into four phases: Phase 1 — Parity with internal annotation tool; Phase 1.5 — Onboard existing Poseidon contractors as beta/power users; Phase 2 — Add permissionless contributor onboarding, lightweight quizzes/prerequisites, and consensus, honeypots, and expert judge workflow; Phase 3 — Mobile HITL / deepfake detection / Toss Mini App. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_31f3f665cce5d74f04e716de98f3dd56` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-18T16:59:00Z`
- The current data model has Task → Item → Submission, with Task having many Items, and Assignment manually assigning a Task to a User, or TeamAssignment to a Team. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_31f3f665cce5d74f04e716de98f3dd56` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-18T16:59:00Z`
- The system supports four task types: TRANSCRIPT_CORRECTION, TRANSCRIPT_AUDIO_MATCH, CORRECTION_VALIDATION, and CUSTOM. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_31f3f665cce5d74f04e716de98f3dd56` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-18T16:59:00Z`
- MVP must‑have features include: task list/detail, submit/draft/skip, audio player, transcript editor, manual and team assignment, profile gating by language/dialect, basic quizzes via CUSTOM tasks, honeypot metadata in Item.metadata, consensus config in Task.schema, admin submission view, and duration tracking. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_b7a8d0baba6c4edf5497db1dc4e24df1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-18T16:59:00Z`
- Should‑have features for later phases include reviewer queue, expert judge team, quality score, shadow mode, quickstart docs, and FAQ/walkthrough. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_b7a8d0baba6c4edf5497db1dc4e24df1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-18T16:59:00Z`
- The implementation sequence is: Step 1 — Internal parity, Step 2 — Contractor beta (30 existing contractors), Step 3 — Quiz + honeypot foundation, Step 4 — Consensus engine, Step 5 — Reviewer/Judge workflow. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_b7a8d0baba6c4edf5497db1dc4e24df1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-18T16:59:00Z`

## Open Questions

- Consensus mechanism details — e.g., exact number of annotators, tie-breaking rules, and honeypot baseline establishment — were initially open but resolved to majority vote and expert judge escalation.

## Related Pages

- `numo-expert-quality-control`

## Sources

- `source_document_id`: `srcdoc_23bd1a05e81b1b9a88ad984844dcd70f`
- `source_revision_id`: `srcrev_35dce993a6ad7c6af743a623466372c6`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b)
