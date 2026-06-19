---
title: "Transcript Generation Testing Approach"
type: "decision"
slug: "decisions/transcript-generation-testing-approach"
freshness: "2026-06-08T14:29:05Z"
tags:
  - "model-failure"
  - "testing"
  - "transcript-generation"
owners:
  - "U0A2D9U625V"
source_revision_ids:
  - "srcrev_90787c47e7f254b4520edbbb96be8e39"
  - "srcrev_cafb346e0e5ff55d75960d39a3864251"
conflict_state: "none"
---

# Transcript Generation Testing Approach

## Summary

Decision to select transcripts where current models fail, rather than randomly, for testing transcript generation.

## Claims

- For transcript generation testing, transcripts should be chosen where current models fail, not randomly. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2adf1a6bfcea20d2865abbb74e510f0` `source_revision_id=srcrev_90787c47e7f254b4520edbbb96be8e39` `chunk_id=srcchunk_2ed6972c9851ebe65498795f53ea7f5c` `native_locator=slack:C0AL7EKNHDF:1780883901.746259:1780883901.746259` `source_timestamp=2026-06-08T01:58:21Z`
- Testing should be done against current models for transcript generation. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2adf1a6bfcea20d2865abbb74e510f0` `source_revision_id=srcrev_cafb346e0e5ff55d75960d39a3864251` `chunk_id=srcchunk_5e63b000a4e0957fb20b805285059ad6` `native_locator=slack:C0AL7EKNHDF:1780883901.746259:1780928945.604989` `source_timestamp=2026-06-08T14:29:05Z`

## Sources

- `source_document_id`: `srcdoc_e2adf1a6bfcea20d2865abbb74e510f0`
- `source_revision_id`: `srcrev_cafb346e0e5ff55d75960d39a3864251`
