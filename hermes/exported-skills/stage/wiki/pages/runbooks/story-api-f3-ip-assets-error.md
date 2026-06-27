---
title: "Story-API F3: IP Assets Retrieval Connection Refused"
type: "runbook"
slug: "runbooks/story-api-f3-ip-assets-error"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "connectivity"
  - "database"
  - "postgresql"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_c2d87b7d060c3c0bf0099b1dd8b2ebbf"
  - "srcrev_e7ef7cc94ee0d152fda04a7cd4e53b6b"
conflict_state: "none"
---

# Story-API F3: IP Assets Retrieval Connection Refused

## Summary

Incident where Story-API failed to retrieve IP assets due to a connection refused error to PostgreSQL at 10.32.101.130:5432, later resolved.

## Claims

- Story-API encountered an error: "failed to retrieve IP assets: dial tcp 10.32.101.130:5432: connect: connection refused" `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_56686ca7c285fd31f1384861584ac388` `source_revision_id=srcrev_c2d87b7d060c3c0bf0099b1dd8b2ebbf` `chunk_id=srcchunk_dae5e83512a72ee418ebd857f866045a` `native_locator=slack:C07K3J4JTH6:1781411724.233429:1781411724.233429` `source_timestamp=2026-06-14T04:35:24Z`
- Blake Huynh marked the corresponding Sentry issue STORY-API-F3 as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_56686ca7c285fd31f1384861584ac388` `source_revision_id=srcrev_e7ef7cc94ee0d152fda04a7cd4e53b6b` `chunk_id=srcchunk_8921c162bdee04d491f5537fbd5c63e7` `native_locator=slack:C07K3J4JTH6:1781411724.233429:1781630303.144819` `source_timestamp=2026-06-16T17:18:23Z`

## Sources

- `source_document_id`: `srcdoc_56686ca7c285fd31f1384861584ac388`
- `source_revision_id`: `srcrev_c2d87b7d060c3c0bf0099b1dd8b2ebbf`
