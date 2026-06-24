---
title: "Story Orchestration Service"
type: "system"
slug: "systems/story-orchestration-service"
freshness: "2026-04-18T06:52:07Z"
tags:
  - "error"
  - "story-orchestration"
  - "timeout"
owners: []
source_revision_ids:
  - "srcrev_719bd46f2709b7be3e068d84b4652555"
  - "srcrev_a7df68cbb8869604c3f187601dedae5b"
  - "srcrev_fc2249db3c5c937a609f642ca588db8f"
conflict_state: "none"
---

# Story Orchestration Service

## Summary

The story-orchestration-service has experienced multiple request timeout errors (context deadline exceeded).

## Claims

- The story-orchestration-service logged a request timeout error: 'fmt.wrapError: request timeout: context deadline exceeded'. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_521f84d36c8b0382f53c871b53846bdd` `source_revision_id=srcrev_fc2249db3c5c937a609f642ca588db8f` `chunk_id=srcchunk_ae13e53de80c7764b760da6f09057428` `native_locator=slack:C08BWTULNPP:1770906899.979959:1770906899.979959` `source_timestamp=2026-02-12T14:34:59Z`
- The Sentry issue STORY-ORCHESTRATION-SERVICE-FK was marked as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_521f84d36c8b0382f53c871b53846bdd` `source_revision_id=srcrev_719bd46f2709b7be3e068d84b4652555` `chunk_id=srcchunk_fe22ca16c94bfff6f753d07336dfe6d4` `native_locator=slack:C08BWTULNPP:1770906899.979959:1772294460.476209` `source_timestamp=2026-02-28T16:01:00Z`
- A later occurrence of the same request timeout error was observed. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_521f84d36c8b0382f53c871b53846bdd` `source_revision_id=srcrev_a7df68cbb8869604c3f187601dedae5b` `chunk_id=srcchunk_cedb97c895d09331692adfdf9ab71b1e` `native_locator=slack:C08BWTULNPP:1770906899.979959:1776495127.641279` `source_timestamp=2026-04-18T06:52:07Z`

## Open Questions

- Why did the request timeout error reoccur after the Sentry issue STORY-ORCHESTRATION-SERVICE-FK was resolved?

## Sources

- `source_document_id`: `srcdoc_521f84d36c8b0382f53c871b53846bdd`
- `source_revision_id`: `srcrev_a7df68cbb8869604c3f187601dedae5b`
