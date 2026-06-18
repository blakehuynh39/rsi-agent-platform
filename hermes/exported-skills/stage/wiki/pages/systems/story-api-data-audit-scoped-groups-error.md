---
title: "Story-API Data Audit Scoped Groups 500 Error"
type: "system"
slug: "systems/story-api-data-audit-scoped-groups-error"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "500-error"
  - "access-denied"
  - "dynamodb"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_5acba376c088eba2a441d9a834ad435d"
  - "srcrev_a4bfab30f1563e4240774b8a1ff1ecdd"
conflict_state: "none"
---

# Story-API Data Audit Scoped Groups 500 Error

## Summary

Incident where story-api POST /api/v1/data-audit/scoped-groups returned 500 due to DynamoDB AccessDeniedException; resolved by blake.huynh.

## Claims

- story-api endpoint POST /api/v1/data-audit/scoped-groups failed with HTTP 500. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d150c996eab5dc4ff8e988a5c9ee13f5` `source_revision_id=srcrev_a4bfab30f1563e4240774b8a1ff1ecdd` `chunk_id=srcchunk_c7482d76b4333e73c0a2653af6d318cf` `native_locator=slack:C07K3J4JTH6:1780434697.439619:1780434697.439619` `source_timestamp=2026-06-02T21:11:37Z`
- The failure was caused by a DynamoDB AccessDeniedException. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d150c996eab5dc4ff8e988a5c9ee13f5` `source_revision_id=srcrev_a4bfab30f1563e4240774b8a1ff1ecdd` `chunk_id=srcchunk_c7482d76b4333e73c0a2653af6d318cf` `native_locator=slack:C07K3J4JTH6:1780434697.439619:1780434697.439619` `source_timestamp=2026-06-02T21:11:37Z`
- The Sentry issue STORY-API-EF was marked as resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d150c996eab5dc4ff8e988a5c9ee13f5` `source_revision_id=srcrev_5acba376c088eba2a441d9a834ad435d` `chunk_id=srcchunk_4654bd245b761884c1fa200a4b378d3e` `native_locator=slack:C07K3J4JTH6:1780434697.439619:1781630302.011249` `source_timestamp=2026-06-16T17:18:22Z`

## Sources

- `source_document_id`: `srcdoc_d150c996eab5dc4ff8e988a5c9ee13f5`
- `source_revision_id`: `srcrev_5acba376c088eba2a441d9a834ad435d`
