---
title: "story-api: ERROR #23502 null created_at in notifications"
type: "runbook"
slug: "runbooks/story-api-null-created-at-notifications"
freshness: "2026-02-26T01:50:31Z"
tags:
  - "error"
  - "notifications"
  - "postgres"
  - "story-api"
owners: []
source_revision_ids:
  - "srcrev_510988811487a6008baf170f55527794"
  - "srcrev_53af6f7602bb2301519fd8067a605ef4"
conflict_state: "none"
---

# story-api: ERROR #23502 null created_at in notifications

## Summary

On 2026-02-26, story-api encountered a PostgreSQL not-null constraint violation (error #23502) when inserting into the notifications table with a null created_at value.

## Claims

- story-api reported error #23502: null value in column "created_at" of relation "notifications" violates not-null constraint `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1ab5f1ec6019cd948bbac151e798eb86` `source_revision_id=srcrev_53af6f7602bb2301519fd8067a605ef4` `chunk_id=srcchunk_ef45f58f823d2f203bbafee970e72fe6` `native_locator=slack:C07K3J4JTH6:1772070585.340179:1772070585.340179` `source_timestamp=2026-02-26T01:49:45Z`
- A message in the Slack thread included a mention of user U0AC11JV8AX. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1ab5f1ec6019cd948bbac151e798eb86` `source_revision_id=srcrev_510988811487a6008baf170f55527794` `chunk_id=srcchunk_e9d6be370d17aff652d4c980686b20c6` `native_locator=slack:C07K3J4JTH6:1772070585.340179:1772070631.716409` `source_timestamp=2026-02-26T01:50:31Z`

## Open Questions

- Is this a recurring issue?
- What caused the created_at to be null?
- What is the expected remediation?

## Sources

- `source_document_id`: `srcdoc_1ab5f1ec6019cd948bbac151e798eb86`
- `source_revision_id`: `srcrev_510988811487a6008baf170f55527794`
