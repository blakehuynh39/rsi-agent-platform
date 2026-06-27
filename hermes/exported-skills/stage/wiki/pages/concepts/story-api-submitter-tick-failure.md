---
title: "Story API Submitter Tick Failure"
type: "concept"
slug: "concepts/story-api-submitter-tick-failure"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "error"
  - "reliability"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_3602078c2e8e61ec7b7010d5d636fed7"
  - "srcrev_58838e335b55f2017862e7a8dc510d1f"
  - "srcrev_64fa7baa5fa380f80d53c23dac443e7f"
  - "srcrev_f1ad49b06d587d53b75cc429491f15d9"
conflict_state: "none"
---

# Story API Submitter Tick Failure

## Summary

Repeated errors from story-api submitter tick resulted in a resolve of STORY-API-EQ issue.

## Claims

- Story API submitter repeatedly failed with 'tick failed; sleeping before retry'. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fdd871f83242f5b35cb041baff90bb62` `source_revision_id=srcrev_58838e335b55f2017862e7a8dc510d1f` `chunk_id=srcchunk_e85cfca96a993b6304a7057765b12a9f` `native_locator=slack:C07K3J4JTH6:1780832205.996339:1780832205.996339` `source_timestamp=2026-06-07T11:36:45Z`
  - citation: `source_document_id=srcdoc_fdd871f83242f5b35cb041baff90bb62` `source_revision_id=srcrev_f1ad49b06d587d53b75cc429491f15d9` `chunk_id=srcchunk_448b9c1e7be89520d52d5be2b413ced4` `native_locator=slack:C07K3J4JTH6:1780832205.996339:1781049619.709639` `source_timestamp=2026-06-10T00:00:19Z`
  - citation: `source_document_id=srcdoc_fdd871f83242f5b35cb041baff90bb62` `source_revision_id=srcrev_64fa7baa5fa380f80d53c23dac443e7f` `chunk_id=srcchunk_8d8ab13859e49cbbc1fc50bae83608d5` `native_locator=slack:C07K3J4JTH6:1780832205.996339:1781411633.417849` `source_timestamp=2026-06-14T04:33:53Z`
- The issue STORY-API-EQ was resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fdd871f83242f5b35cb041baff90bb62` `source_revision_id=srcrev_3602078c2e8e61ec7b7010d5d636fed7` `chunk_id=srcchunk_5e613146053dbeeb1b88d775c85ae469` `native_locator=slack:C07K3J4JTH6:1780832205.996339:1781630303.010979` `source_timestamp=2026-06-16T17:18:23Z`

## Sources

- `source_document_id`: `srcdoc_fdd871f83242f5b35cb041baff90bb62`
- `source_revision_id`: `srcrev_58838e335b55f2017862e7a8dc510d1f`
