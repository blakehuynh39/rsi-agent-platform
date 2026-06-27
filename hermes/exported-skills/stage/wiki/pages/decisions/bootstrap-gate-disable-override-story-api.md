---
title: "Bootstrap Gate Disable Override for story-api"
type: "decision"
slug: "decisions/bootstrap-gate-disable-override-story-api"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "bootstrap"
  - "gate"
  - "incident"
  - "override"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_86b0b3ece5d7a4c51e17d255f0162f14"
  - "srcrev_d15dbe59aec23ca513fd177bf1f13c96"
conflict_state: "none"
---

# Bootstrap Gate Disable Override for story-api

## Summary

The story-api bootstrap gate was disabled via BROADCASTER_SKIP_BOOTSTRAP_GATE=true override. Related Sentry issue STORY-API-EJ was resolved.

## Claims

- Bootstrap gate for story-api was disabled via BROADCASTER_SKIP_BOOTSTRAP_GATE=true operator override. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_d15dbe59aec23ca513fd177bf1f13c96` `chunk_id=srcchunk_58328e92fbdccb77076ab89e88204123` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1780764774.500999` `source_timestamp=2026-06-06T16:52:54Z`
- Issue STORY-API-EJ (Sentry #7532674731) was marked resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_86b0b3ece5d7a4c51e17d255f0162f14` `chunk_id=srcchunk_c93d9f165879dd265df97d5346b06737` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1781630302.606789` `source_timestamp=2026-06-16T17:18:22Z`

## Sources

- `source_document_id`: `srcdoc_87f774c3367c1d9384de5a34db7cd1de`
- `source_revision_id`: `srcrev_d91a4b17e05b1bc25a54e872c51466f7`
