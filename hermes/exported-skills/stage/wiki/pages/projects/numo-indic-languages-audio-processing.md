---
title: "Numo Indic Languages Audio Processing"
type: "project"
slug: "projects/numo-indic-languages-audio-processing"
freshness: "2026-05-23T23:48:12Z"
tags:
  - "asr"
  - "audio-processing"
  - "indic-languages"
  - "numo"
  - "quality"
owners: []
source_revision_ids:
  - "srcrev_4119417e04bdc5646c61f1ff1c10b140"
  - "srcrev_573a503d63f31a4f68ed460056a22411"
  - "srcrev_b7ddbfaa4be2acba6cf512344d2a55c4"
  - "srcrev_c69baccd93b4b6e73fb275483d99d569"
  - "srcrev_d2413a07638be08eb5fad51e70b6327c"
  - "srcrev_decf66ac4537aabc01791d12dddc0ea2"
conflict_state: "none"
---

# Numo Indic Languages Audio Processing

## Summary

Status update on processing 3x Indic languages submissions for Numo, including acceptance rates, cost breakdown, and infrastructure details.

## Claims

- Acceptance rate for Hindi submissions is 68.8% based on ElevenLabs WER/CER passing bar (~18%/16%). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_c69baccd93b4b6e73fb275483d99d569` `chunk_id=srcchunk_2eaab61a0d7ca9d45b6052e550e729ed` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779389000.431729` `source_timestamp=2026-05-21T18:43:20Z`
- Acceptance rate for Telugu submissions is 74.8%. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_c69baccd93b4b6e73fb275483d99d569` `chunk_id=srcchunk_2eaab61a0d7ca9d45b6052e550e729ed` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779389000.431729` `source_timestamp=2026-05-21T18:43:20Z`
- Acceptance rate for Tamil submissions is 52.6%. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_c69baccd93b4b6e73fb275483d99d569` `chunk_id=srcchunk_2eaab61a0d7ca9d45b6052e550e729ed` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779389000.431729` `source_timestamp=2026-05-21T18:43:20Z`
- Processing cost is approximately $2.79 per 1,000 submissions, with ASR API cost comprising 94% and AWS infrastructure 6%. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_c69baccd93b4b6e73fb275483d99d569` `chunk_id=srcchunk_2eaab61a0d7ca9d45b6052e550e729ed` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779389000.431729` `source_timestamp=2026-05-21T18:43:20Z`
- Average audio length for Hindi (first 40,000 submissions) is 43.2 seconds. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_c69baccd93b4b6e73fb275483d99d569` `chunk_id=srcchunk_2eaab61a0d7ca9d45b6052e550e729ed` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779389000.431729` `source_timestamp=2026-05-21T18:43:20Z`
- Processing 40,000 submissions took 3 hours 2 minutes, plus approximately 1 hour for infrastructure spin-up and spin-down. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_c69baccd93b4b6e73fb275483d99d569` `chunk_id=srcchunk_2eaab61a0d7ca9d45b6052e550e729ed` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779389000.431729` `source_timestamp=2026-05-21T18:43:20Z`
- Transcript fidelity is measured using spoken transcript accuracy after normalization, currently applying the standard used by ElevenLabs (WER ~18%, CER ~16%). `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_d2413a07638be08eb5fad51e70b6327c` `chunk_id=srcchunk_ffe4866fd6831b7d170d7ca826b3a916` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779390188.188589` `source_timestamp=2026-05-21T19:03:08Z`
- Deepfake detection filters by Yash are not yet integrated into the processing pipeline, although integration is planned with person on point to integrate them. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_573a503d63f31a4f68ed460056a22411` `chunk_id=srcchunk_81c8307cf883f8544d99ccd8643f0d1a` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779390219.529709` `source_timestamp=2026-05-21T19:03:39Z`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_b7ddbfaa4be2acba6cf512344d2a55c4` `chunk_id=srcchunk_02842fc44004933bd1c5cb2f26afd957` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779493142.709959` `source_timestamp=2026-05-22T23:39:02Z`
- C2PA integration is suggested as a simple first line of defense against AI-generated content, using the open-source SDK to read the manifest to determine if a file is AI-generated. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_decf66ac4537aabc01791d12dddc0ea2` `chunk_id=srcchunk_132df54b22908dcd23bf7e71c97ff518` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779580092.378459` `source_timestamp=2026-05-23T23:48:12Z`
- The current processing phase is solely focused on transcript fidelity, not deepfake detection. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_4119417e04bdc5646c61f1ff1c10b140` `chunk_id=srcchunk_709e7154c56cdea0dad5035f7d320c62` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779389801.536819` `source_timestamp=2026-05-21T18:56:41Z`
  - citation: `source_document_id=srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1` `source_revision_id=srcrev_d2413a07638be08eb5fad51e70b6327c` `chunk_id=srcchunk_ffe4866fd6831b7d170d7ca826b3a916` `native_locator=slack:C0AL7EKNHDF:1779389000.431729:1779390188.188589` `source_timestamp=2026-05-21T19:03:08Z`

## Open Questions

- What benchmark of acceptable error rate does Numo require for transcript fidelity? (Source indicates guidance needed from Numo)

## Sources

- `source_document_id`: `srcdoc_9794cd8cac5a70a901aeaf2fe298fdf1`
- `source_revision_id`: `srcrev_8fb34a28a8f7a9cead65cceecd7e80e0`
