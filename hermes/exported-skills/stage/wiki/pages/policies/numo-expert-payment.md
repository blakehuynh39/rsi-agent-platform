---
title: "Payment Model"
type: "policy"
slug: "policies/numo-expert-payment"
freshness: "2026-06-18T17:14:00Z"
tags:
  - "compensation"
  - "payment"
  - "pricing"
owners: []
source_revision_ids:
  - "srcrev_192d55d8ba1fdee999f11c9430b91b88"
conflict_state: "none"
---

# Payment Model

## Summary

Payment for annotation tasks is planned as per-task or per-item, scaled by task type and complexity (e.g., transcript length, audio duration). Payment requires consensus acceptance and honeypot pass. Payout will use existing Numo payment rails (Stripe) and stablecoins coming soon. The MVP will track duration for payment modeling, but full payout engine comes later.

## Claims

- Payment conversion: task hours to dollar value per N number of tasks; grouping subtasks within a task (e.g., 10 subtasks per task). `claim:claim_5_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_81edae39d82dc06e3dec92436105cf77` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-18T17:14:00Z`
- Payout will use existing Numo payment rails (Stripe) with stablecoin support coming soon. `claim:claim_5_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_81edae39d82dc06e3dec92436105cf77` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-18T17:14:00Z`
- Per-task payment config example: base amount $0.1 USD per task, requiring consensus acceptance and honeypot pass. `claim:claim_5_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_e3b6e2b37890538e113ba0d10007d379` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-18T17:14:00Z`
- Payment types will vary by task: transcription correction per item + transcript length, audio transcript match per item + audio duration, correction validation per item, deepfake detection per item, reviewer decision higher per item, expert judge decision highest per item. `claim:claim_5_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_e3b6e2b37890538e113ba0d10007d379` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-4` `source_timestamp=2026-06-18T17:14:00Z`

## Sources

- `source_document_id`: `srcdoc_23bd1a05e81b1b9a88ad984844dcd70f`
- `source_revision_id`: `srcrev_192d55d8ba1fdee999f11c9430b91b88`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b)
