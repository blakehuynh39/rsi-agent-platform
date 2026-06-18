---
title: "Onboarding and Qualification"
type: "policy"
slug: "policies/numo-expert-onboarding"
freshness: "2026-06-18T17:14:00Z"
tags:
  - "onboarding"
  - "qualification"
  - "quiz"
  - "shadow-tasks"
owners: []
source_revision_ids:
  - "srcrev_192d55d8ba1fdee999f11c9430b91b88"
conflict_state: "none"
---

# Onboarding and Qualification

## Summary

Self-serve onboarding flow for Numo Expert contributors: complete profile, select languages, read quickstart guide, watch walkthrough, take language quiz with ≥85% pass, complete shadow tasks, and pass honeypot baseline before production. Required assets include quickstart guide, task walkthrough, video demo, FAQ, quality guide, language guide, payment guide, and common mistakes.

## Claims

- The self-serve onboarding flow: user enters Numo Expert, completes profile, selects languages/dialect, reads quickstart guide, watches optional walkthrough, takes language quiz, and if passed (≥85%), performs shadow tasks and honeypot baseline before production tasks. `claim:claim_2_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_4285c3868d836ba08d23d9683558fe6d` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-18T17:14:00Z`
- Required onboarding assets include Quickstart Guide, Task Walkthrough, Video Demo, FAQ, Quality Guide, Language Guide, Payment Guide, and Common Mistakes. `claim:claim_2_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_4285c3868d836ba08d23d9683558fe6d` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-18T17:14:00Z`
- Users must score ≥85% on the language quiz AND pass at least 2 honeypot-style known-answer tasks to qualify. `claim:claim_2_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_4285c3868d836ba08d23d9683558fe6d` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-18T17:14:00Z`
- Qualification quizzes include language comprehension, transcript correction, audio match, instruction comprehension, and honeypot baseline types. `claim:claim_2_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_4285c3868d836ba08d23d9683558fe6d` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-18T17:14:00Z`
- Quizzes can be implemented using TaskType.CUSTOM with metadata like passingScore 0.85, maxAttempts 3, isPaid false. `claim:claim_2_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2) `source_document_id=srcdoc_23bd1a05e81b1b9a88ad984844dcd70f` `source_revision_id=srcrev_192d55d8ba1fdee999f11c9430b91b88` `chunk_id=srcchunk_4285c3868d836ba08d23d9683558fe6d` `native_locator=https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b#chunk-2` `source_timestamp=2026-06-18T17:14:00Z`

## Sources

- `source_document_id`: `srcdoc_23bd1a05e81b1b9a88ad984844dcd70f`
- `source_revision_id`: `srcrev_192d55d8ba1fdee999f11c9430b91b88`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Implementation-PRD-37d051299a5480d38794d400f430876b)
