---
title: "AI Model PIL Term Decision"
type: "decision"
slug: "decisions/ai-model-pil-term-decision"
freshness: "2024-04-19T22:48:00Z"
tags:
  - "AI Model"
  - "IPA"
  - "off-chain"
  - "PIL"
owners: []
source_revision_ids:
  - "srcrev_6e05fdf287ed09391b9bb065897566c2"
conflict_state: "none"
---

# AI Model PIL Term Decision

## Summary

Decision to introduce an off-chain AI Model boolean term in PIL and add an off-chain IPA metadata field `source` to track AI Model origin, rejecting the parent-child relationship with AI Model outputs.

## Claims

- The new PIL term `AI Model` is a boolean type expecting value true or false. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db) `source_document_id=srcdoc_9b42e6fab3c620f3a06210578a698d9a` `source_revision_id=srcrev_6e05fdf287ed09391b9bb065897566c2` `chunk_id=srcchunk_0fda6213bf3209227a6caf5906e8003d` `native_locator=https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db` `source_timestamp=2024-04-19T22:48:00Z`
- The `AI Model` term can be true only when the PIL terms attach to an IPA representing an AI Model. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db) `source_document_id=srcdoc_9b42e6fab3c620f3a06210578a698d9a` `source_revision_id=srcrev_6e05fdf287ed09391b9bb065897566c2` `chunk_id=srcchunk_0fda6213bf3209227a6caf5906e8003d` `native_locator=https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db` `source_timestamp=2024-04-19T22:48:00Z`
- The parameters `Sublicensable` and `Category-Specific-Derivatives-Cap` cannot be tagged when `AI Model` is true. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db) `source_document_id=srcdoc_9b42e6fab3c620f3a06210578a698d9a` `source_revision_id=srcrev_6e05fdf287ed09391b9bb065897566c2` `chunk_id=srcchunk_0fda6213bf3209227a6caf5906e8003d` `native_locator=https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db` `source_timestamp=2024-04-19T22:48:00Z`
- After discussion, Option 2 was chosen: the AI model is not the parent IP of its output. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db) `source_document_id=srcdoc_9b42e6fab3c620f3a06210578a698d9a` `source_revision_id=srcrev_6e05fdf287ed09391b9bb065897566c2` `chunk_id=srcchunk_0fda6213bf3209227a6caf5906e8003d` `native_locator=https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db` `source_timestamp=2024-04-19T22:48:00Z`
- An off-chain IPA metadata field named `source` will be added, storing the address of the AI Model IPA to indicate the IPA was generated from that AI Model; otherwise empty. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db) `source_document_id=srcdoc_9b42e6fab3c620f3a06210578a698d9a` `source_revision_id=srcrev_6e05fdf287ed09391b9bb065897566c2` `chunk_id=srcchunk_0fda6213bf3209227a6caf5906e8003d` `native_locator=https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db` `source_timestamp=2024-04-19T22:48:00Z`

## Sources

- `source_document_id`: `srcdoc_9b42e6fab3c620f3a06210578a698d9a`
- `source_revision_id`: `srcrev_6e05fdf287ed09391b9bb065897566c2`
- `source_url`: [Notion source](https://www.notion.so/Support-PIL-AI-Model-Term-bc774c99803542b19be97d7b007cb3db)
