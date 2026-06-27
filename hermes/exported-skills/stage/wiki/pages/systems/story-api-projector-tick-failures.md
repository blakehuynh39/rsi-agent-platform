---
title: "Story API Projector Tick Failures"
type: "system"
slug: "systems/story-api-projector-tick-failures"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "projector"
  - "sentry"
  - "story-api"
  - "tick-failure"
owners: []
source_revision_ids:
  - "srcrev_620f99be090c61ae4218abba4dcf9a2e"
  - "srcrev_63d68175ac896fc21428dd9fdaa27daa"
  - "srcrev_68834830cb0157b12ba0af791aeb60b2"
  - "srcrev_ecd4dffe3efb15c4fb0a1a423b6aee9e"
conflict_state: "none"
---

# Story API Projector Tick Failures

## Summary

Recurring failures of the story-api projector tick, tracked in Sentry as STORY-API-ER. The issue has been resolved multiple times but reappears.

## Claims

- The story-api projector experienced tick failures. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99cf0cc72d0931b2a10e96ac57afbf12` `source_revision_id=srcrev_63d68175ac896fc21428dd9fdaa27daa` `chunk_id=srcchunk_c77e2fd7018ab41cd8ce75c505760288` `native_locator=slack:C07K3J4JTH6:1780882511.566079:1780882511.566079` `source_timestamp=2026-06-08T01:35:11Z`
- The error message observed was 'projector tick failed; sleeping before retry'. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99cf0cc72d0931b2a10e96ac57afbf12` `source_revision_id=srcrev_63d68175ac896fc21428dd9fdaa27daa` `chunk_id=srcchunk_c77e2fd7018ab41cd8ce75c505760288` `native_locator=slack:C07K3J4JTH6:1780882511.566079:1780882511.566079` `source_timestamp=2026-06-08T01:35:11Z`
- Romain Magne resolved the STORY-API-ER Sentry issue. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99cf0cc72d0931b2a10e96ac57afbf12` `source_revision_id=srcrev_620f99be090c61ae4218abba4dcf9a2e` `chunk_id=srcchunk_422a0f38339aa4a61465cf1661ad9ba1` `native_locator=slack:C07K3J4JTH6:1780882511.566079:1780942767.257199` `source_timestamp=2026-06-08T18:19:27Z`
- The projector tick failure recurred after resolution. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99cf0cc72d0931b2a10e96ac57afbf12` `source_revision_id=srcrev_68834830cb0157b12ba0af791aeb60b2` `chunk_id=srcchunk_f6712a6c7d7501f95618c41174d6b333` `native_locator=slack:C07K3J4JTH6:1780882511.566079:1781049681.856229` `source_timestamp=2026-06-10T00:01:21Z`
- Blake Huynh later resolved the STORY-API-ER Sentry issue. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_99cf0cc72d0931b2a10e96ac57afbf12` `source_revision_id=srcrev_ecd4dffe3efb15c4fb0a1a423b6aee9e` `chunk_id=srcchunk_60bc797f358333f2f27289d5f9647358` `native_locator=slack:C07K3J4JTH6:1780882511.566079:1781630303.107649` `source_timestamp=2026-06-16T17:18:23Z`

## Open Questions

- Is there a permanent fix?
- What is the root cause?
- Why does the projector tick keep failing?

## Sources

- `source_document_id`: `srcdoc_99cf0cc72d0931b2a10e96ac57afbf12`
- `source_revision_id`: `srcrev_8d728b65b7dc5fea358dd7f5c4a6efde`
