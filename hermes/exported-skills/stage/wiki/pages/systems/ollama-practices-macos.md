---
title: "Ollama Practices on MacOS"
type: "system"
slug: "systems/ollama-practices-macos"
freshness: "2025-04-16T22:55:00Z"
tags:
  - "benchmark"
  - "llm-ops"
  - "macos"
  - "ollama"
owners: []
source_revision_ids:
  - "srcrev_ba9f0bb93d314162e37dde75028ec9b0"
conflict_state: "none"
---

# Ollama Practices on MacOS

## Summary

Practices for customizing and benchmarking Ollama models on MacOS, including Modelfile examples for llama3.2 and gemma3:4b, and benchmark results on Macbook Pro M4 Max.

## Claims

- Ollama models can be customized using a Modelfile with parameters like temperature and system message. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ollama-1ca051299a54808a8cbcf271ed801d8f#chunk-1) `source_document_id=srcdoc_1e2aafc5cdf5f285bef6c779ec7831e8` `source_revision_id=srcrev_ba9f0bb93d314162e37dde75028ec9b0` `chunk_id=srcchunk_a5cc69425623a3d3f0ec406a636bd2ca` `native_locator=https://www.notion.so/ollama-1ca051299a54808a8cbcf271ed801d8f#chunk-1` `source_timestamp=2025-04-16T22:55:00Z`
- A Modelfile for gemma3:4b was created with temperature 1, context window 4096, and a system message for legal patent review outputting strict JSON. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ollama-1ca051299a54808a8cbcf271ed801d8f#chunk-1) `source_document_id=srcdoc_1e2aafc5cdf5f285bef6c779ec7831e8` `source_revision_id=srcrev_ba9f0bb93d314162e37dde75028ec9b0` `chunk_id=srcchunk_a5cc69425623a3d3f0ec406a636bd2ca` `native_locator=https://www.notion.so/ollama-1ca051299a54808a8cbcf271ed801d8f#chunk-1` `source_timestamp=2025-04-16T22:55:00Z`
- Benchmark on Macbook Pro M4 Max for gemma3:4b shows cold start short prompt at 85.79 gen_tok/s, medium at 85.78 gen_tok/s, long at 84.13 gen_tok/s; warm start short at 88.66 gen_tok/s, medium at 85.40 gen_tok/s, long at 83.29 gen_tok/s. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ollama-1ca051299a54808a8cbcf271ed801d8f#chunk-1) `source_document_id=srcdoc_1e2aafc5cdf5f285bef6c779ec7831e8` `source_revision_id=srcrev_ba9f0bb93d314162e37dde75028ec9b0` `chunk_id=srcchunk_a5cc69425623a3d3f0ec406a636bd2ca` `native_locator=https://www.notion.so/ollama-1ca051299a54808a8cbcf271ed801d8f#chunk-1` `source_timestamp=2025-04-16T22:55:00Z`

## Related Pages

- `projects/e2e-model-performance-observation-evaluation`

## Sources

- `source_document_id`: `srcdoc_1e2aafc5cdf5f285bef6c779ec7831e8`
- `source_revision_id`: `srcrev_ba9f0bb93d314162e37dde75028ec9b0`
- `source_url`: [Notion source](https://www.notion.so/ollama-1ca051299a54808a8cbcf271ed801d8f)
