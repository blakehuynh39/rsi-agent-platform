---
title: "Poseidon Season 1 Infra Lessons"
type: "runbook"
slug: "runbooks/poseidon-season1-infra-lessons"
freshness: "2026-04-22T01:43:41Z"
tags:
  - "infra"
  - "lessons-learned"
  - "poseidon"
  - "season1"
  - "validation"
owners:
  - "Product Team"
source_revision_ids:
  - "srcrev_073abf75ed71a8c599007e95a184083e"
  - "srcrev_1cdd2c21ecfdfb235ef6f0ceb5d7f62b"
  - "srcrev_b818d5eac367f77a6281e6126c37cff0"
  - "srcrev_de93e3e0d6efba4c8cc6cb95b6266640"
conflict_state: "none"
---

# Poseidon Season 1 Infra Lessons

## Summary

Lessons learned from Poseidon Season 1 infrastructure and data validation pipeline, shared asynchronously in preparation for Season 2. Covers bottlenecks, audio requirements, IP registration scaling, status communication, system limits, scaling strategy, and fallback messaging.

## Claims

- The majority of the bottleneck in the old data validation pipeline was rate limiting on the service provider side (HuggingFace API, OpenAI, etc.). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_b818d5eac367f77a6281e6126c37cff0` `chunk_id=srcchunk_3d6541cf6f1397bc52785a8f62edaf71` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776814813.773579` `source_timestamp=2026-04-21T23:40:13Z`
- Implementing async validation gives more flexibility on the infra side. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_b818d5eac367f77a6281e6126c37cff0` `chunk_id=srcchunk_3d6541cf6f1397bc52785a8f62edaf71` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776814813.773579` `source_timestamp=2026-04-21T23:40:13Z`
- Need raw unprocessed audio for recordings, requiring disabling browser’s default audio processing. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_b818d5eac367f77a6281e6126c37cff0` `chunk_id=srcchunk_3d6541cf6f1397bc52785a8f62edaf71` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776814813.773579` `source_timestamp=2026-04-21T23:40:13Z`
- IP registration was slow due to sequential registration with a single wallet because the Python SDK waits for transaction confirmation. Resolved by using 40 wallets in parallel. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_b818d5eac367f77a6281e6126c37cff0` `chunk_id=srcchunk_3d6541cf6f1397bc52785a8f62edaf71` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776814813.773579` `source_timestamp=2026-04-21T23:40:13Z`
- Clear communication on file status is crucial; distinguish system failure from audio quality issues to prevent user frustration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_b818d5eac367f77a6281e6126c37cff0` `chunk_id=srcchunk_3d6541cf6f1397bc52785a8f62edaf71` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776814813.773579` `source_timestamp=2026-04-21T23:40:13Z`
- Know the system limits for processing audio files/queries per second. Season 1 QPS was approximately 500; expected to be higher in Season 2. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_1cdd2c21ecfdfb235ef6f0ceb5d7f62b` `chunk_id=srcchunk_fb88929b8a06f70c01d10168bf8438c7` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776816204.759039` `source_timestamp=2026-04-22T00:03:24Z`
- Do not horizontally scale out services all at once; service cold starts (e.g., multiple open database connections) can cause additional system bottlenecks. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_1cdd2c21ecfdfb235ef6f0ceb5d7f62b` `chunk_id=srcchunk_fb88929b8a06f70c01d10168bf8438c7` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776816204.759039` `source_timestamp=2026-04-22T00:03:24Z`
- Implement a worst-case fallback for users (e.g., no more than N pending jobs per user) with messaging when the backend server fails. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_1cdd2c21ecfdfb235ef6f0ceb5d7f62b` `chunk_id=srcchunk_fb88929b8a06f70c01d10168bf8438c7` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776816204.759039` `source_timestamp=2026-04-22T00:03:24Z`
- Sandeep has written a comprehensive doc 'How to Run Season 2' on Notion containing many todos. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_de93e3e0d6efba4c8cc6cb95b6266640` `chunk_id=srcchunk_5f5391d74edee06e37df4debf3726e83` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776822152.243199` `source_timestamp=2026-04-22T01:42:32Z`
- Low communication efficiency was observed; need a single source of truth (SOT) doc for the Numo app covering app, backend, and everything, that everyone must read. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8` `source_revision_id=srcrev_073abf75ed71a8c599007e95a184083e` `chunk_id=srcchunk_6cf21be320fba920ce1ec5ca17876096` `native_locator=slack:C0AL7EKNHDF:1776811856.779369:1776822221.610459` `source_timestamp=2026-04-22T01:43:41Z`

## Open Questions

- Is there a postmortem doc and where is it?
- What will be the expected QPS for Season 2?

## Related Pages

- `https-/www-notion-so/how-to-run-season-2-2b25654de20e800ea6cbd51827ad930b`

## Sources

- `source_document_id`: `srcdoc_a7637208f0c9c2ac582aa4cc2cbbf4e8`
- `source_revision_id`: `srcrev_073abf75ed71a8c599007e95a184083e`
