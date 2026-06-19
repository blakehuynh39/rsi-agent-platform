---
title: "Campaign Submission Analytics Events"
type: "system"
slug: "systems/campaign-submission-analytics-events"
freshness: "2026-05-07T17:38:40Z"
tags:
  - "analytics"
  - "ga4"
  - "submission"
owners:
  - "analytics_team"
source_revision_ids:
  - "srcrev_78e3bf1b500fec1406ed2ccdfe38148f"
conflict_state: "none"
---

# Campaign Submission Analytics Events

## Summary

Clarification of the Google Analytics event triggers for campaign submission in the Numo app: button click intent vs. success completion.

## Claims

- Submission button click intent is tracked via the GA event campaign_recording_submit for audio submissions, and campaign_submit_attempt for file/RN helpers (both map to campaign_recording_submit in GA). It fires when the user initiates submission, not only on success. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f6c999ab0c37fac8d389e49430c295c2` `source_revision_id=srcrev_78e3bf1b500fec1406ed2ccdfe38148f` `chunk_id=srcchunk_c4c1c160c6c782fea9eed2ce04c3f836` `native_locator=slack:C0AL7EKNHDF:1778085693.849329:1778175520.751499` `source_timestamp=2026-05-07T17:38:40Z`
- Successful submission completion is tracked via campaign_recording_success (or campaign_submit_success on React Native), which fires onSuccess after the mutation finishes, not on button tap. Both map to campaign_submission in GA4. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f6c999ab0c37fac8d389e49430c295c2` `source_revision_id=srcrev_78e3bf1b500fec1406ed2ccdfe38148f` `chunk_id=srcchunk_c4c1c160c6c782fea9eed2ce04c3f836` `native_locator=slack:C0AL7EKNHDF:1778085693.849329:1778175520.751499` `source_timestamp=2026-05-07T17:38:40Z`

## Sources

- `source_document_id`: `srcdoc_f6c999ab0c37fac8d389e49430c295c2`
- `source_revision_id`: `srcrev_78e3bf1b500fec1406ed2ccdfe38148f`
