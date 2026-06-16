---
title: "Numo Expert HITL Validation"
type: "concept"
slug: "concepts/numo-expert-validation"
freshness: "2026-06-16T03:13:00Z"
tags:
  - "consensus"
  - "honeypots"
  - "quality-assurance"
  - "sampling"
  - "validation"
owners:
  - "Nick (Ops)"
  - "Royce (QA)"
  - "Seb (Data Methodology)"
source_revision_ids:
  - "srcrev_4c67f92f7646472d7cf1af6b6218b03d"
conflict_state: "none"
---

# Numo Expert HITL Validation

## Summary

Describes the validation strategy including post‑submission sampling, consensus mechanisms, honeypot tasks, quality scoring, behavioral tracking, and contributor trust states.

## Claims

- Post‑submission validation uses a stratified sampling strategy based on WER buckets; higher error ranges (e.g. 80‑100% WER) are sampled aggressively (100% rate), while low‑WER items are sampled lightly. `claim:claim_3_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-6) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_8a2a15859302c7b09f398a82a9eccf50` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-6` `source_timestamp=2026-06-16T03:13:00Z`
- Consensus types differ by task: majority vote for binary validation, unique error threshold for transcript audio match, majority + reviewer escalation for correction validation, and majority + confidence weighting for deepfake detection. `claim:claim_3_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-6) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_8a2a15859302c7b09f398a82a9eccf50` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-6` `source_timestamp=2026-06-16T03:13:00Z`
- Honeypot tasks are known‑answer items injected to measure annotator quality; the honeypot rate is configurable per task via Task.schema (e.g., honeypotRate: 0.1). `claim:claim_3_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-5) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_793af77b4c0c8064237c76567da3fbd6` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-5` `source_timestamp=2026-06-16T03:13:00Z`
- Quality scoring combines honeypot accuracy (40%), consensus agreement (30%), reviewer agreement (15%), time sanity (10%), and completion reliability (5%). `claim:claim_3_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-7) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_0074b23695636d48479fee4845b577e4` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-7` `source_timestamp=2026-06-16T03:13:00Z`
- Behavioral tracking includes duration per item, mouse/tab inactivity, repeated instant submissions, copy/paste patterns, skip reason quality, and average task time vs peers. `claim:claim_3_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-7) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_0074b23695636d48479fee4845b577e4` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-7` `source_timestamp=2026-06-16T03:13:00Z`
- Contributors move through trust states: New → Training → Shadow → Active → Trusted → Reviewer → Expert Judge; quality issues can lead to Restricted or Banned. `claim:claim_3_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-7) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_0074b23695636d48479fee4845b577e4` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-7` `source_timestamp=2026-06-16T03:13:00Z`

## Open Questions

- Should maxSubmissions default be 3 or 5 for transcript workflows? (Owner: @Seb / @Nick)
- What is the correct unique error threshold? Current note says 3, needs confirmation (Owner: @Seb)
- What is the payout model during beta? (Owner: @Nick / @Allen)
- What WER thresholds should be default for post-submission sampling? (Owner: @Seb)

## Related Pages

- `numo-expert`
- `numo-expert-task-types`

## Sources

- `source_document_id`: `srcdoc_2e170372bdff094145bb549910241d88`
- `source_revision_id`: `srcrev_4c67f92f7646472d7cf1af6b6218b03d`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92)
