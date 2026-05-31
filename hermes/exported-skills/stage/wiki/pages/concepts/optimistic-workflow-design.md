---
title: "Optimistic Workflow Design"
type: "concept"
slug: "concepts/optimistic-workflow-design"
freshness: "2026-02-04T16:44:00Z"
tags:
  - "challenger"
  - "optimistic"
  - "validation"
  - "workflow"
owners: []
source_revision_ids:
  - "srcrev_8231b73015a4add7f41dcee73baf2582"
conflict_state: "none"
---

# Optimistic Workflow Design

## Summary

Proposes a new optimistic workflow validation system that collapses miner and validator roles into a single worker, with an optional challenger to verify outputs, reducing redundant computation.

## Claims

- The current pipeline requires miners and validators to perform similar computationally expensive work, resulting in 4x the work compared to a centralized pipeline. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-1) `source_document_id=srcdoc_7b207c97e1b000117f0b3753c6db8763` `source_revision_id=srcrev_8231b73015a4add7f41dcee73baf2582` `chunk_id=srcchunk_45c11796a1919c2f09b00200c857b4da` `native_locator=https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-1` `source_timestamp=2026-02-04T16:44:00Z`
- The new optimistic workflow design collapses the separate miner and validator roles into a generalized 'worker' role. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-1) `source_document_id=srcdoc_7b207c97e1b000117f0b3753c6db8763` `source_revision_id=srcrev_8231b73015a4add7f41dcee73baf2582` `chunk_id=srcchunk_45c11796a1919c2f09b00200c857b4da` `native_locator=https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-1` `source_timestamp=2026-02-04T16:44:00Z`
- A challenger checks the worker's output and can challenge within a challenge period; if unchallenged, the workflow is marked completed. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-1) `source_document_id=srcdoc_7b207c97e1b000117f0b3753c6db8763` `source_revision_id=srcrev_8231b73015a4add7f41dcee73baf2582` `chunk_id=srcchunk_45c11796a1919c2f09b00200c857b4da` `native_locator=https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-1` `source_timestamp=2026-02-04T16:44:00Z`
- If a workflow is challenged, a new workflow is started to compare outputs; the losing party (original worker or challenger) is penalized. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-1) `source_document_id=srcdoc_7b207c97e1b000117f0b3753c6db8763` `source_revision_id=srcrev_8231b73015a4add7f41dcee73baf2582` `chunk_id=srcchunk_45c11796a1919c2f09b00200c857b4da` `native_locator=https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-1` `source_timestamp=2026-02-04T16:44:00Z`
- The challenger must listen to all workflow outputs onchain and selectively re-run the workflow to produce its own outputs for comparison. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-2) `source_document_id=srcdoc_7b207c97e1b000117f0b3753c6db8763` `source_revision_id=srcrev_8231b73015a4add7f41dcee73baf2582` `chunk_id=srcchunk_0ac8f8a602c2eae78627876be2284333` `native_locator=https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-2` `source_timestamp=2026-02-04T16:44:00Z`
- Challengers can run on powerful local GPUs (e.g., A100) with state-of-the-art open-source models instead of relying on external APIs, saving cost and speeding up processing. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-2) `source_document_id=srcdoc_7b207c97e1b000117f0b3753c6db8763` `source_revision_id=srcrev_8231b73015a4add7f41dcee73baf2582` `chunk_id=srcchunk_0ac8f8a602c2eae78627876be2284333` `native_locator=https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-2` `source_timestamp=2026-02-04T16:44:00Z`
- If all workers are trusted, no challenger is needed, making the system nearly as efficient as a centralized pipeline. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-2) `source_document_id=srcdoc_7b207c97e1b000117f0b3753c6db8763` `source_revision_id=srcrev_8231b73015a4add7f41dcee73baf2582` `chunk_id=srcchunk_0ac8f8a602c2eae78627876be2284333` `native_locator=https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6#chunk-2` `source_timestamp=2026-02-04T16:44:00Z`

## Sources

- `source_document_id`: `srcdoc_7b207c97e1b000117f0b3753c6db8763`
- `source_revision_id`: `srcrev_8231b73015a4add7f41dcee73baf2582`
- `source_url`: [Notion source](https://www.notion.so/Optimistic-Workflow-Design-29d051299a5480cb81a1d85e694df6c6)
