---
title: "Frontend Localization Runbook"
type: "runbook"
slug: "runbooks/frontend-localization-with-lingui"
freshness: "2026-04-03T13:25:17Z"
tags:
  - "elevenlabs"
  - "frontend"
  - "lingui"
  - "localization"
  - "runbook"
owners: []
source_revision_ids:
  - "srcrev_ad7b7d413d9c1c8c1be9bcf326f97b6f"
conflict_state: "none"
---

# Frontend Localization Runbook

## Summary

Step-by-step process to add translations using lingui and configure ElevenLabs agents for voice localization.

## Claims

- The frontend localization uses the js-lingui library, built into poseidon szn 2. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_653c9dfb8bd5695482ab788734e5908b` `source_revision_id=srcrev_ad7b7d413d9c1c8c1be9bcf326f97b6f` `chunk_id=srcchunk_8c5f7636ae3f8684db8b2030e4d9329b` `native_locator=slack:C0AL7EKNHDF:1775153012.510199:1775222717.712879` `source_timestamp=2026-04-03T13:25:17Z`
- To add a translatable string, import Trans and useLingui from @lingui/react/macro and use t`...` or <Trans>...</Trans>. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_653c9dfb8bd5695482ab788734e5908b` `source_revision_id=srcrev_ad7b7d413d9c1c8c1be9bcf326f97b6f` `chunk_id=srcchunk_8c5f7636ae3f8684db8b2030e4d9329b` `native_locator=slack:C0AL7EKNHDF:1775153012.510199:1775222717.712879` `source_timestamp=2026-04-03T13:25:17Z`
- Run `pnpm lingui:extract` to generate .po files in src/locales/. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_653c9dfb8bd5695482ab788734e5908b` `source_revision_id=srcrev_ad7b7d413d9c1c8c1be9bcf326f97b6f` `chunk_id=srcchunk_8c5f7636ae3f8684db8b2030e4d9329b` `native_locator=slack:C0AL7EKNHDF:1775153012.510199:1775222717.712879` `source_timestamp=2026-04-03T13:25:17Z`
- Fill in the empty msgstr strings in the .po file with translations. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_653c9dfb8bd5695482ab788734e5908b` `source_revision_id=srcrev_ad7b7d413d9c1c8c1be9bcf326f97b6f` `chunk_id=srcchunk_8c5f7636ae3f8684db8b2030e4d9329b` `native_locator=slack:C0AL7EKNHDF:1775153012.510199:1775222717.712879` `source_timestamp=2026-04-03T13:25:17Z`
- Run `pnpm lingui:compile` to finalize translations. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_653c9dfb8bd5695482ab788734e5908b` `source_revision_id=srcrev_ad7b7d413d9c1c8c1be9bcf326f97b6f` `chunk_id=srcchunk_8c5f7636ae3f8684db8b2030e4d9329b` `native_locator=slack:C0AL7EKNHDF:1775153012.510199:1775222717.712879` `source_timestamp=2026-04-03T13:25:17Z`
- ElevenLabs agents are required for each language, with their IDs defined in onboarding-profile.tsx. `claim:claim_2_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_653c9dfb8bd5695482ab788734e5908b` `source_revision_id=srcrev_ad7b7d413d9c1c8c1be9bcf326f97b6f` `chunk_id=srcchunk_8c5f7636ae3f8684db8b2030e4d9329b` `native_locator=slack:C0AL7EKNHDF:1775153012.510199:1775222717.712879` `source_timestamp=2026-04-03T13:25:17Z`

## Related Pages

- `localization-initiative`

## Sources

- `source_document_id`: `srcdoc_653c9dfb8bd5695482ab788734e5908b`
- `source_revision_id`: `srcrev_ad7b7d413d9c1c8c1be9bcf326f97b6f`
