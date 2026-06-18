---
title: "story-api GET /api/v1/data-audit/feed 500 Error"
type: "runbook"
slug: "runbooks/incident-story-api-data-audit-feed-500"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "500-error"
  - "data-audit"
  - "incident"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_04a00c2ee478c7ccb8d23280508d3b10"
  - "srcrev_ef1c8f9074c66352c93f31145f23ac09"
conflict_state: "none"
---

# story-api GET /api/v1/data-audit/feed 500 Error

## Summary

On 2026-06-12, the story-api endpoint GET /api/v1/data-audit/feed returned a 500 Internal Server Error with EOF. The issue was later resolved by Blake Huynh on 2026-06-14.

## Claims

- The story-api GET /api/v1/data-audit/feed endpoint failed with a 500 error and EOF. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_26427b55f7197d5929d9b3f58270b393` `source_revision_id=srcrev_04a00c2ee478c7ccb8d23280508d3b10` `chunk_id=srcchunk_fcf8148989b442d8c5fc014abbf6c14b` `native_locator=slack:C07K3J4JTH6:1781411577.416369:1781411577.416369` `source_timestamp=2026-06-14T04:32:57Z`
- Blake Huynh (blake.huynh@storyprotocol.xyz) marked the Sentry issue STORY-API-F0 as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_26427b55f7197d5929d9b3f58270b393` `source_revision_id=srcrev_ef1c8f9074c66352c93f31145f23ac09` `chunk_id=srcchunk_d9c6188f7b8b30bd4fc28941a2819253` `native_locator=slack:C07K3J4JTH6:1781411577.416369:1781630303.007639` `source_timestamp=2026-06-16T17:18:23Z`

## Open Questions

- What was the root cause of the 500 error?

## Sources

- `source_document_id`: `srcdoc_26427b55f7197d5929d9b3f58270b393`
- `source_revision_id`: `srcrev_ef1c8f9074c66352c93f31145f23ac09`
