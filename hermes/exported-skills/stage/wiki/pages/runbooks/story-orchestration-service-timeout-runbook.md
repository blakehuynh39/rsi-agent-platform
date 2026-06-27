---
title: "Story Orchestration Service Timeout Runbook"
type: "runbook"
slug: "runbooks/story-orchestration-service-timeout-runbook"
freshness: "2026-04-18T06:52:07Z"
tags:
  - "error"
  - "sentry"
  - "story-orchestration-service"
  - "timeout"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_719bd46f2709b7be3e068d84b4652555"
  - "srcrev_a7df68cbb8869604c3f187601dedae5b"
  - "srcrev_fc2249db3c5c937a609f642ca588db8f"
conflict_state: "none"
---

# Story Orchestration Service Timeout Runbook

## Summary

Handles occurrences of request timeout: context deadline exceeded in the story-orchestration-service.

## Claims

- The story-orchestration-service emitted a "request timeout: context deadline exceeded" error on at least two occasions. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_521f84d36c8b0382f53c871b53846bdd` `source_revision_id=srcrev_fc2249db3c5c937a609f642ca588db8f` `chunk_id=srcchunk_ae13e53de80c7764b760da6f09057428` `native_locator=slack:C08BWTULNPP:1770906899.979959:1770906899.979959` `source_timestamp=2026-02-12T14:34:59Z`
  - citation: `source_document_id=srcdoc_521f84d36c8b0382f53c871b53846bdd` `source_revision_id=srcrev_a7df68cbb8869604c3f187601dedae5b` `chunk_id=srcchunk_cedb97c895d09331692adfdf9ab71b1e` `native_locator=slack:C08BWTULNPP:1770906899.979959:1776495127.641279` `source_timestamp=2026-04-18T06:52:07Z`
- Blake Huynh marked the Sentry issue STORY-ORCHESTRATION-SERVICE-FK as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_521f84d36c8b0382f53c871b53846bdd` `source_revision_id=srcrev_719bd46f2709b7be3e068d84b4652555` `chunk_id=srcchunk_fe22ca16c94bfff6f753d07336dfe6d4` `native_locator=slack:C08BWTULNPP:1770906899.979959:1772294460.476209` `source_timestamp=2026-02-28T16:01:00Z`
- The same timeout error recurred after the Sentry issue was resolved. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_521f84d36c8b0382f53c871b53846bdd` `source_revision_id=srcrev_a7df68cbb8869604c3f187601dedae5b` `chunk_id=srcchunk_cedb97c895d09331692adfdf9ab71b1e` `native_locator=slack:C08BWTULNPP:1770906899.979959:1776495127.641279` `source_timestamp=2026-04-18T06:52:07Z`

## Open Questions

- Is there a permanent fix to prevent recurrence?
- What is the root cause of the timeout?

## Related Pages

- `story-orchestration-service`

## Sources

- `source_document_id`: `srcdoc_521f84d36c8b0382f53c871b53846bdd`
- `source_revision_id`: `srcrev_719bd46f2709b7be3e068d84b4652555`
