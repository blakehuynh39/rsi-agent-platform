---
title: "LLM Inference Decoding Strategies"
type: "concept"
slug: "concepts/llm-inference-decoding-strategies"
freshness: "2025-04-01T17:16:00Z"
tags:
  - "decoding"
  - "inference"
  - "llm"
owners: []
source_revision_ids:
  - "srcrev_a2ff6c4b838679c126e4d46920ecc8cf"
conflict_state: "none"
---

# LLM Inference Decoding Strategies

## Summary

Decoding is how a language model picks the next word/token step by step to generate text.

## Claims

- Decoding is the process by which a language model selects the next word/token step by step to generate text. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04) `source_document_id=srcdoc_eeb3e0650d9d683138b764ae10d4688a` `source_revision_id=srcrev_a2ff6c4b838679c126e4d46920ecc8cf` `chunk_id=srcchunk_70692654ae5267dd0e5e702df8b264de` `native_locator=https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04` `source_timestamp=2025-04-01T17:16:00Z`
- Greedy Search always picks the most likely next word, is fast and deterministic, but can loop and lacks creativity; best for quick facts and predictable answers. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04) `source_document_id=srcdoc_eeb3e0650d9d683138b764ae10d4688a` `source_revision_id=srcrev_a2ff6c4b838679c126e4d46920ecc8cf` `chunk_id=srcchunk_70692654ae5267dd0e5e702df8b264de` `native_locator=https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04` `source_timestamp=2025-04-01T17:16:00Z`
- Beam Search keeps multiple top paths and picks the best overall, offering more accuracy and coherence but is slower and can be repetitive/generic; used in translation and summarization. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04) `source_document_id=srcdoc_eeb3e0650d9d683138b764ae10d4688a` `source_revision_id=srcrev_a2ff6c4b838679c126e4d46920ecc8cf` `chunk_id=srcchunk_70692654ae5267dd0e5e702df8b264de` `native_locator=https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04` `source_timestamp=2025-04-01T17:16:00Z`
- Top-K Sampling randomly picks from the top K likely words, adding creativity and diversity, but choosing K is tricky (too small = dull, too large = messy); used in storytelling and brainstorming. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04) `source_document_id=srcdoc_eeb3e0650d9d683138b764ae10d4688a` `source_revision_id=srcrev_a2ff6c4b838679c126e4d46920ecc8cf` `chunk_id=srcchunk_70692654ae5267dd0e5e702df8b264de` `native_locator=https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04` `source_timestamp=2025-04-01T17:16:00Z`
- Top-P Sampling picks from the smallest set of words with total probability ≥ P, providing adaptive, high-quality, diverse output but requires tuning of P; used in chatbots and creative writing. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04) `source_document_id=srcdoc_eeb3e0650d9d683138b764ae10d4688a` `source_revision_id=srcrev_a2ff6c4b838679c126e4d46920ecc8cf` `chunk_id=srcchunk_70692654ae5267dd0e5e702df8b264de` `native_locator=https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04` `source_timestamp=2025-04-01T17:16:00Z`
- Temperature adjusts randomness by scaling word probabilities before sampling, offering easy randomness control but does not filter out low-probability words on its own; used for slogans (high T) and summaries (low T). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04) `source_document_id=srcdoc_eeb3e0650d9d683138b764ae10d4688a` `source_revision_id=srcrev_a2ff6c4b838679c126e4d46920ecc8cf` `chunk_id=srcchunk_70692654ae5267dd0e5e702df8b264de` `native_locator=https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04` `source_timestamp=2025-04-01T17:16:00Z`
- Beam Search example: translating "Le chat est sur le tapis" with k=3 explores paths like "The cat is on the mat", "The cat sat on the rug", "A cat is on the mat" simultaneously, choosing the one with highest overall score. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04) `source_document_id=srcdoc_eeb3e0650d9d683138b764ae10d4688a` `source_revision_id=srcrev_a2ff6c4b838679c126e4d46920ecc8cf` `chunk_id=srcchunk_70692654ae5267dd0e5e702df8b264de` `native_locator=https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04` `source_timestamp=2025-04-01T17:16:00Z`

## Sources

- `source_document_id`: `srcdoc_eeb3e0650d9d683138b764ae10d4688a`
- `source_revision_id`: `srcrev_a2ff6c4b838679c126e4d46920ecc8cf`
- `source_url`: [Notion source](https://www.notion.so/LLM-Inference-Decoding-Strategies-1c6051299a548003bb0de043008fff04)
