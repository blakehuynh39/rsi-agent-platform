---
title: "Honeypots and Quality Control"
type: "system"
slug: "systems/numo-expert-honeypots"
freshness: "2026-06-18T17:14:00Z"
tags:
  - "baseline"
  - "honeypots"
  - "known-answers"
  - "quality-control"
owners: []
source_revision_ids:
  - "srcrev_192d55d8ba1fdee999f11c9430b91b88"
conflict_state: "none"
---

# Honeypots and Quality Control

## Summary

Honeypots are known-answer tasks inserted into the normal flow to establish baseline quality and catch careless annotators. Types include known transcript error, valid transcript, audio mismatch, AI correction failure, deepfake, human audio, and instruction trap. They are created via seeded known-answer items, synthetic red herrings, or golden datasets from existing reviewed annotations. Honeypot frequency varies by user trust level: 20-30% for new/shadow users, decreasing to rare manual audits for expert judges.

## Claims

- Honeypots are known-answer tasks inserted into normal task flow to establish baseline quality and catch careless annotators. `claim:claim_4_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_f56a3a3583c6d771ae2dd105fe8f7e34` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T17:14:00Z`
- Honeypot types include known transcript error, known valid transcript, known audio mismatch, known AI correction failure, known deepfake, known human audio, and instruction trap. `claim:claim_4_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_f56a3a3583c6d771ae2dd105fe8f7e34` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T17:14:00Z`
- Honeypots can be created via Option A (seeded known-answer items with expected output in metadata), Option B (synthetic red herrings like swapping a word), or Option C (golden dataset from existing high-confidence contractor-reviewed items). `claim:claim_4_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_f56a3a3583c6d771ae2dd105fe8f7e34` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T17:14:00Z`
- Honeypot insertion rate depends on user level: new/shadow user 20-30%, qualified annotator 5-10%, trusted annotator 2-5%, reviewer 1-3%, expert judge rare manual audit. `claim:claim_4_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_f56a3a3583c6d771ae2dd105fe8f7e34` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-3` `source_timestamp=2026-06-18T17:14:00Z`

## Sources

- `source_document_id`: `srcdoc_23bd1a05e81b1b9a88ad984844dcd70f`
- `source_revision_id`: `srcrev_192d55d8ba1fdee999f11c9430b91b88`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b)
