---
title: "Payment Flow UI Review"
type: "project"
slug: "projects/payment-flow-ui-review"
freshness: "2026-05-15T21:30:13Z"
tags:
  - "design"
  - "payment"
  - "staging"
  - "ui"
owners:
  - "U04L0DD6B6F"
  - "U06A5AQ1VD3"
  - "U0AU3DWLVE2"
source_revision_ids:
  - "srcrev_3f3b7c8036e397e0c7c7ffe113a9d7b6"
  - "srcrev_3f5834c30363709c87104a0af7fdc6da"
  - "srcrev_41b1ffc81f15cb1614426b74bc5193bf"
  - "srcrev_67d549b87b1cd23b633d9f11790babf6"
  - "srcrev_76b89b765411629d2c0c6ad333cd3e8e"
  - "srcrev_866b89f1d722db64524c384c0cbdaf9f"
conflict_state: "none"
---

# Payment Flow UI Review

## Summary

Design review for the payment flow, covering UI improvements to button layout and balance breakdown display. The flow is functional on staging with Stripe sandbox but needs design refinement, including reducing primary buttons and reorganizing the balance breakdown section.

## Claims

- End-to-end payment flow is working on staging and connected to Stripe sandbox. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d0c71b81f504f48dc42b8245fe8b511` `source_revision_id=srcrev_67d549b87b1cd23b633d9f11790babf6` `chunk_id=srcchunk_95a886c388572c24ca53de885f2b070d` `native_locator=slack:C0AL7EKNHDF:1778792890.341159:1778792890.341159` `source_timestamp=2026-05-14T21:08:10Z`
- Current payment flow UI is rough and requires design improvement. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d0c71b81f504f48dc42b8245fe8b511` `source_revision_id=srcrev_67d549b87b1cd23b633d9f11790babf6` `chunk_id=srcchunk_95a886c388572c24ca53de885f2b070d` `native_locator=slack:C0AL7EKNHDF:1778792890.341159:1778792890.341159` `source_timestamp=2026-05-14T21:08:10Z`
- The UI currently has two primary buttons; feedback says there should be only one primary button. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d0c71b81f504f48dc42b8245fe8b511` `source_revision_id=srcrev_41b1ffc81f15cb1614426b74bc5193bf` `chunk_id=srcchunk_23155c2ce08b8f38457466508a11c58e` `native_locator=slack:C0AL7EKNHDF:1778792890.341159:1778793576.452009` `source_timestamp=2026-05-14T21:19:36Z`
- The 'Balance breakdown' title is confusing because it resembles total earnings history rather than a current balance breakdown. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d0c71b81f504f48dc42b8245fe8b511` `source_revision_id=srcrev_41b1ffc81f15cb1614426b74bc5193bf` `chunk_id=srcchunk_23155c2ce08b8f38457466508a11c58e` `native_locator=slack:C0AL7EKNHDF:1778792890.341159:1778793576.452009` `source_timestamp=2026-05-14T21:19:36Z`
- The demo screencast has no audio; it is a silent screen recording. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d0c71b81f504f48dc42b8245fe8b511` `source_revision_id=srcrev_866b89f1d722db64524c384c0cbdaf9f` `chunk_id=srcchunk_fb3f5cdee0b5a03f05cd471a0613bbcc` `native_locator=slack:C0AL7EKNHDF:1778792890.341159:1778793861.557309` `source_timestamp=2026-05-14T21:24:21Z`
- Proposal: Show only 'your balance' and pending balance at top; move balance breakdown to a separate 'view history' page or make it collapsible. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d0c71b81f504f48dc42b8245fe8b511` `source_revision_id=srcrev_3f5834c30363709c87104a0af7fdc6da` `chunk_id=srcchunk_63d03951eea1bbac542dd60752b95680` `native_locator=slack:C0AL7EKNHDF:1778792890.341159:1778795204.093069` `source_timestamp=2026-05-14T21:46:44Z`
- Design mocks and a Figma file have been requested for the payment flow. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d0c71b81f504f48dc42b8245fe8b511` `source_revision_id=srcrev_76b89b765411629d2c0c6ad333cd3e8e` `chunk_id=srcchunk_d7277de87018a00575198cc90cf9b3dd` `native_locator=slack:C0AL7EKNHDF:1778792890.341159:1778878340.359299` `source_timestamp=2026-05-15T20:52:53Z`
- U0AU3DWLVE2 has added the design work to their list and will work on it. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d0c71b81f504f48dc42b8245fe8b511` `source_revision_id=srcrev_3f3b7c8036e397e0c7c7ffe113a9d7b6` `chunk_id=srcchunk_38e911ad4a09d15c78f338a282c7bb79` `native_locator=slack:C0AL7EKNHDF:1778792890.341159:1778880613.807229` `source_timestamp=2026-05-15T21:30:13Z`

## Open Questions

- What will the final design for the balance breakdown section be?
- When will the design mocks be provided?
- Who will implement the UI changes?

## Sources

- `source_document_id`: `srcdoc_6d0c71b81f504f48dc42b8245fe8b511`
- `source_revision_id`: `srcrev_3f3b7c8036e397e0c7c7ffe113a9d7b6`
