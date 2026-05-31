---
title: "DePIN App Onboarding and Voice Profile Setup"
type: "runbook"
slug: "runbooks/depin-app-onboarding-and-voice-profile-setup"
freshness: "2025-08-21T22:32:00Z"
tags:
  - "depin"
  - "onboarding"
  - "voice-profile"
owners: []
source_revision_ids:
  - "srcrev_6699c29c39b069a4a111c4fd5eb8bdb8"
conflict_state: "none"
---

# DePIN App Onboarding and Voice Profile Setup

## Summary

Step-by-step guide for user onboarding and voice profile creation in the DePIN app, covering welcome screen, terms acceptance, login, onboarding screens, microphone permission, recording, and completion.

## Claims

- The onboarding begins with a welcome screen showing a checkbox for Terms & Privacy, with a greyed-out 'Get Started' button until the checkbox is checked. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- Tapping 'Get Started' without checking the Terms & Privacy box prevents the user from proceeding. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- Checking the box and then tapping 'Get Started' takes the user to the Login screen. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- Login offers 'Continue with Email', which sends a verification code to the user's email, or 'Continue with Google' for direct sign-in via Google account selection. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- After login, onboarding screens with animations, a skip button, and dot indicators are displayed; tapping Next through all screens leads to 'Create your voice profile'. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- Voice profile creation starts by tapping 'Allow Microphone'; denying access shows an error screen stating microphone must be enabled to continue. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- After allowing microphone access, the user is shown a 'Get ready to record' screen with tips, and then a 'Read these words aloud' screen with a Start Recording button. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- Tapping Start Recording and immediately stopping without speaking results in submission failure and a prompt to retry. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- After a proper recording is stopped, the user is shown options: Preview (to hear their recording), Retry (to re-record), and Submit. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`
- Tapping Submit completes the voice profile and shows a 'Voice profile complete' screen with a unique orb and a CTA to 'Enter Poseidon'. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e) `source_document_id=srcdoc_5ff502af8d900c47cfc3d65b69c2287c` `source_revision_id=srcrev_6699c29c39b069a4a111c4fd5eb8bdb8` `chunk_id=srcchunk_24534300b67bb68dbc00e980b243e115` `native_locator=https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e` `source_timestamp=2025-08-21T22:32:00Z`

## Sources

- `source_document_id`: `srcdoc_5ff502af8d900c47cfc3d65b69c2287c`
- `source_revision_id`: `srcrev_6699c29c39b069a4a111c4fd5eb8bdb8`
- `source_url`: [Notion source](https://www.notion.so/DePIN-App-User-Guide-256051299a5480378763fdff9aa2993e)
