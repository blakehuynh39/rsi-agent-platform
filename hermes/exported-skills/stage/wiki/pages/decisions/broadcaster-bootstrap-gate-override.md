---
title: "Broadcaster Bootstrap Gate Override"
type: "decision"
slug: "decisions/broadcaster-bootstrap-gate-override"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "bootstrap-gate"
  - "broadcaster"
  - "operator-override"
  - "story-api"
owners: []
source_revision_ids:
  - "srcrev_86b0b3ece5d7a4c51e17d255f0162f14"
  - "srcrev_a3ddb4fde10a5048cfad5844d1030b79"
  - "srcrev_d15dbe59aec23ca513fd177bf1f13c96"
  - "srcrev_d91a4b17e05b1bc25a54e872c51466f7"
conflict_state: "none"
---

# Broadcaster Bootstrap Gate Override

## Summary

The broadcaster bootstrap gate in story-api is disabled via the environment variable BROADCASTER_SKIP_BOOTSTRAP_GATE=true, as an operator override. Associated Sentry issue STORY-API-EJ was resolved by Blake Huynh.

## Claims

- The story-api broadcaster bootstrap gate is disabled via operator override using environment variable BROADCASTER_SKIP_BOOTSTRAP_GATE=true. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_d15dbe59aec23ca513fd177bf1f13c96` `chunk_id=srcchunk_58328e92fbdccb77076ab89e88204123` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1780764774.500999` `source_timestamp=2026-06-06T16:52:54Z`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_a3ddb4fde10a5048cfad5844d1030b79` `chunk_id=srcchunk_bd8451c121cca77cc40ecf0b53b932c1` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1780858682.816089` `source_timestamp=2026-06-07T18:58:02Z`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_d91a4b17e05b1bc25a54e872c51466f7` `chunk_id=srcchunk_922a967e8c1c412ed8b47de90b341874` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1780958677.485309` `source_timestamp=2026-06-08T22:44:37Z`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_86b0b3ece5d7a4c51e17d255f0162f14` `chunk_id=srcchunk_c93d9f165879dd265df97d5346b06737` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1781630302.606789` `source_timestamp=2026-06-16T17:18:22Z`
- blake.huynh@storyprotocol.xyz marked Sentry issue STORY-API-EJ as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_86b0b3ece5d7a4c51e17d255f0162f14` `chunk_id=srcchunk_c93d9f165879dd265df97d5346b06737` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1781630302.606789` `source_timestamp=2026-06-16T17:18:22Z`

## Open Questions

- Why was the bootstrap gate disabled? Is this a permanent override?

## Sources

- `source_document_id`: `srcdoc_87f774c3367c1d9384de5a34db7cd1de`
- `source_revision_id`: `srcrev_d15dbe59aec23ca513fd177bf1f13c96`
