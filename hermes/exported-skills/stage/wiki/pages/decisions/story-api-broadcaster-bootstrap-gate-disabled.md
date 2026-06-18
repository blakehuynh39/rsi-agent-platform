---
title: "Story API Broadcaster Bootstrap Gate Disabled"
type: "decision"
slug: "decisions/story-api-broadcaster-bootstrap-gate-disabled"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "bootstrap-gate"
  - "broadcaster"
  - "configuration"
  - "story-api"
owners: []
source_revision_ids:
  - "srcrev_86b0b3ece5d7a4c51e17d255f0162f14"
  - "srcrev_d15dbe59aec23ca513fd177bf1f13c96"
conflict_state: "none"
---

# Story API Broadcaster Bootstrap Gate Disabled

## Summary

The bootstrap gate for the story-api broadcaster was disabled by operator override, setting BROADCASTER_SKIP_BOOTSTRAP_GATE=true. Additionally, related Sentry issue STORY-API-EJ was resolved.

## Claims

- The bootstrap gate for story-api broadcaster is disabled via operator override, with environment variable BROADCASTER_SKIP_BOOTSTRAP_GATE set to true. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_d15dbe59aec23ca513fd177bf1f13c96` `chunk_id=srcchunk_58328e92fbdccb77076ab89e88204123` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1780764774.500999` `source_timestamp=2026-06-06T16:52:54Z`
- Sentry issue STORY-API-EJ (ID 7532674731) was marked as resolved by Blake Huynh (blake.huynh@storyprotocol.xyz). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_86b0b3ece5d7a4c51e17d255f0162f14` `chunk_id=srcchunk_c93d9f165879dd265df97d5346b06737` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1781630302.606789` `source_timestamp=2026-06-16T17:18:22Z`

## Sources

- `source_document_id`: `srcdoc_87f774c3367c1d9384de5a34db7cd1de`
- `source_revision_id`: `srcrev_86b0b3ece5d7a4c51e17d255f0162f14`
