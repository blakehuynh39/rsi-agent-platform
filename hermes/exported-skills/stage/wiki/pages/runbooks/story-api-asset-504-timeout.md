---
title: "Story API Asset 504 Timeout Incident"
type: "runbook"
slug: "runbooks/story-api-asset-504-timeout"
freshness: "2026-02-28T16:00:43Z"
tags:
  - "504"
  - "incident"
  - "resolved"
  - "story-api"
owners:
  - "Blake Huynh"
source_revision_ids:
  - "srcrev_03dd760aa8ed2f8c6c910bb68817034c"
  - "srcrev_6e213ba260f0a098ee1ce26d814b9831"
conflict_state: "none"
---

# Story API Asset 504 Timeout Incident

## Summary

On 2026-02-28, the story-api endpoint POST /api/v4/assets returned a 504 Request Timeout. Blake Huynh resolved the issue, marking Sentry issue STORY-API-E2 as resolved.

## Claims

- story-api POST /api/v4/assets failed with 504: Request timeout `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00a83b33937e05dccccc889406623a2a` `source_revision_id=srcrev_6e213ba260f0a098ee1ce26d814b9831` `chunk_id=srcchunk_d5990cdab8d2fa7abc63b9fedf70ba29` `native_locator=slack:C07K3J4JTH6:1772236577.127539:1772236577.127539` `source_timestamp=2026-02-27T23:56:17Z`
- Blake Huynh marked the Sentry issue STORY-API-E2 as resolved `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00a83b33937e05dccccc889406623a2a` `source_revision_id=srcrev_03dd760aa8ed2f8c6c910bb68817034c` `chunk_id=srcchunk_c7dfe431f265a1692183d138e32cdfaa` `native_locator=slack:C07K3J4JTH6:1772236577.127539:1772294443.310999` `source_timestamp=2026-02-28T16:00:43Z`

## Open Questions

- What was the root cause of the 504 timeout?

## Sources

- `source_document_id`: `srcdoc_00a83b33937e05dccccc889406623a2a`
- `source_revision_id`: `srcrev_03dd760aa8ed2f8c6c910bb68817034c`
