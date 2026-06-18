---
title: "TaxBandits Integration"
type: "project"
slug: "projects/taxbandits-integration"
freshness: "2026-06-02T08:48:30Z"
tags:
  - "compliance"
  - "tax-forms"
  - "third-party-integration"
  - "w8-w9"
owners: []
source_revision_ids:
  - "srcrev_3b0cfa2b70f7b2ac3819d4c4049448a3"
  - "srcrev_a55689bb739ed1679be94e62cb746683"
  - "srcrev_b83ef2dce4b2ced05dfc92566499242b"
  - "srcrev_e833a6653bb3c873cfc79320dd8aab26"
conflict_state: "none"
---

# TaxBandits Integration

## Summary

Integration of TaxBandits for W-8 / W-9 form collection, including UX evaluation, sandbox testing, and parallel Stripe exploration.

## Claims

- An MNDA with TaxBandits was signed to enable the integration. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_3b0cfa2b70f7b2ac3819d4c4049448a3` `chunk_id=srcchunk_79ed402d8ca7426dc582df336057cbe6` `native_locator=slack:C0AL7EKNHDF:1779992666.699459:1779992666.699459` `source_timestamp=2026-05-28T18:24:26Z`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_e833a6653bb3c873cfc79320dd8aab26` `chunk_id=srcchunk_1b9119d46d10fe14d5abcb303a07893e` `native_locator=slack:C0AL7EKNHDF:1779994971.364659:1779994971.364659` `source_timestamp=2026-05-28T19:02:51Z`
- A POC page at team-ux.vercel.app showcased the TaxBandits W-9/W-8 form drop-in with a combined chooser and option to load specific forms directly. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_b83ef2dce4b2ced05dfc92566499242b` `chunk_id=srcchunk_86371d252d645f15b993f3044b9e58f8` `native_locator=slack:C0AL7EKNHDF:1780389450.652649:1780389450.652649` `source_timestamp=2026-06-02T08:48:30Z`
- TaxBandits UX was rated best among reviewed options, but improvements were being negotiated. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_b83ef2dce4b2ced05dfc92566499242b` `chunk_id=srcchunk_86371d252d645f15b993f3044b9e58f8` `native_locator=slack:C0AL7EKNHDF:1780389450.652649:1780389450.652649` `source_timestamp=2026-06-02T08:48:30Z`
- Integration on sandbox was planned to begin immediately, with go-live approval targeted for the same week. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_b83ef2dce4b2ced05dfc92566499242b` `chunk_id=srcchunk_86371d252d645f15b993f3044b9e58f8` `native_locator=slack:C0AL7EKNHDF:1780389450.652649:1780389450.652649` `source_timestamp=2026-06-02T08:48:30Z`
- Parallel work on Stripe for W8/W9 and stablecoin rails was in progress, expected to be cheaper at scale. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_a55689bb739ed1679be94e62cb746683` `chunk_id=srcchunk_9b535c60a8559a2530d3347364898159` `native_locator=slack:C0AL7EKNHDF:1780390019.465499:1780390019.465499` `source_timestamp=2026-06-02T08:46:59Z`
- Priority for TaxBandits work was obtaining live account status within 2–3 days and further UX enhancements. `claim:claim_2_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_a55689bb739ed1679be94e62cb746683` `chunk_id=srcchunk_9b535c60a8559a2530d3347364898159` `native_locator=slack:C0AL7EKNHDF:1780390019.465499:1780390019.465499` `source_timestamp=2026-06-02T08:46:59Z`

## Related Pages

- `numo-external-communication-guideline`

## Sources

- `source_document_id`: `srcdoc_02882053947288838246eb0a4d96bb56`
- `source_revision_id`: `srcrev_7d25cc1a4053a5e78027f9bd1b56293e`
