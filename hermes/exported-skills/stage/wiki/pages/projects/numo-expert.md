---
title: "Numo Expert"
type: "project"
slug: "projects/numo-expert"
freshness: "2026-06-16T03:13:00Z"
tags:
  - "annotation"
  - "hitl"
  - "numo-expert"
  - "product-prd"
owners:
  - "Allen (Ops)"
  - "Core Dev (26cd872b)"
  - "Nick (Ops)"
  - "Royce (QA)"
  - "Sasi (Product Lead)"
  - "Seb (Data Methodology)"
source_revision_ids:
  - "srcrev_4c67f92f7646472d7cf1af6b6218b03d"
conflict_state: "none"
---

# Numo Expert

## Summary

Numo Expert is the public-facing evolution of Poseidon's internal annotation tool, aiming to scale human-in-the-loop annotation through a permissionless, reputation-based contributor network.

## Claims

- Numo Expert transforms the internal contractor-only annotation platform into a public, permissionless, reputation-based HITL annotation network. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-1) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_8b847fff859219960d3bd7efdedecc38` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-1` `source_timestamp=2026-06-16T03:13:00Z`
- Currently, annotation work is limited to internal Poseidon contractors and requires manual coordination by Nick. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-1) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_8b847fff859219960d3bd7efdedecc38` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-1` `source_timestamp=2026-06-16T03:13:00Z`
- The product has two major phases: Phase 1 – internal parity and contractor migration; Phase 2 – deepfake detection, mobile HITL, and Toss Mini App support. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-2) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_ee2952e53c784c76a71b3030c5eb616d` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-2` `source_timestamp=2026-06-16T03:13:00Z`
- The annotation system uses a schema‑driven architecture with JSONB fields (Task.schema, Item.input, Item.metadata, Submission.output) to support multiple workflow types without database migrations. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-5) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_793af77b4c0c8064237c76567da3fbd6` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-5` `source_timestamp=2026-06-16T03:13:00Z`
- Existing contractors (~30) coordinated by Nick will become the initial beta / power users to establish quality baselines and seed the platform. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-3) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_dea5fe1f8b39178dd1c4ee7a0f30f50b` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-3` `source_timestamp=2026-06-16T03:13:00Z`
- The launch plan is structured in 5 stages: Internal Parity Build, Contractor Beta, Controlled Public Beta, Permissionless Launch, and Mobile/Deepfake HITL. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-11) `source_document_id=srcdoc_2e170372bdff094145bb549910241d88` `source_revision_id=srcrev_4c67f92f7646472d7cf1af6b6218b03d` `chunk_id=srcchunk_80b8b8b4f47f8f6866013c51b6a368d3` `native_locator=https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92#chunk-11` `source_timestamp=2026-06-16T03:13:00Z`

## Open Questions

- How much of reviewer workflow is required for MVP? (Owner: @Allen / @Sasi)
- How should existing contractor quality baselines be imported? (Owner: @Nick)
- Should maxSubmissions default be 3 or 5 for transcript workflows? (Owner: @Seb / @Nick)
- Should pre-submission sanity check remain CUSTOM or get its own enum later? (Owner: @Sasi)
- What is the correct unique error threshold? Current note says 3, needs confirmation (Owner: @Seb)
- What is the exact random sampling rate for pre-submission validation? (Owner: @Seb)
- What is the payout model during beta? (Owner: @Nick / @Allen)
- What minimum docs are needed before permissionless launch? (Owner: @Nick / @Allen)
- What WER thresholds should be default for post-submission sampling? (Owner: @Seb)
- When do we promote someone to Reviewer? (Owner: @Allen / @Nick)

## Related Pages

- `numo-expert-launch-plan`
- `numo-expert-task-types`
- `numo-expert-validation`

## Sources

- `source_document_id`: `srcdoc_2e170372bdff094145bb549910241d88`
- `source_revision_id`: `srcrev_4c67f92f7646472d7cf1af6b6218b03d`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Product-PRD-37d051299a5480d7b29be0914fdaef92)
