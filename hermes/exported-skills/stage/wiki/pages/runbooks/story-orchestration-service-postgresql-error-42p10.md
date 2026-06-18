---
title: "Story Orchestration Service: PostgreSQL Error 42P10"
type: "runbook"
slug: "runbooks/story-orchestration-service-postgresql-error-42p10"
freshness: "2026-02-28T16:00:58Z"
tags:
  - "database"
  - "incident"
  - "postgresql"
  - "sentry"
  - "story-orchestration-service"
owners: []
source_revision_ids:
  - "srcrev_1a138ad5bb681d984c1c961b05adf23e"
  - "srcrev_86ea418f26d378160921f10b9480b8a9"
conflict_state: "none"
---

# Story Orchestration Service: PostgreSQL Error 42P10

## Summary

A PostgreSQL error 42P10 (no unique or exclusion constraint matching ON CONFLICT) was encountered in the story-orchestration-service. The issue was resolved via Sentry issue #7298595772.

## Claims

- The story-orchestration-service encountered a PostgreSQL error #42P10: no unique or exclusion constraint matching the ON CONFLICT specification. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_27d9e99ff7bf6ad346cfcca3547bc855` `source_revision_id=srcrev_1a138ad5bb681d984c1c961b05adf23e` `chunk_id=srcchunk_c4af253d4fb528a6a0a721c2ac113031` `native_locator=slack:C08BWTULNPP:1772273448.650679` `source_timestamp=2026-02-28T10:10:48Z`
- Blake Huynh marked the Sentry issue STORY-ORCHESTRATION-SERVICE-FX (issue #7298595772) as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_27d9e99ff7bf6ad346cfcca3547bc855` `source_revision_id=srcrev_86ea418f26d378160921f10b9480b8a9` `chunk_id=srcchunk_642ce387a7281c911a3f3e30e47110af` `native_locator=slack:C08BWTULNPP:1772294458.858989` `source_timestamp=2026-02-28T16:00:58Z`
- The error was tracked as Sentry issue #7298595772 (STORY-ORCHESTRATION-SERVICE-FX). `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_27d9e99ff7bf6ad346cfcca3547bc855` `source_revision_id=srcrev_1a138ad5bb681d984c1c961b05adf23e` `chunk_id=srcchunk_c4af253d4fb528a6a0a721c2ac113031` `native_locator=slack:C08BWTULNPP:1772273448.650679` `source_timestamp=2026-02-28T10:10:48Z`
  - citation: `source_document_id=srcdoc_27d9e99ff7bf6ad346cfcca3547bc855` `source_revision_id=srcrev_86ea418f26d378160921f10b9480b8a9` `chunk_id=srcchunk_642ce387a7281c911a3f3e30e47110af` `native_locator=slack:C08BWTULNPP:1772294458.858989` `source_timestamp=2026-02-28T16:00:58Z`

## Sources

- `source_document_id`: `srcdoc_27d9e99ff7bf6ad346cfcca3547bc855`
- `source_revision_id`: `srcrev_86ea418f26d378160921f10b9480b8a9`
