---
title: "Story-API Confirmer Batch Failure"
type: "runbook"
slug: "runbooks/story-api-confirmer-batch-failure"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "confirmer"
  - "incident"
  - "story-api"
owners: []
source_revision_ids:
  - "srcrev_293bdfe46a2017d90950ab38274bb2d1"
  - "srcrev_304a40d134003f440fb613d9ccdaa919"
  - "srcrev_5b9ff39121bb80eb4266b3e63c140ebf"
  - "srcrev_be356a050436af3142413071c4fd6630"
conflict_state: "none"
---

# Story-API Confirmer Batch Failure

## Summary

The story-api confirmer batch repeatedly failed, but sibling batches continued. The issue STORY-API-EM was resolved by Blake Huynh.

## Claims

- The story-api confirmer batch failed on 2026-05-31. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_be356a050436af3142413071c4fd6630` `chunk_id=srcchunk_6f1df322f27be726db7f33664af74528` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1780809555.116029` `source_timestamp=2026-06-07T05:19:15Z`
- Despite the confirmer batch failure, sibling batches continued processing, indicating resilient design. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_be356a050436af3142413071c4fd6630` `chunk_id=srcchunk_6f1df322f27be726db7f33664af74528` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1780809555.116029` `source_timestamp=2026-06-07T05:19:15Z`
- The confirmer batch failure recurred on 2026-06-02 and 2026-06-04. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_293bdfe46a2017d90950ab38274bb2d1` `chunk_id=srcchunk_aea5e8893fd56cb46b5ff5290b51b837` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1780975135.844959` `source_timestamp=2026-06-09T03:18:55Z`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_304a40d134003f440fb613d9ccdaa919` `chunk_id=srcchunk_9399eca50d425753cfcf35cf885062dd` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1781148095.122019` `source_timestamp=2026-06-11T03:21:35Z`
- The issue was tracked as STORY-API-EM in Sentry. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_5b9ff39121bb80eb4266b3e63c140ebf` `chunk_id=srcchunk_96a783ba73fb8faacf669c7ba2ee3f62` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1781630302.918789` `source_timestamp=2026-06-16T17:18:22Z`
- Blake Huynh (blake.huynh@storyprotocol.xyz) resolved the STORY-API-EM issue on 2026-06-14. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_5b9ff39121bb80eb4266b3e63c140ebf` `chunk_id=srcchunk_96a783ba73fb8faacf669c7ba2ee3f62` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1781630302.918789` `source_timestamp=2026-06-16T17:18:22Z`

## Open Questions

- What was the root cause of the confirmer batch failures?

## Sources

- `source_document_id`: `srcdoc_08c431942e702fb41d9eacec58132b28`
- `source_revision_id`: `srcrev_5b9ff39121bb80eb4266b3e63c140ebf`
