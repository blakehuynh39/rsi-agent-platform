---
title: "Debias and Denoise Attribution (DDA)"
type: "concept"
slug: "concepts/debias-and-denoise-attribution-dda"
freshness: "2025-04-15T01:01:00Z"
tags:
  - "influence-functions"
  - "interpretability"
  - "llm"
  - "training-data-attribution"
owners: []
source_revision_ids:
  - "srcrev_fb208cada6d437aee268d4a0b15a77ff"
conflict_state: "none"
---

# Debias and Denoise Attribution (DDA)

## Summary

A novel Training Data Attribution method that addresses fitting errors in LLM training through debiasing and denoising strategies.

## Claims

- DDA was introduced in a paper published on arXiv in November 2024 (v2). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`
- DDA explicitly addresses fitting errors that occur in real-world LLM training, unlike standard influence function-based methods that assume perfect empirical risk minimization. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`
- DDA achieves more accurate attribution through debiasing (correcting for base model biases) and denoising (smoothing scores across training stages). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`
- Use cases for DDA include interpretability, data IP protection, hallucination tracing, ensuring trustworthiness, and diagnosing/improving models. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`
- DDA significantly outperforms baseline methods (TRAK, CEA, TracIN, BM25) on the hallucination tracing task, achieving high AUC and R@500 scores. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`
- DDA shows robust performance across different 7B LLM architectures (LLaMA2, Qwen2, Mistral). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`
- DDA maintains strong performance on models of varying sizes within the tested range (0.5B, 1.5B, 7B parameters). `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`
- Current DDA method is text-only and has not been demonstrated on multimodal LLMs. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`
- Scalability of DDA beyond 7B parameters (e.g., 100B-scale) has not been validated due to GPU resource constraints. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_49a7d4c7c065931451c5bc4ac617155e` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-1` `source_timestamp=2025-04-15T01:01:00Z`

## Open Questions

- Can DDA be extended to multimodal LLMs?
- Does DDA scale effectively to 100B+ parameter models?
- How complete are the scaling law experiments?

## Related Pages

- `trackstar`

## Sources

- `source_document_id`: `srcdoc_84bf701847d4f50d60e4c6f67fceb209`
- `source_revision_id`: `srcrev_fb208cada6d437aee268d4a0b15a77ff`
- `source_url`: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556)
