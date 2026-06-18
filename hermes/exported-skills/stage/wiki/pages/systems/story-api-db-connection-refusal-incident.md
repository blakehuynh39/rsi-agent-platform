---
title: "Story API Database Connection Refusal Incident"
type: "system"
slug: "systems/story-api-db-connection-refusal-incident"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "database"
  - "incident"
  - "postgresql"
  - "resolved"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_c2d87b7d060c3c0bf0099b1dd8b2ebbf"
  - "srcrev_e7ef7cc94ee0d152fda04a7cd4e53b6b"
conflict_state: "none"
---

# Story API Database Connection Refusal Incident

## Summary

On 2026-06-16, story-api encountered a PostgreSQL connection refusal error. The incident was tracked as STORY-API-F3 and resolved by blake.huynh.

## Claims

- The story-api service produced an error: "failed to retrieve IP assets: dial tcp 10.32.101.130:5432: connect: connection refused". `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_56686ca7c285fd31f1384861584ac388` `source_revision_id=srcrev_c2d87b7d060c3c0bf0099b1dd8b2ebbf` `chunk_id=srcchunk_dae5e83512a72ee418ebd857f866045a` `native_locator=slack:C07K3J4JTH6:1781411724.233429:1781411724.233429` `source_timestamp=2026-06-14T04:35:24Z`
- blake.huynh marked the Sentry issue STORY-API-F3 as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_56686ca7c285fd31f1384861584ac388` `source_revision_id=srcrev_e7ef7cc94ee0d152fda04a7cd4e53b6b` `chunk_id=srcchunk_8921c162bdee04d491f5537fbd5c63e7` `native_locator=slack:C07K3J4JTH6:1781411724.233429:1781630303.144819` `source_timestamp=2026-06-16T17:18:23Z`

## Open Questions

- Was there any broader impact on IP asset retrieval?
- What caused the database connection refusal?

## Sources

- `source_document_id`: `srcdoc_56686ca7c285fd31f1384861584ac388`
- `source_revision_id`: `srcrev_e7ef7cc94ee0d152fda04a7cd4e53b6b`
