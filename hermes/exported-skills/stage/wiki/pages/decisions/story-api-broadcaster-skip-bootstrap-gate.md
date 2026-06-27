---
title: "story-api: Broadcaster Bootstrap Gate Override"
type: "decision"
slug: "decisions/story-api-broadcaster-skip-bootstrap-gate"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "broadcaster"
  - "configuration"
  - "feature-flag"
  - "story-api"
owners:
  - "story-api team"
source_revision_ids:
  - "srcrev_86b0b3ece5d7a4c51e17d255f0162f14"
  - "srcrev_d91a4b17e05b1bc25a54e872c51466f7"
conflict_state: "none"
---

# story-api: Broadcaster Bootstrap Gate Override

## Summary

The story-api broadcaster bootstrap gate is currently disabled by an operator override via the environment variable BROADCASTER_SKIP_BOOTSTRAP_GATE=true. The related Sentry issue STORY-API-EJ has been resolved.

## Claims

- The environment variable BROADCASTER_SKIP_BOOTSTRAP_GATE is set to true, disabling the bootstrap gate by operator override. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_d91a4b17e05b1bc25a54e872c51466f7` `chunk_id=srcchunk_922a967e8c1c412ed8b47de90b341874` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1780958677.485309` `source_timestamp=2026-06-08T22:44:37Z`
- The Sentry issue STORY-API-EJ was resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_87f774c3367c1d9384de5a34db7cd1de` `source_revision_id=srcrev_86b0b3ece5d7a4c51e17d255f0162f14` `chunk_id=srcchunk_c93d9f165879dd265df97d5346b06737` `native_locator=slack:C07K3J4JTH6:1780764774.500999:1781630302.606789` `source_timestamp=2026-06-16T17:18:22Z`

## Open Questions

- Is the bootstrap gate override temporary or permanent?
- Is the override directly related to the resolution of Sentry issue STORY-API-EJ?

## Sources

- `source_document_id`: `srcdoc_87f774c3367c1d9384de5a34db7cd1de`
- `source_revision_id`: `srcrev_a3ddb4fde10a5048cfad5844d1030b79`
