---
title: "Decision: Resolve fatal story-api submitter tick failure"
type: "decision"
slug: "decisions/decision-resolve-story-api-submitter-tick-failure"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "fatal-error"
  - "incident"
  - "resolved"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_24a3de1606cc0eced090c86b248a5077"
  - "srcrev_a55130b04816d90cdbf12bea0f6a3d80"
conflict_state: "none"
---

# Decision: Resolve fatal story-api submitter tick failure

## Summary

The story-api submitter tick failed terminally, stopping the worker. The corresponding Sentry issue STORY-API-EP was marked resolved by Blake Huynh.

## Claims

- The story-api submitter experienced a fatal tick failure, causing the worker to terminate. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d7e8fbcb722e57fa90e9c7c2b1422060` `source_revision_id=srcrev_24a3de1606cc0eced090c86b248a5077` `chunk_id=srcchunk_affb84976b0c353ecc33ad58ae7cd022` `native_locator=slack:C07K3J4JTH6:1780827971.695219:1780827971.695219` `source_timestamp=2026-06-07T10:26:11Z`
- Blake Huynh marked the Sentry issue STORY-API-EP as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d7e8fbcb722e57fa90e9c7c2b1422060` `source_revision_id=srcrev_a55130b04816d90cdbf12bea0f6a3d80` `chunk_id=srcchunk_9029b0f2945cafaa2210cdee381d0f21` `native_locator=slack:C07K3J4JTH6:1780827971.695219:1781630302.918259` `source_timestamp=2026-06-16T17:18:22Z`

## Open Questions

- Was a root cause fix deployed, or was the issue closed without resolution?
- What caused the fatal tick failure in story-api submitter?

## Sources

- `source_document_id`: `srcdoc_d7e8fbcb722e57fa90e9c7c2b1422060`
- `source_revision_id`: `srcrev_24a3de1606cc0eced090c86b248a5077`
