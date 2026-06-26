---
title: "Story-API DynamoDB AccessDeniedException on POST /webhook/v1/data-audit/records"
type: "runbook"
slug: "runbooks/story-api-dynamodb-access-denied-incident"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "AccessDeniedException"
  - "DynamoDB"
  - "incident"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_ad08c1a2b5182e5bfd09b6b0069b9a80"
  - "srcrev_e67384184cbd8b20beb7395b0909fa5a"
conflict_state: "none"
---

# Story-API DynamoDB AccessDeniedException on POST /webhook/v1/data-audit/records

## Summary

The Story-API experienced a 500 error on endpoint POST /webhook/v1/data-audit/records due to an AccessDeniedException when calling DynamoDB BatchGetItem. The issue was resolved by Blake Huynh.

## Claims

- The story-api POST /webhook/v1/data-audit/records operation failed with a 500 error, encountering a DynamoDB AccessDeniedException during a BatchGetItem call with status 400. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_756e4b27c63137d5551add1d152a6567` `source_revision_id=srcrev_e67384184cbd8b20beb7395b0909fa5a` `chunk_id=srcchunk_bdb474336191342907ad9b8320e1d581` `native_locator=slack:C07K3J4JTH6:1780534946.310219:1780534946.310219` `source_timestamp=2026-06-04T01:02:26Z`
- Blake Huynh marked the issue as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_756e4b27c63137d5551add1d152a6567` `source_revision_id=srcrev_ad08c1a2b5182e5bfd09b6b0069b9a80` `chunk_id=srcchunk_e20fab7483c4d0e3abea9333bcfeae63` `native_locator=slack:C07K3J4JTH6:1780534946.310219:1781630302.848689` `source_timestamp=2026-06-16T17:18:22Z`

## Open Questions

- What was the root cause of the AccessDeniedException for the DynamoDB BatchGetItem call?

## Sources

- `source_document_id`: `srcdoc_756e4b27c63137d5551add1d152a6567`
- `source_revision_id`: `srcrev_e67384184cbd8b20beb7395b0909fa5a`
