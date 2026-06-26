---
title: "Incident: Story API data-audit scoped groups AccessDenied"
type: "runbook"
slug: "runbooks/incident-story-api-data-audit-scoped-groups-access-denied"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "access-denied"
  - "dynamodb"
  - "incident"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_5acba376c088eba2a441d9a834ad435d"
  - "srcrev_a4bfab30f1563e4240774b8a1ff1ecdd"
conflict_state: "none"
---

# Incident: Story API data-audit scoped groups AccessDenied

## Summary

The Story API endpoint /api/v1/data-audit/scoped-groups returned a 500 error due to an AccessDeniedException when calling DynamoDB PutItem. The issue was resolved by Blake Huynh in Sentry (STORY-API-EF).

## Claims

- POST /api/v1/data-audit/scoped-groups failed with a 500 error due to DynamoDB AccessDeniedException. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d150c996eab5dc4ff8e988a5c9ee13f5` `source_revision_id=srcrev_a4bfab30f1563e4240774b8a1ff1ecdd` `chunk_id=srcchunk_c7482d76b4333e73c0a2653af6d318cf` `native_locator=slack:C07K3J4JTH6:1780434697.439619:1780434697.439619` `source_timestamp=2026-06-02T21:11:37Z`
- Blake Huynh resolved the related Sentry issue STORY-API-EF. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d150c996eab5dc4ff8e988a5c9ee13f5` `source_revision_id=srcrev_5acba376c088eba2a441d9a834ad435d` `chunk_id=srcchunk_4654bd245b761884c1fa200a4b378d3e` `native_locator=slack:C07K3J4JTH6:1780434697.439619:1781630302.011249` `source_timestamp=2026-06-16T17:18:22Z`

## Open Questions

- What specific IAM permission was missing for the DynamoDB PutItem operation?

## Sources

- `source_document_id`: `srcdoc_d150c996eab5dc4ff8e988a5c9ee13f5`
- `source_revision_id`: `srcrev_a4bfab30f1563e4240774b8a1ff1ecdd`
