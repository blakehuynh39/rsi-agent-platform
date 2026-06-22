---
title: "TaxBandits Integration"
type: "project"
slug: "projects/taxbandits-integration"
freshness: "2026-06-02T08:48:30Z"
tags:
  - "integration"
  - "stablecoin"
  - "tax-compliance"
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

Integration with TaxBandits for W-8/W-9 tax form collection. Includes MNDA signing, POC UX testing, and go-live preparations. Also exploring Stripe and stable coin rails for cheaper payments at scale.

## Claims

- Upcoming meeting with TaxBandits required signing an MNDA; request for approval to sign sent. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_3b0cfa2b70f7b2ac3819d4c4049448a3` `chunk_id=srcchunk_79ed402d8ca7426dc582df336057cbe6` `native_locator=slack:C0AL7EKNHDF:1779992666.699459:1779992666.699459` `source_timestamp=2026-05-28T18:24:26Z`
- Approval given to sign the TaxBandits MNDA. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_e833a6653bb3c873cfc79320dd8aab26` `chunk_id=srcchunk_1b9119d46d10fe14d5abcb303a07893e` `native_locator=slack:C0AL7EKNHDF:1779994971.364659:1779994971.364659` `source_timestamp=2026-05-28T19:02:51Z`
- A POC test page for TaxBandits W-9/W-8 form drop-in is hosted at https://team-ux.vercel.app. It defaults to a combined chooser for W-8 or W-9, with option to load a specific form. Use sandbox (dummy data) and submit. TaxBandits has the best UX among reviewed options, though not perfect. A note banner is used at the top. If UX is acceptable, sandbox integration can start, while go-live approval is sought this week. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_b83ef2dce4b2ced05dfc92566499242b` `chunk_id=srcchunk_86371d252d645f15b993f3044b9e58f8` `native_locator=slack:C0AL7EKNHDF:1780389450.652649:1780389450.652649` `source_timestamp=2026-06-02T08:48:30Z`
- Current focus on W-8/W-9: get live status for the account ASAP (target 2-3 days), improve UX with TaxBandits, and parallel threads on Stripe for W-8/W-9 and stable coin rails (much cheaper at scale). `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_02882053947288838246eb0a4d96bb56` `source_revision_id=srcrev_a55689bb739ed1679be94e62cb746683` `chunk_id=srcchunk_9b535c60a8559a2530d3347364898159` `native_locator=slack:C0AL7EKNHDF:1780390019.465499:1780390019.465499` `source_timestamp=2026-06-02T08:46:59Z`

## Open Questions

- Is payment via on-chain transactions? ('on-chain tx?' asked in channel)
- What is stable coin rail setup?

## Related Pages

- `finding-numo`
- `numo`

## Sources

- `source_document_id`: `srcdoc_02882053947288838246eb0a4d96bb56`
- `source_revision_id`: `srcrev_82a245207902de03a8cf942b5b68cc8c`
