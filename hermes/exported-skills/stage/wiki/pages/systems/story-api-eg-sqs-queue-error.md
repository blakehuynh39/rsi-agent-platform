---
title: "STORY-API-EG: SQS Queue Non-Existent Error"
type: "system"
slug: "systems/story-api-eg-sqs-queue-error"
freshness: "2026-06-16T17:18:21Z"
tags:
  - "incident"
  - "sqs"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_849c28922da7ca2ef32a7e29ace2c9a6"
  - "srcrev_b39171e007052cf05bf94f8e243f9316"
conflict_state: "none"
---

# STORY-API-EG: SQS Queue Non-Existent Error

## Summary

The story-api encountered an HTTP 503 error when calling POST /api/v1/data-audit/scoped-groups due to a non-existent SQS queue. The issue was tracked as STORY-API-EG and resolved by Blake Huynh.

## Claims

- POST /api/v1/data-audit/scoped-groups failed with HTTP 503 due to SQS error: queue does not exist. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_966726b344f437a2abfaf81ab6fb4e74` `source_revision_id=srcrev_849c28922da7ca2ef32a7e29ace2c9a6` `chunk_id=srcchunk_4f8b392add695a5129056828be8923fa` `native_locator=slack:C07K3J4JTH6:1780437530.620189:1780437530.620189` `source_timestamp=2026-06-02T21:58:50Z`
- Issue STORY-API-EG was marked as resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_966726b344f437a2abfaf81ab6fb4e74` `source_revision_id=srcrev_b39171e007052cf05bf94f8e243f9316` `chunk_id=srcchunk_8951dba139d16961d39d9bdb49ce5938` `native_locator=slack:C07K3J4JTH6:1780437530.620189:1781630301.981489` `source_timestamp=2026-06-16T17:18:21Z`

## Open Questions

- Why was the SQS queue missing?

## Sources

- `source_document_id`: `srcdoc_966726b344f437a2abfaf81ab6fb4e74`
- `source_revision_id`: `srcrev_849c28922da7ca2ef32a7e29ace2c9a6`
