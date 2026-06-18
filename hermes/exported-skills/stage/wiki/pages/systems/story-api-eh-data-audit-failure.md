---
title: "STORY-API-EH: Data Audit Webhook Failure"
type: "system"
slug: "systems/story-api-eh-data-audit-failure"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "access-denied"
  - "dynamodb"
  - "error"
  - "resolved"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_ad08c1a2b5182e5bfd09b6b0069b9a80"
  - "srcrev_e67384184cbd8b20beb7395b0909fa5a"
conflict_state: "none"
---

# STORY-API-EH: Data Audit Webhook Failure

## Summary

The story-api endpoint POST /webhook/v1/data-audit/records failed with a 500 error due to an AccessDeniedException in DynamoDB. Blake Huynh resolved the issue as STORY-API-EH.

## Claims

- The story-api endpoint POST /webhook/v1/data-audit/records returned a 500 error. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_756e4b27c63137d5551add1d152a6567` `source_revision_id=srcrev_e67384184cbd8b20beb7395b0909fa5a` `chunk_id=srcchunk_bdb474336191342907ad9b8320e1d581` `native_locator=slack:C07K3J4JTH6:1780534946.310219:1780534946.310219` `source_timestamp=2026-06-04T01:02:26Z`
- The error involved a DynamoDB BatchGetItem returning a 400 with an AccessDeniedException. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_756e4b27c63137d5551add1d152a6567` `source_revision_id=srcrev_e67384184cbd8b20beb7395b0909fa5a` `chunk_id=srcchunk_bdb474336191342907ad9b8320e1d581` `native_locator=slack:C07K3J4JTH6:1780534946.310219:1780534946.310219` `source_timestamp=2026-06-04T01:02:26Z`
- Blake Huynh marked the Sentry issue STORY-API-EH as resolved. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_756e4b27c63137d5551add1d152a6567` `source_revision_id=srcrev_ad08c1a2b5182e5bfd09b6b0069b9a80` `chunk_id=srcchunk_e20fab7483c4d0e3abea9333bcfeae63` `native_locator=slack:C07K3J4JTH6:1780534946.310219:1781630302.848689` `source_timestamp=2026-06-16T17:18:22Z`

## Sources

- `source_document_id`: `srcdoc_756e4b27c63137d5551add1d152a6567`
- `source_revision_id`: `srcrev_ad08c1a2b5182e5bfd09b6b0069b9a80`
