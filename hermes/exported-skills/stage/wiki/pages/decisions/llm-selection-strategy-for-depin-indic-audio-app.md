---
title: "LLM Selection Strategy for DePIN Indic Audio App"
type: "decision"
slug: "decisions/llm-selection-strategy-for-depin-indic-audio-app"
freshness: "2026-03-24T20:44:06Z"
tags:
  - "audio-app"
  - "depin"
  - "indic-languages"
  - "llm-strategy"
  - "model-selection"
owners:
  - "U09QGMMUDPC"
source_revision_ids:
  - "srcrev_3bee3ffc2e3de6137a7fa87d694eed79"
  - "srcrev_85c99c089c4133249e33a804f4fef619"
  - "srcrev_c29a6c0f489a54e1a637eda201d563aa"
  - "srcrev_e59362d5384e62dc0315e5f22ffb7b33"
conflict_state: "none"
---

# LLM Selection Strategy for DePIN Indic Audio App

## Summary

A strategy document to select separate speech-to-text (STT), large language model (LLM), and text-to-speech (TTS) models for low-resource Indic languages, covering deployment options (cloud API vs self-hosted vs on-device), cost math, and a phased implementation approach from proof-of-concept to scaled production.

## Claims

- The LLM Selection Strategy document is ready for review and hosted in Google Docs at https://docs.google.com/document/d/1NiHix4FUO5wwmxaa5PT7LGxtpderp4tNVdWEOTMH52M/edit. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77b22ef58f2b362fb9868822e093f391` `source_revision_id=srcrev_3bee3ffc2e3de6137a7fa87d694eed79` `chunk_id=srcchunk_43271a2aeba1ca226c060569984d24bf` `native_locator=slack:C0AL7EKNHDF:1774379349.255519:1774379349.255519` `source_timestamp=2026-03-24T19:09:09Z`
- Because no single model covers all tasks for low-resource Indic languages, the strategy requires separate model selections for STT, text generation, and TTS. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77b22ef58f2b362fb9868822e093f391` `source_revision_id=srcrev_3bee3ffc2e3de6137a7fa87d694eed79` `chunk_id=srcchunk_43271a2aeba1ca226c060569984d24bf` `native_locator=slack:C0AL7EKNHDF:1774379349.255519:1774379349.255519` `source_timestamp=2026-03-24T19:09:09Z`
- The document includes deployment strategy for each task (cloud API vs self-hosted vs on-device) and cost calculations. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77b22ef58f2b362fb9868822e093f391` `source_revision_id=srcrev_3bee3ffc2e3de6137a7fa87d694eed79` `chunk_id=srcchunk_43271a2aeba1ca226c060569984d24bf` `native_locator=slack:C0AL7EKNHDF:1774379349.255519:1774379349.255519` `source_timestamp=2026-03-24T19:09:09Z`
- Section 7 defines a phased stack: models usable immediately for proof-of-concept versus models to migrate to at scale. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77b22ef58f2b362fb9868822e093f391` `source_revision_id=srcrev_3bee3ffc2e3de6137a7fa87d694eed79` `chunk_id=srcchunk_43271a2aeba1ca226c060569984d24bf` `native_locator=slack:C0AL7EKNHDF:1774379349.255519:1774379349.255519` `source_timestamp=2026-03-24T19:09:09Z`
- The author plans to begin hands-on testing with shortlisted models: Sarvam Saaras v3 (STT), Gemini Flash (LLM), and Bulbul v3 (TTS) against Tamil, Telugu, Gujarati, and Bengali audio. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77b22ef58f2b362fb9868822e093f391` `source_revision_id=srcrev_3bee3ffc2e3de6137a7fa87d694eed79` `chunk_id=srcchunk_43271a2aeba1ca226c060569984d24bf` `native_locator=slack:C0AL7EKNHDF:1774379349.255519:1774379349.255519` `source_timestamp=2026-03-24T19:09:09Z`
- The strategy document has been added to the Notion workspace under the 'Tiger Team' page. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77b22ef58f2b362fb9868822e093f391` `source_revision_id=srcrev_e59362d5384e62dc0315e5f22ffb7b33` `chunk_id=srcchunk_2dbf00f1de0a7cbb09f7f56292f84c88` `native_locator=slack:C0AL7EKNHDF:1774379349.255519:1774384708.318829` `source_timestamp=2026-03-24T20:38:28Z`
- Latency data for the shortlisted models is not yet available; the author intends to test them and report back. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77b22ef58f2b362fb9868822e093f391` `source_revision_id=srcrev_c29a6c0f489a54e1a637eda201d563aa` `chunk_id=srcchunk_0d7cbd97bf728b2d9325db7ee0cb37db` `native_locator=slack:C0AL7EKNHDF:1774379349.255519:1774385046.658489` `source_timestamp=2026-03-24T20:44:06Z`
  - citation: `source_document_id=srcdoc_77b22ef58f2b362fb9868822e093f391` `source_revision_id=srcrev_85c99c089c4133249e33a804f4fef619` `chunk_id=srcchunk_924007bbdc54a04029efef880322d445` `native_locator=slack:C0AL7EKNHDF:1774379349.255519:1774384994.665129` `source_timestamp=2026-03-24T20:43:26Z`

## Open Questions

- What are the actual latency figures for Sarvam Saaras v3, Gemini Flash, and Bulbul v3 on the target Indic languages?

## Sources

- `source_document_id`: `srcdoc_77b22ef58f2b362fb9868822e093f391`
- `source_revision_id`: `srcrev_c29a6c0f489a54e1a637eda201d563aa`
