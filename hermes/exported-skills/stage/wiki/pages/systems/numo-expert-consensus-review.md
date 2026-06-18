---
title: "Consensus and Review Workflow"
type: "system"
slug: "systems/numo-expert-consensus-review"
freshness: "2026-06-18T17:14:00Z"
tags:
  - "consensus"
  - "expert-judge"
  - "quality-control"
  - "reviewer"
owners: []
source_revision_ids:
  - "srcrev_192d55d8ba1fdee999f11c9430b91b88"
conflict_state: "none"
---

# Consensus and Review Workflow

## Summary

Consensus mechanism uses majority voting (2 of 3 or 3 of 5) with escalation to Reviewer and then Expert Judge on mismatch. Reviewer pool handles conflicts, low-confidence consensus, high WER, many corrections, user complaints, and failed honeypots. Expert Judge provides final arbitration for unresolved cases and high-value data.

## Claims

- The current consensus mechanism picks 2 out of 3 or 3/5 (majority) and escalates mismatches to Expert Judge. `claim:claim_3_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_81edae39d82dc06e3dec92436105cf77` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-1` `source_timestamp=2026-06-18T17:14:00Z`
- Reviewers handle conflicting annotator results, low-confidence consensus, items with high WER, items with many corrections, user-reported confusion, and failed honeypot patterns. `claim:claim_3_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_f56a3a3583c6d771ae2dd105fe8f7e34` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T17:14:00Z`
- Expert Judges provide final arbitration when consensus fails, reviewer is uncertain, for high-value data, customer-sensitive datasets, or training data benchmark creation. `claim:claim_3_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_f56a3a3583c6d771ae2dd105fe8f7e34` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T17:14:00Z`
- The escalation flow: multiple submissions → consensus reached? Yes → accept; No → Reviewer Pool → reviewer confident? Yes → accept reviewer decision; No → Expert Judge → final decision. `claim:claim_3_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_f56a3a3583c6d771ae2dd105fe8f7e34` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T17:14:00Z`

## Sources

- `source_document_id`: `srcdoc_23bd1a05e81b1b9a88ad984844dcd70f`
- `source_revision_id`: `srcrev_192d55d8ba1fdee999f11c9430b91b88`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b)
