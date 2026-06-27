---
title: "Data Audit Feed 500 Incident"
type: "runbook"
slug: "runbooks/data-audit-feed-500-incident"
freshness: "2026-06-16T17:18:23Z"
tags:
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

# Data Audit Feed 500 Incident

## Summary

The story-api endpoint GET /api/v1/data-audit/feed failed with a 500 Internal Server Error, reading EOF. The incident was tracked in Sentry as STORY-API-F0 and later resolved by Blake Huynh.

## Claims

- GET /api/v1/data-audit/feed returned 500 with EOF `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_26427b55f7197d5929d9b3f58270b393` `source_revision_id=srcrev_04a00c2ee478c7ccb8d23280508d3b10` `chunk_id=srcchunk_fcf8148989b442d8c5fc014abbf6c14b` `native_locator=slack:C07K3J4JTH6:1781411577.416369:1781411577.416369` `source_timestamp=2026-06-14T04:32:57Z`
- Incident tracked as Sentry issue STORY-API-F0 `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_26427b55f7197d5929d9b3f58270b393` `source_revision_id=srcrev_ef1c8f9074c66352c93f31145f23ac09` `chunk_id=srcchunk_d9c6188f7b8b30bd4fc28941a2819253` `native_locator=slack:C07K3J4JTH6:1781411577.416369:1781630303.007639` `source_timestamp=2026-06-16T17:18:23Z`
- Resolved by Blake Huynh (blake.huynh@storyprotocol.xyz) `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_26427b55f7197d5929d9b3f58270b393` `source_revision_id=srcrev_ef1c8f9074c66352c93f31145f23ac09` `chunk_id=srcchunk_d9c6188f7b8b30bd4fc28941a2819253` `native_locator=slack:C07K3J4JTH6:1781411577.416369:1781630303.007639` `source_timestamp=2026-06-16T17:18:23Z`

## Sources

- `source_document_id`: `srcdoc_26427b55f7197d5929d9b3f58270b393`
- `source_revision_id`: `srcrev_04a00c2ee478c7ccb8d23280508d3b10`
