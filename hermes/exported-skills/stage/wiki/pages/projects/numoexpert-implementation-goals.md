---
title: "NumoExpertImplementationGoals"
type: "project"
slug: "projects/numoexpert-implementation-goals"
freshness: "2026-06-12T07:18:00Z"
tags:
  - "annotation"
  - "expert"
  - "implementation"
  - "numo"
  - "quality-control"
owners: []
source_revision_ids:
  - "srcrev_bc34e726162cf62fd468f998c23c3d33"
  - "srcrev_f2b31f445fbe72e514cff13704a225cd"
conflict_state: "none"
---

# NumoExpertImplementationGoals

## Summary

Implementation-focused PRD for Numo Expert: bringing internal Poseidon annotation workflows into app.numolabs.ai, with phased goals, data model, qualification, task assignment, consensus, honeypots, quality scoring, and payment.

## Claims

- Numo Expert brings the current internal Poseidon annotation workflow from annotation.psdn.ai into app.numolabs.ai, enabling audio/transcript annotation tasks with quality controls, consensus, honeypots, onboarding, and payout tracking. `claim:claim_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_47488c940022c168b367b4b50cf5b585` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-12T07:18:00Z`
- Core product goals are organized in phases: Phase 1 aims for parity with the internal annotation tool; Phase 1.5 focuses on onboarding existing Poseidon contractors as beta/power users; Phase 2 adds permissionless contributor onboarding, lightweight quizzes/prerequisites, and consensus/honeypots/expert judge workflow; Phase 3 targets mobile HITL, deepfake detection, and Toss Mini App. `claim:claim_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_47488c940022c168b367b4b50cf5b585` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-12T07:18:00Z`
- The current Prisma schema supports an MVP with Task, Item, Submission, Assignment, and TeamAssignment models. TaskType enum includes TRANSCRIPT_CORRECTION, TRANSCRIPT_AUDIO_MATCH, CORRECTION_VALIDATION, and CUSTOM (covering quizzes, honeypots, deepfake detection). `claim:claim_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_47488c940022c168b367b4b50cf5b585` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-12T07:18:00Z`
- To participate in MVP, a user must have an account, verified email, accepted terms, completed profile, selected language matching the task, specified native dialect, not be banned, passed a language quiz, and passed a honeypot baseline check. `claim:claim_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_47488c940022c168b367b4b50cf5b585` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-12T07:18:00Z`
- Lightweight qualification tasks (quizzes) include Language comprehension, Transcript correction, Audio match, Instruction comprehension, and Honeypot baseline; they are implemented using TaskType.CUSTOM with Item.metadata containing expected answers and a passing threshold. `claim:claim_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_6b48f38d548b076a3f56312102047687` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-12T07:18:00Z`
- Task assignment supports Direct assignment (admin to user), Team assignment, and a future Open Pool (qualified users can claim tasks). The recommended progression is manual assignment only (Phase 1), team assignment for contractor groups (Phase 1.5), and self-service claim with automatic routing (Phases 2-3). `claim:claim_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_6b48f38d548b076a3f56312102047687` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-12T07:18:00Z`
- Consensus requires multiple submissions per item until a configurable threshold is met. Recommended minimum annotators: Transcript correction (3), Audio transcript match (5), Correction validation (3), Deepfake detection (5), Honeypot/quiz (1), Reviewer escalation (1 reviewer), Expert judge (1 judge). `claim:claim_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_6b48f38d548b076a3f56312102047687` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-12T07:18:00Z`
- Honeypots can be created via Option A (seeded known-answer items with expected output in Item.metadata), Option B (synthetic red herrings like swapped words or wrong audio), or Option C (golden dataset from high-confidence contractor-reviewed items). `claim:claim_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_6e7139b5fb00c2277591003469b7170c` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-12T07:18:00Z`
- Honeypot insertion rates vary by user level: New/Shadow user 20-30%, Qualified annotator 5-10%, Trusted annotator 2-5%, Reviewer 1-3%, Expert Judge rare manual audit. Honeypots should be invisible to users. `claim:claim_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_6e7139b5fb00c2277591003469b7170c` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-12T07:18:00Z`
- Quality Score is calculated as 40% Honeypot Accuracy, 30% Consensus Agreement, 15% Reviewer Agreement, 10% Time Sanity, and 5% Completion Reliability, with enforcement actions like reducing score, restricting access, or banning for poor performance. `claim:claim_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_6e7139b5fb00c2277591003469b7170c` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-12T07:18:00Z`
- Payment will remain hourly for beta contractors, then convert to per-task rates once average durations are measured; only accepted tasks will be paid. `claim:claim_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_6e7139b5fb00c2277591003469b7170c` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-12T07:18:00Z`
- The implementation sequence covers six steps: (1) internal parity with existing workflows, (2) contractor beta with 30 existing contractors, (3) quiz and honeypot foundation, (4) consensus engine, (5) reviewer/judge workflow, and (6) payment conversion. `claim:claim_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_8ec273251c4520b84f661a3f260e40a2` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-12T07:18:00Z`
- Key technical decisions include: representing quizzes as TaskType.CUSTOM, marking honeypots via Item.metadata.isHoneypot, storing expected answers in Item.metadata.expectedOutput, configuring consensus in Task.schema.validation, managing reviewers using Teams, gating users with profile+quiz+honeypot baseline, defaulting to 3 annotators (5 for audio match), handling conflicts via Reviewer→Expert Judge escalation, and paying hourly for beta/public per-task. `claim:claim_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_bc34e726162cf62fd468f998c23c3d33` `chunk_id=srcchunk_8ec273251c4520b84f661a3f260e40a2` `native_locator=https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-12T07:18:00Z`
- Represents NumoExpert implementation goals referenced in the Implementation PRD. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Implementation-PRD-37d051299a5480129b0df970639ade82) `source_document_id=srcdoc_d5da7a34a4e8aaca7158c402f4ffe6f8` `source_revision_id=srcrev_f2b31f445fbe72e514cff13704a225cd` `chunk_id=srcchunk_3beb045b46cd68e6fcb6bc792ea19e8e` `native_locator=https://app.notion.com/p/Implementation-PRD-37d051299a5480129b0df970639ade82` `source_timestamp=2026-06-12T07:18:00Z`

## Sources

- `source_document_id`: `srcdoc_23bd1a05e81b1b9a88ad984844dcd70f`
- `source_revision_id`: `srcrev_bc34e726162cf62fd468f998c23c3d33`
- `source_url`: [source](https://app.notion.com/p/NumoExpertImplementationGoals-37d051299a5480d38794d400f430876b)
