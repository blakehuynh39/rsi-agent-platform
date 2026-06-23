---
title: "Story-API Block-Util Observer RPC Failure"
type: "system"
slug: "systems/story-api-block-util-observer-rpc-failure"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "block-util"
  - "rpc-failure"
  - "story-api"
owners: []
source_revision_ids:
  - "srcrev_3d2d1e1e1900d2c3006c91c514d1133d"
  - "srcrev_b2a356be6866920cfe31071a362b9a1b"
conflict_state: "none"
---

# Story-API Block-Util Observer RPC Failure

## Summary

The story-api block-util observer experienced an RPC failure, causing the gate to fail open. The issue was marked resolved by Blake Huynh.

## Claims

- The story-api block-util observer experienced an RPC failure. `claim:rpc-failure-occurred` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b33dadb4668fdb482b79e351050b6fb5` `source_revision_id=srcrev_3d2d1e1e1900d2c3006c91c514d1133d` `chunk_id=srcchunk_3bd600a0385404fb2fd3d5a7da367025` `native_locator=slack:C07K3J4JTH6:1780900837.053459:1780900837.053459` `source_timestamp=2026-06-08T06:40:37Z`
- The gate was failing open as a result of the RPC failure. `claim:gate-fail-open` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b33dadb4668fdb482b79e351050b6fb5` `source_revision_id=srcrev_3d2d1e1e1900d2c3006c91c514d1133d` `chunk_id=srcchunk_3bd600a0385404fb2fd3d5a7da367025` `native_locator=slack:C07K3J4JTH6:1780900837.053459:1780900837.053459` `source_timestamp=2026-06-08T06:40:37Z`
- Blake Huynh resolved the issue. `claim:issue-resolved-by-blake` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b33dadb4668fdb482b79e351050b6fb5` `source_revision_id=srcrev_b2a356be6866920cfe31071a362b9a1b` `chunk_id=srcchunk_09448e3f52b07fbb044a4334e55e30d4` `native_locator=slack:C07K3J4JTH6:1780900837.053459:1781630303.024849` `source_timestamp=2026-06-16T17:18:23Z`
- The Sentry issue STORY-API-ET was marked resolved. `claim:sentry-issue-marked-resolved` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b33dadb4668fdb482b79e351050b6fb5` `source_revision_id=srcrev_b2a356be6866920cfe31071a362b9a1b` `chunk_id=srcchunk_09448e3f52b07fbb044a4334e55e30d4` `native_locator=slack:C07K3J4JTH6:1780900837.053459:1781630303.024849` `source_timestamp=2026-06-16T17:18:23Z`

## Open Questions

- What caused the RPC failure?

## Sources

- `source_document_id`: `srcdoc_b33dadb4668fdb482b79e351050b6fb5`
- `source_revision_id`: `srcrev_b2a356be6866920cfe31071a362b9a1b`
