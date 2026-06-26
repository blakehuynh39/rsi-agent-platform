---
title: "504 Timeout on /api/v4/search POST"
type: "runbook"
slug: "runbooks/api-v4-search-504-timeout"
freshness: "2026-04-21T05:12:27Z"
tags:
  - "504"
  - "api"
  - "incident"
  - "search"
  - "timeout"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_447b7c3e695918a211a9bb7a75a64e84"
  - "srcrev_5a4487b6d91f512ff8465dfac59d0452"
  - "srcrev_6a7110e1a26d276d35710a9789a66307"
  - "srcrev_d0697e05177a06157feea6883d632d13"
  - "srcrev_e70742940ba1d15fdaee7d6523137993"
conflict_state: "none"
---

# 504 Timeout on /api/v4/search POST

## Summary

Multiple occurrences of POST /api/v4/search returning 504 Request Timeout, eventually resolved.

## Claims

- The story-api endpoint POST /api/v4/search returned 504 Request Timeout on multiple occasions. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3731dd88fe850cf663f5518d6a9b0918` `source_revision_id=srcrev_d0697e05177a06157feea6883d632d13` `chunk_id=srcchunk_deb503518574f1e378d849e403be0dea` `native_locator=slack:C07K3J4JTH6:1772243481.441719:1772243481.441719` `source_timestamp=2026-02-28T01:51:21Z`
  - citation: `source_document_id=srcdoc_3731dd88fe850cf663f5518d6a9b0918` `source_revision_id=srcrev_e70742940ba1d15fdaee7d6523137993` `chunk_id=srcchunk_b0265cdc9515beb6814064fc1dcc34aa` `native_locator=slack:C07K3J4JTH6:1772243481.441719:1776495177.243589` `source_timestamp=2026-04-18T06:52:57Z`
  - citation: `source_document_id=srcdoc_3731dd88fe850cf663f5518d6a9b0918` `source_revision_id=srcrev_6a7110e1a26d276d35710a9789a66307` `chunk_id=srcchunk_9664af9918f81dad4e15cc83048dca9e` `native_locator=slack:C07K3J4JTH6:1772243481.441719:1776616017.099169` `source_timestamp=2026-04-19T16:26:57Z`
  - citation: `source_document_id=srcdoc_3731dd88fe850cf663f5518d6a9b0918` `source_revision_id=srcrev_447b7c3e695918a211a9bb7a75a64e84` `chunk_id=srcchunk_4871540c43c7cb52d1c1a0d8c18b79e5` `native_locator=slack:C07K3J4JTH6:1772243481.441719:1776748347.932369` `source_timestamp=2026-04-21T05:12:27Z`
- The issue was resolved: blake.huynh marked the Sentry issue STORY-API-E9 as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3731dd88fe850cf663f5518d6a9b0918` `source_revision_id=srcrev_5a4487b6d91f512ff8465dfac59d0452` `chunk_id=srcchunk_a02a6ce2ba62e2d3c4b77c1f0c2d3737` `native_locator=slack:C07K3J4JTH6:1772243481.441719:1772294440.912719` `source_timestamp=2026-02-28T16:00:40Z`

## Open Questions

- What was the root cause of the timeouts?

## Sources

- `source_document_id`: `srcdoc_3731dd88fe850cf663f5518d6a9b0918`
- `source_revision_id`: `srcrev_5a4487b6d91f512ff8465dfac59d0452`
