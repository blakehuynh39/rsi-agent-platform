---
title: "PyTorch Practices on MacOS"
type: "runbook"
slug: "runbooks/pytorch-practices-macos"
freshness: "2025-04-04T23:17:00Z"
tags:
  - "accelerate"
  - "gemma3"
  - "macos"
  - "pytorch"
  - "transformers"
owners: []
source_revision_ids:
  - "srcrev_9b4d621baac1e2191b6eed69127b4094"
conflict_state: "none"
---

# PyTorch Practices on MacOS

## Summary

Guide for running the Gemma3 model on MacOS using PyTorch, transformers, and accelerate.

## Claims

- To run gemma3 on MacOS, install transformers and accelerate via pip, and ensure pip is upgraded. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/pytorch-1cb051299a5480afad23d8a65ce6f9cd) `source_document_id=srcdoc_580d2236bc62781654dc94b1ffb191b1` `source_revision_id=srcrev_9b4d621baac1e2191b6eed69127b4094` `chunk_id=srcchunk_d909b47ff2d321d430d3ff539b38e198` `native_locator=https://www.notion.so/pytorch-1cb051299a5480afad23d8a65ce6f9cd` `source_timestamp=2025-04-04T23:17:00Z`
- The sample code uses Hugging Face pipeline with model 'google/gemma-3-4b-it', device set to 'mps', and torch_dtype=torch.bfloat16. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/pytorch-1cb051299a5480afad23d8a65ce6f9cd) `source_document_id=srcdoc_580d2236bc62781654dc94b1ffb191b1` `source_revision_id=srcrev_9b4d621baac1e2191b6eed69127b4094` `chunk_id=srcchunk_d909b47ff2d321d430d3ff539b38e198` `native_locator=https://www.notion.so/pytorch-1cb051299a5480afad23d8a65ce6f9cd` `source_timestamp=2025-04-04T23:17:00Z`
- Alternatively, the model can be loaded using Gemma3ForConditionalGeneration.from_pretrained with device_map and AutoProcessor. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/pytorch-1cb051299a5480afad23d8a65ce6f9cd) `source_document_id=srcdoc_580d2236bc62781654dc94b1ffb191b1` `source_revision_id=srcrev_9b4d621baac1e2191b6eed69127b4094` `chunk_id=srcchunk_d909b47ff2d321d430d3ff539b38e198` `native_locator=https://www.notion.so/pytorch-1cb051299a5480afad23d8a65ce6f9cd` `source_timestamp=2025-04-04T23:17:00Z`

## Open Questions

- How to dynamically set device (mps/cpu/cuda) based on availability?

## Sources

- `source_document_id`: `srcdoc_580d2236bc62781654dc94b1ffb191b1`
- `source_revision_id`: `srcrev_9b4d621baac1e2191b6eed69127b4094`
- `source_url`: [Notion source](https://www.notion.so/pytorch-1cb051299a5480afad23d8a65ce6f9cd)
