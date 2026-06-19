---
title: "Pre-Dogfooding App Feedback"
type: "runbook"
slug: "runbooks/app-pre-dogfooding-feedback"
freshness: "2026-04-28T19:57:52Z"
tags:
  - "app"
  - "dogfooding"
  - "feedback"
  - "localization"
owners: []
source_revision_ids:
  - "srcrev_339e4056400ba10d6a2d487bd252cea6"
  - "srcrev_65c9a073751c570521496c83da8763fb"
  - "srcrev_df307c8d318f65ca1eda2bbc7e9dd46e"
  - "srcrev_e1367200b0311e4a50c068c3b6a8748d"
conflict_state: "none"
---

# Pre-Dogfooding App Feedback

## Summary

Collected feedback from pre-dogfooding review of the Story Protocol app, covering localization issues, a placeholder task for IP registration testing, UI placement suggestions, and non‑functional contact button.

## Claims

- In the app, Tamil, Telugu, and Bengali language options display an "English" label instead of their respective language names. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d97482c78a4a39f8aba0ded888f75d6` `source_revision_id=srcrev_339e4056400ba10d6a2d487bd252cea6` `chunk_id=srcchunk_593c1fd0c09fb869b98fb3b1c3f9fa71` `native_locator=slack:C0AL7EKNHDF:1777388397.474779:1777388581.802659` `source_timestamp=2026-04-28T15:06:30Z`
- A task called "Quartz Wave Society" is visible in the app; it is a placeholder name to hide the real campaign for IP registration testing on mainnet, and will be removed soon. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d97482c78a4a39f8aba0ded888f75d6` `source_revision_id=srcrev_339e4056400ba10d6a2d487bd252cea6` `chunk_id=srcchunk_593c1fd0c09fb869b98fb3b1c3f9fa71` `native_locator=slack:C0AL7EKNHDF:1777388397.474779:1777388581.802659` `source_timestamp=2026-04-28T15:06:30Z`
  - citation: `source_document_id=srcdoc_4d97482c78a4a39f8aba0ded888f75d6` `source_revision_id=srcrev_e1367200b0311e4a50c068c3b6a8748d` `chunk_id=srcchunk_132b8f8bc7a4ec70e468cc15cc3e361d` `native_locator=slack:C0AL7EKNHDF:1777388397.474779:1777391818.731589` `source_timestamp=2026-04-28T15:56:58Z`
- Low-priority suggestion: the "Invite Friends" banner would be better placed above the Tasks Landing Page. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d97482c78a4a39f8aba0ded888f75d6` `source_revision_id=srcrev_339e4056400ba10d6a2d487bd252cea6` `chunk_id=srcchunk_593c1fd0c09fb869b98fb3b1c3f9fa71` `native_locator=slack:C0AL7EKNHDF:1777388397.474779:1777388581.802659` `source_timestamp=2026-04-28T15:06:30Z`
- The language selector is displayed at the bottom of the page instead of the top; a fix may already be in progress. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d97482c78a4a39f8aba0ded888f75d6` `source_revision_id=srcrev_339e4056400ba10d6a2d487bd252cea6` `chunk_id=srcchunk_593c1fd0c09fb869b98fb3b1c3f9fa71` `native_locator=slack:C0AL7EKNHDF:1777388397.474779:1777388581.802659` `source_timestamp=2026-04-28T15:06:30Z`
- The "contact" button is non‑functional and should either open an info@ email or be removed. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d97482c78a4a39f8aba0ded888f75d6` `source_revision_id=srcrev_339e4056400ba10d6a2d487bd252cea6` `chunk_id=srcchunk_593c1fd0c09fb869b98fb3b1c3f9fa71` `native_locator=slack:C0AL7EKNHDF:1777388397.474779:1777388581.802659` `source_timestamp=2026-04-28T15:06:30Z`
- Landing page issues have been addressed on staging but the changes have not yet been merged to production. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d97482c78a4a39f8aba0ded888f75d6` `source_revision_id=srcrev_df307c8d318f65ca1eda2bbc7e9dd46e` `chunk_id=srcchunk_a07ab6b76d3eed9e1780199880dec99f` `native_locator=slack:C0AL7EKNHDF:1777388397.474779:1777389391.294429` `source_timestamp=2026-04-28T15:16:31Z`
- A request was made for different images to be included for the tasks in the app. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4d97482c78a4a39f8aba0ded888f75d6` `source_revision_id=srcrev_65c9a073751c570521496c83da8763fb` `chunk_id=srcchunk_09a5203a76e46aec2aba49a0afea61c0` `native_locator=slack:C0AL7EKNHDF:1777388397.474779:1777406272.417109` `source_timestamp=2026-04-28T19:57:52Z`

## Open Questions

- Has the language selector position been updated in a PR?
- When will the staging landing‑page fixes be merged to production?

## Sources

- `source_document_id`: `srcdoc_4d97482c78a4a39f8aba0ded888f75d6`
- `source_revision_id`: `srcrev_e9eba71d39566d9725306e25064650ce`
