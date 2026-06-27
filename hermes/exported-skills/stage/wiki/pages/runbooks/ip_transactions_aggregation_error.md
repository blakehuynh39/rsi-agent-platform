---
title: "ip_transactions Aggregation Error"
type: "runbook"
slug: "runbooks/ip_transactions_aggregation_error"
freshness: "2026-02-28T16:01:00Z"
tags:
  - "aggregation"
  - "error"
  - "incident"
  - "ip_transactions"
  - "story-orchestration-service"
owners:
  - "blake.huynh"
source_revision_ids:
  - "srcrev_3442962645f6afbcefca5420092da972"
  - "srcrev_345f0dbc7042faa16ea8c0f4a3f8d1bd"
conflict_state: "none"
---

# ip_transactions Aggregation Error

## Summary

The story-orchestration-service reported an aggregation error for ip_transactions. The issue was later resolved by Blake Huynh via Sentry issue STORY-ORCHESTRATION-SERVICE-FQ.

## Claims

- The story-orchestration-service encountered an aggregation error with ip_transactions. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2cd8712e780e6d2fd03fe1b46e7ab26f` `source_revision_id=srcrev_345f0dbc7042faa16ea8c0f4a3f8d1bd` `chunk_id=srcchunk_2cf5bcb14edd8ecc9233d8eae11bbdbc` `native_locator=slack:C08BWTULNPP:1772239373.481999` `source_timestamp=2026-02-28T00:42:53Z`
- blake.huynh@storyprotocol.xyz resolved the related Sentry issue STORY-ORCHESTRATION-SERVICE-FQ. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2cd8712e780e6d2fd03fe1b46e7ab26f` `source_revision_id=srcrev_3442962645f6afbcefca5420092da972` `chunk_id=srcchunk_1e5b8c494859d2b933c22fb894d7cf9f` `native_locator=slack:C08BWTULNPP:1772294460.622699` `source_timestamp=2026-02-28T16:01:00Z`

## Open Questions

- Root cause of ip_transactions aggregation error?

## Related Pages

- `story-orchestration-service`

## Sources

- `source_document_id`: `srcdoc_2cd8712e780e6d2fd03fe1b46e7ab26f`
- `source_revision_id`: `srcrev_345f0dbc7042faa16ea8c0f4a3f8d1bd`
