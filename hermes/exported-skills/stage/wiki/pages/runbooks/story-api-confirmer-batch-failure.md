---
title: "Story API Confirmer Batch Failure"
type: "runbook"
slug: "runbooks/story-api-confirmer-batch-failure"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "batch-failure"
  - "confirmer"
  - "sentry"
  - "story-api"
owners:
  - "story-api-team"
source_revision_ids:
  - "srcrev_5b9ff39121bb80eb4266b3e63c140ebf"
  - "srcrev_be356a050436af3142413071c4fd6630"
conflict_state: "none"
---

# Story API Confirmer Batch Failure

## Summary

Runbook for the recurring 'confirmer batch failed; continuing sibling batches' error in story-api, including its resolution by marking the Sentry issue as resolved.

## Claims

- The story-api confirmer batch failed, but the system continued processing sibling batches. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_be356a050436af3142413071c4fd6630` `chunk_id=srcchunk_6f1df322f27be726db7f33664af74528` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1780809555.116029` `source_timestamp=2026-06-07T05:19:15Z`
- Blake Huynh marked the Sentry issue STORY-API-EM as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_08c431942e702fb41d9eacec58132b28` `source_revision_id=srcrev_5b9ff39121bb80eb4266b3e63c140ebf` `chunk_id=srcchunk_96a783ba73fb8faacf669c7ba2ee3f62` `native_locator=slack:C07K3J4JTH6:1780809555.116029:1781630302.918789` `source_timestamp=2026-06-16T17:18:22Z`

## Sources

- `source_document_id`: `srcdoc_08c431942e702fb41d9eacec58132b28`
- `source_revision_id`: `srcrev_304a40d134003f440fb613d9ccdaa919`
