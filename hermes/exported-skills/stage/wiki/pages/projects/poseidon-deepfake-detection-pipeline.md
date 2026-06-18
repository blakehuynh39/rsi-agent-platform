---
title: "Poseidon Deepfake Detection Pipeline"
type: "project"
slug: "projects/poseidon-deepfake-detection-pipeline"
freshness: "2026-06-17T16:18:36Z"
tags:
  - "audio-classification"
  - "deepfake-detection"
  - "ensemble-models"
  - "poseidon"
owners:
  - "Poseidon team"
source_revision_ids:
  - "srcrev_a1dfd1a2cd1670b257386a7e04184369"
  - "srcrev_a860659250b2cd65f0a671c1e246b402"
  - "srcrev_c672b80f5d107c8fdf1404a6a5eb91a3"
  - "srcrev_f483858687a5ba792eba6869f5413d26"
  - "srcrev_fae5749beb8d3e0f2a8e2243ff5b92a7"
conflict_state: "none"
---

# Poseidon Deepfake Detection Pipeline

## Summary

The pipeline for detecting deepfake audio submissions in Poseidon's Season 1 voice collection campaign. Early results identified a spammer using deepfakes that scored highly, highlighting challenges. The team is exploring ensemble models and feedback loops with user behavior data.

## Claims

- A spammer who joined all language tasks and used deepfakes still achieved a 97% score after deepfake and WER analysis in the Poseidon validation pipeline. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_10755201d86671d2e8612c60bed54737` `source_revision_id=srcrev_c672b80f5d107c8fdf1404a6a5eb91a3` `chunk_id=srcchunk_7d9b3fdb1e7db34d051158ff02c81d39` `native_locator=slack:C0AL7EKNHDF:1781671196.306859:1781671196.306859` `source_timestamp=2026-06-17T04:39:56Z`
- User in-app behavior is important for classification, and data from misclassified users should be sent back to Poseidon to refine future models. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_10755201d86671d2e8612c60bed54737` `source_revision_id=srcrev_a860659250b2cd65f0a671c1e246b402` `chunk_id=srcchunk_cce7af492fc54a1d4c12f4777ccef508` `native_locator=slack:C0AL7EKNHDF:1781671196.306859:1781672503.406979` `source_timestamp=2026-06-17T05:12:23Z`
- A suggestion was made to add metadata or use an ensemble of classifiers to improve deepfake detection. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_10755201d86671d2e8612c60bed54737` `source_revision_id=srcrev_f483858687a5ba792eba6869f5413d26` `chunk_id=srcchunk_a405e69102e49b2da6ee18b91a1cb832` `native_locator=slack:C0AL7EKNHDF:1781671196.306859:1781675682.029489` `source_timestamp=2026-06-17T05:54:42Z`
- Ensemble models were under development; inferencing using ensemble models on Numo data is planned by the end of the week. The current best model is SVM, using only audio metrics without user metadata, and ensemble models will also not use user metadata. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_10755201d86671d2e8612c60bed54737` `source_revision_id=srcrev_a1dfd1a2cd1670b257386a7e04184369` `chunk_id=srcchunk_41c8f7ced606d336224b1c5020a06676` `native_locator=slack:C0AL7EKNHDF:1781671196.306859:1781675690.365129` `source_timestamp=2026-06-17T05:54:50Z`
  - citation: `source_document_id=srcdoc_10755201d86671d2e8612c60bed54737` `source_revision_id=srcrev_fae5749beb8d3e0f2a8e2243ff5b92a7` `chunk_id=srcchunk_4afed82c7c5c41cc23452d91cc8c2457` `native_locator=slack:C0AL7EKNHDF:1781671196.306859:1781713116.303719` `source_timestamp=2026-06-17T16:18:36Z`

## Open Questions

- How can user in-app behavior be integrated into the classification pipeline?
- When will inferencing on Numo data start?
- Will ensemble models improve deepfake detection accuracy?

## Sources

- `source_document_id`: `srcdoc_10755201d86671d2e8612c60bed54737`
- `source_revision_id`: `srcrev_fae5749beb8d3e0f2a8e2243ff5b92a7`
