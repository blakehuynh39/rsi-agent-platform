---
title: "Numo Expert Implementation"
type: "project"
slug: "projects/numo-expert-implementation"
freshness: "2026-06-16T03:08:00Z"
tags:
  - "annotation"
  - "expert"
  - "numo"
  - "prd"
owners: []
source_revision_ids:
  - "srcrev_7f2b88527fdfca31996d4083f8b159e0"
conflict_state: "none"
---

# Numo Expert Implementation

## Summary

Numo Expert brings the Poseidon annotation workflow into app.numolabs.ai, with phases for parity, contractor beta, permissionless onboarding, consensus, honeypots, and payment.

## Claims

- Numo Expert brings the current internal Poseidon annotation workflow from annotation.psdn.ai into app.numolabs.ai. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_9694d0948c35f0335203d89d4f58f9c4` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-16T03:08:00Z`
- The MVP should focus on enabling qualified users to complete audio / transcript annotation tasks with quality controls, consensus, honeypots, onboarding, and payout tracking. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_9694d0948c35f0335203d89d4f58f9c4` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-16T03:08:00Z`
- Supported task types include TRANSCRIPT_CORRECTION, TRANSCRIPT_AUDIO_MATCH, CORRECTION_VALIDATION, and CUSTOM. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_9694d0948c35f0335203d89d4f58f9c4` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-16T03:08:00Z`
- A qualified user must have an existing account, verified email, accepted terms, completed profile, selected language, native dialect provided, not banned, passed language quiz, and passed shadow/honeypot baseline. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_9694d0948c35f0335203d89d4f58f9c4` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-16T03:08:00Z`
- Quizzes are implemented using TaskType.CUSTOM with a passing score of 85% and a maximum of 3 attempts, and are unpaid. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_0703c5cdc76573e75ec7f85147f48f93` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-16T03:08:00Z`
- Honeypot insertion rates vary by user level: new/shadow users 20-30%, qualified annotators 5-10%, trusted annotators 2-5%, reviewers 1-3%. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_162abdb601feda6dae0c9d30a0ecb04d` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-16T03:08:00Z`
- Annotator quality score is computed as 40% honeypot accuracy, 30% consensus agreement, 15% reviewer agreement, 10% time sanity, and 5% completion reliability. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_162abdb601feda6dae0c9d30a0ecb04d` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-16T03:08:00Z`
- Recommended minimum annotators per item: 3 for transcript correction, 5 for audio transcript match, 3 for correction validation, 5 for deepfake detection, 1 for honeypots/quizzes, 1 reviewer, and 1 expert judge. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_0703c5cdc76573e75ec7f85147f48f93` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-16T03:08:00Z`
- Implementation sequence: internal parity, contractor beta, quiz/honeypot foundation, consensus engine, reviewer/judge workflow, payment conversion. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_293ca47b2c952310f4040b2882745a9a` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-16T03:08:00Z`
- Key technical decisions: quizzes are CUSTOM tasks, honeypots use Item.metadata.isHoneypot, consensus config in Task.schema.validation, payment config in Task.schema.payment, reviewers managed via Team, user gating via profile+quiz+honeypot baseline. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_7f2b88527fdfca31996d4083f8b159e0` `chunk_id=srcchunk_293ca47b2c952310f4040b2882745a9a` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-16T03:08:00Z`

## Sources

- `source_document_id`: `srcdoc_23bd1a05e81b1b9a88ad984844dcd70f`
- `source_revision_id`: `srcrev_7f2b88527fdfca31996d4083f8b159e0`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b)
