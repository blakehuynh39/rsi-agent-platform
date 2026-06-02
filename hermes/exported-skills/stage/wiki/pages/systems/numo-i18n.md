---
title: "Numo Internationalization (i18n)"
type: "system"
slug: "systems/numo-i18n"
freshness: "2026-06-02T00:25:00Z"
tags:
  - "i18n"
  - "languages"
  - "localization"
owners:
  - "PIP Labs"
source_revision_ids:
  - "srcrev_f4f0737c3faeb96f40356202cf217726"
conflict_state: "none"
---

# Numo Internationalization (i18n)

## Summary

Supported languages and i18n tooling for the Numo web and React Native apps.

## Claims

- Live languages: English (en), Hindi (hi), Bengali (bn), Tamil (ta), Telugu (te). Gated behind flags: Vietnamese (vi, VITE_VIETNAMESE_ENABLED), Filipino (fil, VITE_FILIPINO_ENABLED). In progress: Korean (ko). `claim:claim_6_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`
- Web uses Lingui 5.x with PO files in src/locales/{locale}/; React Native uses i18next with JSON in i18n/locales/. `claim:claim_6_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`
- After copy changes, update translations: for web, pnpm lingui extract then pnpm lingui compile; for React Native, manually update i18n/locales/*.json. `claim:claim_6_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`

## Related Pages

- `numo-feature-flags`
- `numo-platform-overview`

## Sources

- `source_document_id`: `srcdoc_5b2f0266ee274de6a7ca742845dfbcd4`
- `source_revision_id`: `srcrev_f4f0737c3faeb96f40356202cf217726`
- `source_url`: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3)
