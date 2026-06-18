---
title: "Numo Expert Quality Control"
type: "concept"
slug: "concepts/numo-expert-quality-control"
freshness: "2026-06-18T16:59:00Z"
tags:
  - "consensus"
  - "expert-judge"
  - "honeypots"
  - "quality-control"
  - "reviewer"
owners: []
source_revision_ids:
  - "srcrev_35dce993a6ad7c6af743a623466372c6"
conflict_state: "none"
---

# Numo Expert Quality Control

## Summary

Quality assurance mechanisms for Numo Expert annotations, including consensus, honeypots, reviewer escalation, and quality scoring.

## Claims

- Consensus follows a flow: item receives multiple submissions → if consensus reached, accept result; otherwise route to reviewer pool → if reviewer confident, accept decision; else escalate to expert judge for final decision. `claim:claim_2_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_f217fe0f5ecc238ce224548c2686b5e1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T16:59:00Z`
- Honeypots are known-answer items inserted into the normal task flow; they include known transcript errors, known valid transcripts, known audio mismatches, known AI correction failures, known deepfakes/real audio, and instruction traps. `claim:claim_2_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_f217fe0f5ecc238ce224548c2686b5e1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T16:59:00Z`
- Honeypots can be created via seeded known-answer items (store expected answer in Item.metadata), synthetic red herrings (manipulated tasks), or a golden dataset of contractor-reviewed items. `claim:claim_2_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_f217fe0f5ecc238ce224548c2686b5e1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T16:59:00Z`
- Honeypot insertion frequency depends on user level: new/shadow users 20-30%, qualified annotators 5-10%, trusted annotators 2-5%, reviewers 1-3%, expert judges rare manual audits. `claim:claim_2_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_f217fe0f5ecc238ce224548c2686b5e1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T16:59:00Z`
- Quality score formula: 40% honeypot accuracy, 30% consensus agreement, 15% reviewer agreement, 10% time sanity, 5% completion reliability. `claim:claim_2_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_f217fe0f5ecc238ce224548c2686b5e1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T16:59:00Z`
- Tracked signals for quality include honeypot accuracy, consensus agreement, reviewer agreement, time sanity, and completion reliability. `claim:claim_2_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_35dce993a6ad7c6af743a623466372c6` `chunk_id=srcchunk_f217fe0f5ecc238ce224548c2686b5e1` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T16:59:00Z`

## Related Pages

- `numo-expert-platform`

## Sources

- `source_document_id`: `srcdoc_23bd1a05e81b1b9a88ad984844dcd70f`
- `source_revision_id`: `srcrev_35dce993a6ad7c6af743a623466372c6`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b)
