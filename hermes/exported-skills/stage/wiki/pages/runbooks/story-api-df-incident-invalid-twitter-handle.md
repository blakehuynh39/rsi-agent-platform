---
title: "STORY-API-DF Incident: Invalid twitter_handle field"
type: "runbook"
slug: "runbooks/story-api-df-incident-invalid-twitter-handle"
freshness: "2025-12-22T17:12:26Z"
tags:
  - "incident"
  - "story-api"
  - "twitter_handle"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_4cd8b1c9ae4185b05aef2accc1d6147c"
  - "srcrev_a76a7e960b392d8f7b7a1f109aa34d62"
  - "srcrev_ada0a1cb537732e9c5fc742e0fdcc545"
conflict_state: "none"
---

# STORY-API-DF Incident: Invalid twitter_handle field

## Summary

On 2025-12-22, the story-api encountered an 'errors.fundamental: invalid field: twitter_handle' error. The issue was identified and marked as resolved by Blake Huynh.

## Claims

- The story-api error 'errors.fundamental: invalid field: twitter_handle' was logged. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_53ddcef37c64dd272a5ef8257429dbac` `source_revision_id=srcrev_ada0a1cb537732e9c5fc742e0fdcc545` `chunk_id=srcchunk_f84d789d30704db731d37833beb96253` `native_locator=slack:C07K3J4JTH6:1766393630.121209:1766393630.121209` `source_timestamp=2025-12-22T08:53:50Z`
- Blake Huynh marked the issue STORY-API-DF as resolved on 2025-12-22 17:22:16 UTC (initial notification). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_53ddcef37c64dd272a5ef8257429dbac` `source_revision_id=srcrev_4cd8b1c9ae4185b05aef2accc1d6147c` `chunk_id=srcchunk_783c4af6a2ec48c74179f757f8781031` `native_locator=slack:C07K3J4JTH6:1766393630.121209:1766422536.483539` `source_timestamp=2025-12-22T16:55:36Z`
- Blake Huynh marked the issue STORY-API-DF as resolved again at 2025-12-22 17:25:46 UTC (second notification). `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_53ddcef37c64dd272a5ef8257429dbac` `source_revision_id=srcrev_a76a7e960b392d8f7b7a1f109aa34d62` `chunk_id=srcchunk_2d63a1001297908431bba6522c6f85f4` `native_locator=slack:C07K3J4JTH6:1766393630.121209:1766423546.351019` `source_timestamp=2025-12-22T17:12:26Z`

## Open Questions

- What is the root cause of the invalid twitter_handle field error?
- Why was the issue marked as resolved twice?

## Sources

- `source_document_id`: `srcdoc_53ddcef37c64dd272a5ef8257429dbac`
- `source_revision_id`: `srcrev_a76a7e960b392d8f7b7a1f109aa34d62`
