---
title: "E2E flow for model performance observation and evaluation"
type: "project"
slug: "projects/e2e-model-performance-observation-evaluation"
freshness: "2025-04-11T15:23:00Z"
tags:
  - "evaluation"
  - "inference"
  - "llm-ops"
  - "model-training"
owners: []
source_revision_ids:
  - "srcrev_551a4588e7258b1d76a7b316e3f4553e"
conflict_state: "none"
---

# E2E flow for model performance observation and evaluation

## Summary

Proposed approach to train a model using Group1's codebase on Prime Intellect, implement inference with Ollama/Llama.cpp for experiments and vllm/sglang for production, and validate using Confident AI and LangSmith. Known issue: gguf conversion causes auto-populated tokens in llama.cpp.

## Claims

- The proposed implementation approach includes training a model using Group1's codebase and implementing the inference methodology discussed last week. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac) `source_document_id=srcdoc_f69650e90ba0b6fad4625ef9dd884653` `source_revision_id=srcrev_551a4588e7258b1d76a7b316e3f4553e` `chunk_id=srcchunk_f67c042cd2acab3482bc1d785a716860` `native_locator=https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac` `source_timestamp=2025-04-11T15:23:00Z`
- Model training should leverage Group1's existing implementation on Prime Intellect. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac) `source_document_id=srcdoc_f69650e90ba0b6fad4625ef9dd884653` `source_revision_id=srcrev_551a4588e7258b1d76a7b316e3f4553e` `chunk_id=srcchunk_f67c042cd2acab3482bc1d785a716860` `native_locator=https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac` `source_timestamp=2025-04-11T15:23:00Z`
- Experimental inference should use Ollama/Llama.cpp from previous experiments. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac) `source_document_id=srcdoc_f69650e90ba0b6fad4625ef9dd884653` `source_revision_id=srcrev_551a4588e7258b1d76a7b316e3f4553e` `chunk_id=srcchunk_f67c042cd2acab3482bc1d785a716860` `native_locator=https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac` `source_timestamp=2025-04-11T15:23:00Z`
- Production inference pipeline should transition to vllm or sglang for enterprise-grade capabilities and optimized runtime performance. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac) `source_document_id=srcdoc_f69650e90ba0b6fad4625ef9dd884653` `source_revision_id=srcrev_551a4588e7258b1d76a7b316e3f4553e` `chunk_id=srcchunk_f67c042cd2acab3482bc1d785a716860` `native_locator=https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac` `source_timestamp=2025-04-11T15:23:00Z`
- Validation and testing should adopt Confident AI and LangSmith as suggested by Sarick. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac) `source_document_id=srcdoc_f69650e90ba0b6fad4625ef9dd884653` `source_revision_id=srcrev_551a4588e7258b1d76a7b316e3f4553e` `chunk_id=srcchunk_f67c042cd2acab3482bc1d785a716860` `native_locator=https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac` `source_timestamp=2025-04-11T15:23:00Z`
- When converting the trained model to gguf and running with llama.cpp, it auto-populates tokens instead of waiting for user input, leading to endless duplicated tokens. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac) `source_document_id=srcdoc_f69650e90ba0b6fad4625ef9dd884653` `source_revision_id=srcrev_551a4588e7258b1d76a7b316e3f4553e` `chunk_id=srcchunk_f67c042cd2acab3482bc1d785a716860` `native_locator=https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac` `source_timestamp=2025-04-11T15:23:00Z`
- Training models are available at https://huggingface.co/linkanjou/story-llm-influence-experiment/tree/main. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac) `source_document_id=srcdoc_f69650e90ba0b6fad4625ef9dd884653` `source_revision_id=srcrev_551a4588e7258b1d76a7b316e3f4553e` `chunk_id=srcchunk_f67c042cd2acab3482bc1d785a716860` `native_locator=https://www.notion.so/E2E-flow-for-model-performance-observation-and-evaluation-1d0051299a5480869973c3615c17ccac` `source_timestamp=2025-04-11T15:23:00Z`

## Related Pages

- `systems/ollama-practices-macos`

## Sources

- `source_document_id`: `srcdoc_1e2aafc5cdf5f285bef6c779ec7831e8`
- `source_revision_id`: `srcrev_ba9f0bb93d314162e37dde75028ec9b0`
- `source_url`: [Notion source](https://www.notion.so/ollama-1ca051299a54808a8cbcf271ed801d8f)
