---
title: "Story API confirmer batch failure (2026)"
type: "runbook"
slug: "runbooks/story-api-confirmer-batch-failure-2026"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "batch-failure"
  - "confirmer"
  - "incident"
  - "sentry"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_293bdfe46a2017d90950ab38274bb2d1"
  - "srcrev_304a40d134003f440fb613d9ccdaa919"
  - "srcrev_5b9ff39121bb80eb4266b3e63c140ebf"
  - "srcrev_be356a050436af3142413071c4fd6630"
conflict_state: "none"
---

# Story API confirmer batch failure (2026)

## Summary

Story API confirmer batch encountered multiple failures, but sibling batches continued. The Sentry issue STORY-API-EM was tracked and later resolved by Blake Huynh.

## Claims

- Story API confirmer batch failed; sibling batches continued processing. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_be356a050436af3142413071c4fd6630` `chunk_id=srcchunk_6f1df322f27be726db7f33664af74528` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1780809555.116029` `source_timestamp=2026-06-07T05:19:15Z`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_293bdfe46a2017d90950ab38274bb2d1` `chunk_id=srcchunk_aea5e8893fd56cb46b5ff5290b51b837` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1780975135.844959` `source_timestamp=2026-06-09T03:18:55Z`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_304a40d134003f440fb613d9ccdaa919` `chunk_id=srcchunk_9399eca50d425753cfcf35cf885062dd` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1781148095.122019` `source_timestamp=2026-06-11T03:21:35Z`
- The Sentry issue STORY-API-EM was resolved by Blake Huynh. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_5b9ff39121bb80eb4266b3e63c140ebf` `chunk_id=srcchunk_96a783ba73fb8faacf669c7ba2ee3f62` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1781630302.918789` `source_timestamp=2026-06-16T17:18:22Z`

## Sources

- `source_document_id`: `srcdoc_08c431942e702fb41d9eacec58132b28`
- `source_revision_id`: `srcrev_be356a050436af3142413071c4fd6630`
