---
title: "Referral Automation Risk Signal"
type: "decision"
slug: "decisions/referral-automation-risk-signal"
freshness: "2026-06-17T20:34:00Z"
tags:
  - "fraud"
  - "referrals"
  - "risk"
owners:
  - "U04L0DD6B6F"
source_revision_ids:
  - "srcrev_3384274a117d23ca465c6be9973d3b84"
  - "srcrev_c177330adc8bffbe8f20d8cc0e05454f"
conflict_state: "none"
---

# Referral Automation Risk Signal

## Summary

A user a9baf1af-f6fa-4aa2-a54c-a664548fdecb generated 4676 referrals in 1-2 days, likely via script. This pattern should be added as a risk signal and included in the payout-obligations table.

## Claims

- User a9baf1af-f6fa-4aa2-a54c-a664548fdecb had 4676 referrals in 1-2 days, suggesting automated script usage. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1d1510911b7c2ff745b2a59a28518fd1` `source_revision_id=srcrev_3384274a117d23ca465c6be9973d3b84` `chunk_id=srcchunk_6b183ff025e15d35d3b265343be824d4` `native_locator=slack:C0AL7EKNHDF:1781725068.898829:1781725068.898829` `source_timestamp=2026-06-17T19:37:48Z`
- This pattern should be added as a risk signal and included in the payout-obligations table. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1d1510911b7c2ff745b2a59a28518fd1` `source_revision_id=srcrev_3384274a117d23ca465c6be9973d3b84` `chunk_id=srcchunk_6b183ff025e15d35d3b265343be824d4` `native_locator=slack:C0AL7EKNHDF:1781725068.898829:1781725068.898829` `source_timestamp=2026-06-17T19:37:48Z`
- Another user 6e7cd118-0bd5-4652-b88f-efffd2f13b29 was also mentioned as potentially suspicious. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1d1510911b7c2ff745b2a59a28518fd1` `source_revision_id=srcrev_c177330adc8bffbe8f20d8cc0e05454f` `chunk_id=srcchunk_fcda48b230f7bb4561d8d1fb2e030c67` `native_locator=slack:C0AL7EKNHDF:1781725068.898829:1781728440.899159` `source_timestamp=2026-06-17T20:34:00Z`

## Open Questions

- Are there other users exhibiting similar referral patterns?
- How to implement this risk signal?
- What threshold of referrals per day should trigger a risk flag?

## Sources

- `source_document_id`: `srcdoc_1d1510911b7c2ff745b2a59a28518fd1`
- `source_revision_id`: `srcrev_c177330adc8bffbe8f20d8cc0e05454f`
