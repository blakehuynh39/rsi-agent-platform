---
title: "Story Orchestration Service"
type: "system"
slug: "systems/story-orchestration-service"
freshness: "2026-06-04T17:17:00Z"
tags:
  - "golang"
  - "microservice"
  - "temporal"
owners: []
source_revision_ids:
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
conflict_state: "none"
---

# Story Orchestration Service

## Summary

Production microservice that orchestrates story workflows using Temporal. Recently experienced a nil pointer dereference bug (see incident runbook).

## Claims

- story-orchestration-service runs in production. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- It uses Temporal workflow handler methods that can dereference nil pointers. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`

## Related Pages

- `commit-signing-requirements`
- `nil-pointer-dereference-in-story-orchestration-service-2026-06-04`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_c324be3df9811270afe5ab85f79a8722`
