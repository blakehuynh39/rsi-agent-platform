---
title: "story-api data-audit stats 500 error incident"
type: "system"
slug: "systems/story-api-data-audit-stats-500-incident"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "api"
  - "error"
  - "incident"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_08be8919fd11907da50d4e03d20bb2b8"
  - "srcrev_865b766fac2b34182497d5ff9e79be2f"
conflict_state: "none"
---

# story-api data-audit stats 500 error incident

## Summary

On 2026-06-14, the story-api endpoint GET /api/v1/data-audit/stats returned a 500 Internal Server Error. The issue was later resolved as indicated by Blake Huynh marking it resolved in Sentry.

## Claims

- The endpoint GET /api/v1/data-audit/stats returned a 500 Internal Server Error. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9b2fd6209b5b525f2e11804f3e32ab5c` `source_revision_id=srcrev_08be8919fd11907da50d4e03d20bb2b8` `chunk_id=srcchunk_003468752cfab35140c9df931ad3e8a7` `native_locator=slack:C07K3J4JTH6:1781411570.515519:1781411570.515519` `source_timestamp=2026-06-14T04:32:50Z`
- The issue was resolved by Blake Huynh. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9b2fd6209b5b525f2e11804f3e32ab5c` `source_revision_id=srcrev_865b766fac2b34182497d5ff9e79be2f` `chunk_id=srcchunk_e35b564ea3eae5b464e16371968e9140` `native_locator=slack:C07K3J4JTH6:1781411570.515519:1781630303.056629` `source_timestamp=2026-06-16T17:18:23Z`

## Open Questions

- What was the root cause of the 500 error?

## Sources

- `source_document_id`: `srcdoc_9b2fd6209b5b525f2e11804f3e32ab5c`
- `source_revision_id`: `srcrev_08be8919fd11907da50d4e03d20bb2b8`
