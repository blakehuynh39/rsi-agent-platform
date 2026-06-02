---
title: "Numo Feature Flags"
type: "system"
slug: "systems/numo-feature-flags"
freshness: "2026-06-02T00:25:00Z"
tags:
  - "configuration"
  - "feature-flags"
owners:
  - "PIP Labs"
source_revision_ids:
  - "srcrev_f4f0737c3faeb96f40356202cf217726"
conflict_state: "none"
---

# Numo Feature Flags

## Summary

Current feature flags used in the Numo platform, gating features like professional multimodal, voice phrase, and locale support.

## Claims

- VITE_VOICE_PHRASE_ENABLED (default false) gates per-campaign seed-phrase verification task. `claim:claim_5_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`
- VITE_FILIPINO_ENABLED (default false) gates Filipino UI locale; VITE_VIETNAMESE_ENABLED (default false) gates Vietnamese UI locale. `claim:claim_5_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`
- VITE_PROFESSIONAL_MULTIMODAL_ENABLED (default false) gates Pro tasks, resume upload, LinkedIn connect, file upload campaigns. `claim:claim_5_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`
- VITE_DEV_MOCK_API (default false in prod) bypasses backend with mock data; VITE_GA_DEBUG_MODE (default false) enables GA4 DebugView. `claim:claim_5_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`
- Backend flags: VOICE_PHRASE_ENABLED (default false), BEEHIIV_ENABLED (default false) for email signup hook, LINKEDIN_ENABLED (default false) for LinkedIn OIDC connect. `claim:claim_5_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`

## Related Pages

- `numo-platform-overview`

## Sources

- `source_document_id`: `srcdoc_5b2f0266ee274de6a7ca742845dfbcd4`
- `source_revision_id`: `srcrev_f4f0737c3faeb96f40356202cf217726`
- `source_url`: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3)
