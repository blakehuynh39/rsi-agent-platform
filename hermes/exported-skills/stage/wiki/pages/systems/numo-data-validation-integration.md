---
title: "Numo Data Validation Integration"
type: "system"
slug: "systems/numo-data-validation-integration"
freshness: "2026-06-05T03:27:33Z"
tags:
  - "bot"
  - "data-validation"
  - "deepfake"
  - "numo"
  - "payout"
owners:
  - "U0A2D9U625V"
  - "U0AKN387CMS"
source_revision_ids:
  - "srcrev_15e0ac399cd5bafb2476bd6b0ddd1ba4"
  - "srcrev_20470fb94d41ba34a434185c611a3570"
  - "srcrev_21d266c4d0978922b7599c57a44aa292"
  - "srcrev_92a08a0622982911f7ddc342cc6be077"
  - "srcrev_95f7e8544d8939bc309299cd9084cd90"
  - "srcrev_a9417c905b904fab35b7ccff6e4b264b"
  - "srcrev_bb9798b09e932887f590c302bd1891b8"
  - "srcrev_d2f8c1eec0f556e20838144e96ad32e7"
  - "srcrev_df318d3a4420bd1e3544ac253c3573aa"
conflict_state: "none"
---

# Numo Data Validation Integration

## Summary

Integration between Numo and the data validation system (DVP). Handles submission review, validation, deepfake detection, and publishing results back to Numo API for payout. Currently blocked on deepfake results before publishing final validation results.

## Claims

- There are 746k submissions pending review. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_d2f8c1eec0f556e20838144e96ad32e7` `chunk_id=srcchunk_6dff1f629dcf64e91f63b25ff4de4f18` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780621257.609609` `source_timestamp=2026-06-05T01:00:57Z`
- The bot does not know if submissions have been validated; progress can be viewed at https://validate.psdn.app/numo. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_95f7e8544d8939bc309299cd9084cd90` `chunk_id=srcchunk_fd5f84314fc5988b1777f7f8c1454964` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780621618.736129` `source_timestamp=2026-06-05T01:06:58Z`
- Apart from the latest Tamil submissions over the past 1-2 days, the rest have already gone through validation. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_95f7e8544d8939bc309299cd9084cd90` `chunk_id=srcchunk_fd5f84314fc5988b1777f7f8c1454964` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780621618.736129` `source_timestamp=2026-06-05T01:06:58Z`
- The missing piece is sending the results back to the Numo API, which the bot tracks from the Numo database. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_21d266c4d0978922b7599c57a44aa292` `chunk_id=srcchunk_f113c7d5616a52ed0c9f1dc5c06a8104` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780621679.285889` `source_timestamp=2026-06-05T01:07:59Z`
- Sending results back would result in finalized reviews for payout. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_df318d3a4420bd1e3544ac253c3573aa` `chunk_id=srcchunk_c1b11a074c296edd4d900b116b890fdf` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780621691.650239` `source_timestamp=2026-06-05T01:08:11Z`
- A Notion page for integration details is at https://app.notion.com/p/Numo-Data-Validation-App-Integration-3535654de20e80e9aac4eebcb8e525de. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_a9417c905b904fab35b7ccff6e4b264b` `chunk_id=srcchunk_e93cc96fa4769f04773b6521508b3920` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780621747.638219` `source_timestamp=2026-06-05T01:09:07Z`
- They have not yet published any results back to Numo; they intend to combine deepfake results before doing so. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_20470fb94d41ba34a434185c611a3570` `chunk_id=srcchunk_ea51117bdfa6386d6e64c217239ca6fc` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780621748.359969` `source_timestamp=2026-06-05T01:09:08Z`
- The process is currently blocked on deepfake (DF) results. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_92a08a0622982911f7ddc342cc6be077` `chunk_id=srcchunk_3da8373a9aca75a4c5882d346087414d` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780621762.135609` `source_timestamp=2026-06-05T01:09:22Z`
- It is suggested that the bot’s reports should be made accurate, include deepfake results once ready, and not mark items as validated before all results are in to avoid confusion. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_15e0ac399cd5bafb2476bd6b0ddd1ba4` `chunk_id=srcchunk_539442ace20c164d8d985d3875f6a921` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780629508.991259` `source_timestamp=2026-06-05T03:18:28Z`
- This suggestion was agreed upon. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5` `source_revision_id=srcrev_bb9798b09e932887f590c302bd1891b8` `chunk_id=srcchunk_d412e0d7fe5ac65d8bd75f3224aa8b00` `native_locator=slack:C0AL7EKNHDF:1780621257.609609:1780630053.092969` `source_timestamp=2026-06-05T03:27:33Z`

## Open Questions

- What is the exact process for combining deepfake results before publishing to Numo?
- When will deepfake results be available for combination?

## Related Pages

- `numo-data-validation-app`

## Sources

- `source_document_id`: `srcdoc_2ccd318ef4f027f3542d5c1b0208d3b5`
- `source_revision_id`: `srcrev_bb9798b09e932887f590c302bd1891b8`
