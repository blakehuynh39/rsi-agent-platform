---
title: "llama.cpp on macOS: Setup and Benchmarks"
type: "runbook"
slug: "runbooks/llama-cpp-macos-setup-and-benchmarks"
freshness: "2025-04-16T22:47:00Z"
tags:
  - "benchmark"
  - "gemma3"
  - "llama.cpp"
  - "macos"
  - "metal"
owners: []
source_revision_ids:
  - "srcrev_f5058e25572e480491bb41b780678740"
conflict_state: "none"
---

# llama.cpp on macOS: Setup and Benchmarks

## Summary

A runbook for installing llama.cpp on macOS and running benchmarks with the Gemma 3 4B Q4_0 model on MacBook Pro M4 Max and Mac Studio M3 Ultra.

## Claims

- To install llama.cpp on macOS, clone the repository, configure the build with cmake (defaults to Metal + Accelerate), and compile in Release mode. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-1) `source_document_id=srcdoc_a8dd47ca9daa2af00a0c2b2e62c628ab` `source_revision_id=srcrev_f5058e25572e480491bb41b780678740` `chunk_id=srcchunk_c06450dc01dd0e5233341250db4a52c2` `native_locator=https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-1` `source_timestamp=2025-04-16T22:47:00Z`
- On a MacBook Pro M4 Max, the Gemma 3 4B Q4_0 model achieves text generation speeds of 82.86 t/s (tg128), 82.22 t/s (tg256), and 80.91 t/s (tg512) using the Metal backend with 10 threads. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-1) `source_document_id=srcdoc_a8dd47ca9daa2af00a0c2b2e62c628ab` `source_revision_id=srcrev_f5058e25572e480491bb41b780678740` `chunk_id=srcchunk_c06450dc01dd0e5233341250db4a52c2` `native_locator=https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-1` `source_timestamp=2025-04-16T22:47:00Z`
- On a MacBook Pro M4 Max, prompt processing speeds for the Gemma 3 4B Q4_0 model with a 1024-token prompt vary by batch size: 1324.04 t/s (batch 128), 1395.28 t/s (batch 256), and 1327.91 t/s (batch 512). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-1) `source_document_id=srcdoc_a8dd47ca9daa2af00a0c2b2e62c628ab` `source_revision_id=srcrev_f5058e25572e480491bb41b780678740` `chunk_id=srcchunk_c06450dc01dd0e5233341250db4a52c2` `native_locator=https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-1` `source_timestamp=2025-04-16T22:47:00Z`
- On a Mac Studio M3 Ultra, the Gemma 3 4B Q4_0 model achieves text generation speeds of 98.99 t/s (tg128), 98.75 t/s (tg256), and 98.23 t/s (tg512) using the Metal backend with 24 threads. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-2) `source_document_id=srcdoc_a8dd47ca9daa2af00a0c2b2e62c628ab` `source_revision_id=srcrev_f5058e25572e480491bb41b780678740` `chunk_id=srcchunk_35c28d665341600027588e0bb3f03d82` `native_locator=https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-2` `source_timestamp=2025-04-16T22:47:00Z`
- On a Mac Studio M3 Ultra, prompt processing performance for the Gemma 3 4B Q4_0 model with 512 tokens and 24 threads varies significantly with the number of GPU layers offloaded, ranging from 325.82 t/s (10 layers) to 1043.76 t/s (30 layers). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-3) `source_document_id=srcdoc_a8dd47ca9daa2af00a0c2b2e62c628ab` `source_revision_id=srcrev_f5058e25572e480491bb41b780678740` `chunk_id=srcchunk_2a2bfa4f61688c021a4f91549904839f` `native_locator=https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad#chunk-3` `source_timestamp=2025-04-16T22:47:00Z`

## Sources

- `source_document_id`: `srcdoc_a8dd47ca9daa2af00a0c2b2e62c628ab`
- `source_revision_id`: `srcrev_f5058e25572e480491bb41b780678740`
- `source_url`: [Notion source](https://www.notion.so/llama-cpp-1cb051299a5480a7ab86c6ea97a6adad)
