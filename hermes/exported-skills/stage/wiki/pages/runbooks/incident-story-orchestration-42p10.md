---
title: "story-orchestration-service: ON CONFLICT constraint error (42P10)"
type: "runbook"
slug: "runbooks/incident-story-orchestration-42p10"
freshness: "2026-03-02T19:34:39Z"
tags:
  - "database"
  - "incident"
  - "postgresql"
  - "story-orchestration-service"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_3976e362609c42e5630daca91454e848"
  - "srcrev_8a9356661dd55c49f66dfcd041debb66"
conflict_state: "none"
---

# story-orchestration-service: ON CONFLICT constraint error (42P10)

## Summary

story-orchestration-service experienced a PostgreSQL error #42P10 due to a missing unique or exclusion constraint for an ON CONFLICT clause. The Sentry issue was resolved by Blake Huynh.

## Claims

- story-orchestration-service encountered error #42P10: there is no unique or exclusion constraint matching the ON CONFLICT specification. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80486e8e8d4b883ae426064c39b30548` `source_revision_id=srcrev_3976e362609c42e5630daca91454e848` `chunk_id=srcchunk_3bb561ee7b839b70901b5a00a0b90531` `native_locator=slack:C08BWTULNPP:1772460164.208669:1772460164.208669` `source_timestamp=2026-03-02T14:02:44Z`
- The Sentry issue STORY-ORCHESTRATION-SERVICE-FY was marked as resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80486e8e8d4b883ae426064c39b30548` `source_revision_id=srcrev_8a9356661dd55c49f66dfcd041debb66` `chunk_id=srcchunk_2e215895b064b8cd494e2f76420d4411` `native_locator=slack:C08BWTULNPP:1772460164.208669:1772480079.841629` `source_timestamp=2026-03-02T19:34:39Z`

## Sources

- `source_document_id`: `srcdoc_80486e8e8d4b883ae426064c39b30548`
- `source_revision_id`: `srcrev_3976e362609c42e5630daca91454e848`
