---
title: "Story API Submitter Tick Failure Incident"
type: "system"
slug: "systems/story-api-submitter-tick-failure"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "error"
  - "sentry"
  - "story-api"
owners:
  - "Blake Huynh"
source_revision_ids:
  - "srcrev_3602078c2e8e61ec7b7010d5d636fed7"
  - "srcrev_58838e335b55f2017862e7a8dc510d1f"
conflict_state: "none"
---

# Story API Submitter Tick Failure Incident

## Summary

The story-api system experienced submitter tick failures, causing the process to sleep before retrying. The issue was tracked as STORY-API-EQ in Sentry and resolved by Blake Huynh.

## Claims

- The story-api submitter encountered a tick failure, and the system slept before retrying. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fdd871f83242f5b35cb041baff90bb62` `source_revision_id=srcrev_58838e335b55f2017862e7a8dc510d1f` `chunk_id=srcchunk_e85cfca96a993b6304a7057765b12a9f` `native_locator=slack:C07K3J4JTH6:1780832205.996339:1780832205.996339` `source_timestamp=2026-06-07T11:36:45Z`
- The issue was tracked as Sentry issue STORY-API-EQ and was marked resolved by Blake Huynh. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fdd871f83242f5b35cb041baff90bb62` `source_revision_id=srcrev_3602078c2e8e61ec7b7010d5d636fed7` `chunk_id=srcchunk_5e613146053dbeeb1b88d775c85ae469` `native_locator=slack:C07K3J4JTH6:1780832205.996339:1781630303.010979` `source_timestamp=2026-06-16T17:18:23Z`

## Sources

- `source_document_id`: `srcdoc_fdd871f83242f5b35cb041baff90bb62`
- `source_revision_id`: `srcrev_3602078c2e8e61ec7b7010d5d636fed7`
