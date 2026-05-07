---
title: "TrackStar"
type: "concept"
slug: "concepts/trackstar"
freshness: "2025-04-15T01:01:00Z"
tags:
  - "gradient-based-methods"
  - "llm"
  - "training-data-attribution"
owners: []
source_revision_ids:
  - "srcrev_fb208cada6d437aee268d4a0b15a77ff"
conflict_state: "none"
---

# TrackStar

## Summary

A gradient-based method for identifying influential training examples, outperforming prior methods but with limitations in fact-entailment retrieval.

## Claims

- TrackStar outperforms prior gradient-based methods in identifying examples that influence model predictions. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_daaf8eb685dd973b274d6a3fdf38012c` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2` `source_timestamp=2025-04-15T01:01:00Z`
- Classical retrieval approaches like BM25 still excel at retrieving fact-entailing examples compared to TrackStar. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_daaf8eb685dd973b274d6a3fdf38012c` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2` `source_timestamp=2025-04-15T01:01:00Z`
- TDA methods like TrackStar can be used for improved model transparency, debugging, and data curation by pinpointing influential pretraining examples. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_daaf8eb685dd973b274d6a3fdf38012c` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2` `source_timestamp=2025-04-15T01:01:00Z`
- There is a misalignment between influence and attribution: TrackStar retrieves examples that greatly impact predictions, but classic methods better capture text entailment of facts. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_daaf8eb685dd973b274d6a3fdf38012c` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2` `source_timestamp=2025-04-15T01:01:00Z`
- TrackStar may suffer when handling noisy or repetitive training examples. `claim:claim_2_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_daaf8eb685dd973b274d6a3fdf38012c` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2` `source_timestamp=2025-04-15T01:01:00Z`
- TrackStar requires careful tuning of corrections (optimizer state, Hessian approximation, unit normalization) and faces computational cost challenges with high-dimensional gradients. `claim:claim_2_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2) `source_document_id=srcdoc_84bf701847d4f50d60e4c6f67fceb209` `source_revision_id=srcrev_fb208cada6d437aee268d4a0b15a77ff` `chunk_id=srcchunk_daaf8eb685dd973b274d6a3fdf38012c` `native_locator=https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556#chunk-2` `source_timestamp=2025-04-15T01:01:00Z`

## Open Questions

- Can computational cost be reduced for high-dimensional gradients?
- How can the misalignment between influence and fact-entailment be resolved?
- What strategies mitigate noise and repetition in training data for TrackStar?

## Related Pages

- `debias-and-denoise-attribution-dda`

## Sources

- `source_document_id`: `srcdoc_84bf701847d4f50d60e4c6f67fceb209`
- `source_revision_id`: `srcrev_fb208cada6d437aee268d4a0b15a77ff`
- `source_url`: [Notion source](https://www.notion.so/Influence-Papers-Overview-1c8051299a54802192baf6d490c55556)
