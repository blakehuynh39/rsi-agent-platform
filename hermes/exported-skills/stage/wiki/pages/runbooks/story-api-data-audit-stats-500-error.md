---
title: "Story API /data-audit/stats 500 Error Runbook"
type: "runbook"
slug: "runbooks/story-api-data-audit-stats-500-error"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "500-error"
  - "data-audit"
  - "incident"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_08be8919fd11907da50d4e03d20bb2b8"
  - "srcrev_865b766fac2b34182497d5ff9e79be2f"
conflict_state: "none"
---

# Story API /data-audit/stats 500 Error Runbook

## Summary

On ~2026-06-16, the endpoint GET /api/v1/data-audit/stats failed with HTTP 500 and EOF error. The issue was raised in Slack and later resolved by Blake Huynh via Sentry issue STORY-API-EZ.

## Claims

- GET /api/v1/data-audit/stats failed with 500: EOF `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9b2fd6209b5b525f2e11804f3e32ab5c` `source_revision_id=srcrev_08be8919fd11907da50d4e03d20bb2b8` `chunk_id=srcchunk_003468752cfab35140c9df931ad3e8a7` `native_locator=slack:C07K3J4JTH6:1781411570.515519:1781411570.515519` `source_timestamp=2026-06-14T04:32:50Z`
- Blake Huynh resolved the issue via Sentry (STORY-API-EZ) `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9b2fd6209b5b525f2e11804f3e32ab5c` `source_revision_id=srcrev_865b766fac2b34182497d5ff9e79be2f` `chunk_id=srcchunk_e35b564ea3eae5b464e16371968e9140` `native_locator=slack:C07K3J4JTH6:1781411570.515519:1781630303.056629` `source_timestamp=2026-06-16T17:18:23Z`

## Open Questions

- What caused the 500 EOF error on /api/v1/data-audit/stats?

## Sources

- `source_document_id`: `srcdoc_9b2fd6209b5b525f2e11804f3e32ab5c`
- `source_revision_id`: `srcrev_865b766fac2b34182497d5ff9e79be2f`
