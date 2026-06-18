---
title: "story-api projection conflict: refusing to overwrite existing tx_hash"
type: "runbook"
slug: "runbooks/story-api-projection-conflict-txhash-overwrite"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "conflict"
  - "projection"
  - "story-api"
  - "tx_hash"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_0313ef5ff33949ce5eb44bcb9d1ee553"
  - "srcrev_d8a4da57e2237e3f1f3cdde06d4b8601"
conflict_state: "none"
---

# story-api projection conflict: refusing to overwrite existing tx_hash

## Summary

A projection conflict in story-api prevented overwriting an existing tx_hash. The issue was resolved by Blake Huynh.

## Claims

- A projection conflict occurred in story-api, refusing to overwrite an existing tx_hash. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_84c7b5e454a2fd8b36914e9ee394cfca` `source_revision_id=srcrev_0313ef5ff33949ce5eb44bcb9d1ee553` `chunk_id=srcchunk_ce148bfd5d607bc5b206781188a638d3` `native_locator=slack:C07K3J4JTH6:1780967321.307329:1780967321.307329` `source_timestamp=2026-06-09T01:08:41Z`
- The story-api Sentry issue STORY-API-EV was marked resolved by Blake Huynh. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_84c7b5e454a2fd8b36914e9ee394cfca` `source_revision_id=srcrev_d8a4da57e2237e3f1f3cdde06d4b8601` `chunk_id=srcchunk_8cc4b2cc728357cd070380bfb37750f4` `native_locator=slack:C07K3J4JTH6:1780967321.307329:1781630303.117919` `source_timestamp=2026-06-16T17:18:23Z`

## Sources

- `source_document_id`: `srcdoc_84c7b5e454a2fd8b36914e9ee394cfca`
- `source_revision_id`: `srcrev_d8a4da57e2237e3f1f3cdde06d4b8601`
