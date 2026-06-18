---
title: "Story API Lineage Timeout Incident"
type: "runbook"
slug: "runbooks/story-api-lineage-timeout-incident"
freshness: "2026-04-29T18:11:21Z"
tags:
  - "api"
  - "lineage"
  - "story-api"
  - "timeout"
owners: []
source_revision_ids:
  - "srcrev_97ad9e1c0a043d6f9715b3cf2261c5ed"
  - "srcrev_eb1df81a50f1368a0830c2b496981840"
conflict_state: "none"
---

# Story API Lineage Timeout Incident

## Summary

The story-api POST /api/v4/assets/lineage endpoint returned a 504 timeout error.

## Claims

- The story-api POST /api/v4/assets/lineage endpoint returned a 504 Request timeout. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_523f9ccef40672ee919bd2357e2b95d0` `source_revision_id=srcrev_97ad9e1c0a043d6f9715b3cf2261c5ed` `chunk_id=srcchunk_155cfa8643e7474271f29c7b46bb6952` `native_locator=slack:C07K3J4JTH6:1777486222.977879:1777486222.977879` `source_timestamp=2026-04-29T18:10:22Z`
- A user was mentioned in the Slack thread. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_523f9ccef40672ee919bd2357e2b95d0` `source_revision_id=srcrev_eb1df81a50f1368a0830c2b496981840` `chunk_id=srcchunk_b23dd2d4101859fa38cbc7c7fb0c2cf7` `native_locator=slack:C07K3J4JTH6:1777486222.977879:1777486281.368249` `source_timestamp=2026-04-29T18:11:21Z`

## Open Questions

- Has the issue been resolved?
- What caused the timeout?

## Sources

- `source_document_id`: `srcdoc_523f9ccef40672ee919bd2357e2b95d0`
- `source_revision_id`: `srcrev_eb1df81a50f1368a0830c2b496981840`
