---
title: "Resume Activate Endpoint Removal Decision"
type: "decision"
slug: "decisions/resume-activate-endpoint-removal"
freshness: "2026-04-25T00:00:55Z"
tags:
  - "api"
  - "deprecation"
  - "resume"
owners: []
source_revision_ids:
  - "srcrev_7b7e007a8a94739b97efe9d01bbafc95"
  - "srcrev_f5cae4222356a0cc5bf50d0305795901"
conflict_state: "none"
---

# Resume Activate Endpoint Removal Decision

## Summary

Decision to remove the unused /v1/me/resumes/{resume_id}/activate endpoint.

## Claims

- The /v1/me/resumes/{resume_id}/activate endpoint was requested as a feature to let people upload their resume, but its current status is uncertain. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8fc90a224160b65d12023aaedc3e5b94` `source_revision_id=srcrev_7b7e007a8a94739b97efe9d01bbafc95` `chunk_id=srcchunk_d71abecf41760ef01fdb87d2cb7a0492` `native_locator=slack:C0AL7EKNHDF:1777075109.214869:1777075195.310899` `source_timestamp=2026-04-25T00:00:05Z`
- It is suggested to remove the endpoint if it is not in use. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8fc90a224160b65d12023aaedc3e5b94` `source_revision_id=srcrev_f5cae4222356a0cc5bf50d0305795901` `chunk_id=srcchunk_57406ab7efd3c450b200af41676bc394` `native_locator=slack:C0AL7EKNHDF:1777075109.214869:1777075255.716779` `source_timestamp=2026-04-25T00:00:55Z`

## Open Questions

- Is the resume activate endpoint still in use or planned?

## Sources

- `source_document_id`: `srcdoc_8fc90a224160b65d12023aaedc3e5b94`
- `source_revision_id`: `srcrev_b287af2a3dbf43d1c2258c2b045b85eb`
