---
title: "Replacement Gas Re-estimate Deterministic Failure in Story API"
type: "concept"
slug: "concepts/story-api-replacement-gas-reestimate-failure"
freshness: "2026-06-16T17:18:24Z"
tags:
  - "failure"
  - "gas"
  - "incident"
  - "quarantine"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_21eaf0bbac9e029a043bd746f9d40848"
  - "srcrev_75bbe7f9ed7e482003a639354c5b37da"
conflict_state: "none"
---

# Replacement Gas Re-estimate Deterministic Failure in Story API

## Summary

A replacement gas re-estimate failed deterministically, causing the batch to be quarantined. The incident was tracked as Sentry issue STORY-API-EK and resolved.

## Claims

- Replacement gas re-estimate failed deterministically in story-api. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9d20e04e2cc7ef75da2ae649d16cd825` `source_revision_id=srcrev_75bbe7f9ed7e482003a639354c5b37da` `chunk_id=srcchunk_1e47de9762ab44e8016629434de933f4` `native_locator=slack:C07K3J4JTH6:1780777693.854829:1780777693.854829` `source_timestamp=2026-06-06T20:28:13Z`
- The batch was quarantined due to the failure. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9d20e04e2cc7ef75da2ae649d16cd825` `source_revision_id=srcrev_75bbe7f9ed7e482003a639354c5b37da` `chunk_id=srcchunk_1e47de9762ab44e8016629434de933f4` `native_locator=slack:C07K3J4JTH6:1780777693.854829:1780777693.854829` `source_timestamp=2026-06-06T20:28:13Z`
- Blake Huynh marked Sentry issue STORY-API-EK as resolved. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9d20e04e2cc7ef75da2ae649d16cd825` `source_revision_id=srcrev_21eaf0bbac9e029a043bd746f9d40848` `chunk_id=srcchunk_e863c3fbb8c4a368efe08ee9e95044bc` `native_locator=slack:C07K3J4JTH6:1780777693.854829:1781630304.433319` `source_timestamp=2026-06-16T17:18:24Z`

## Open Questions

- Was the batch retried successfully after resolution?
- What caused the deterministic failure in replacement gas re-estimate?

## Sources

- `source_document_id`: `srcdoc_9d20e04e2cc7ef75da2ae649d16cd825`
- `source_revision_id`: `srcrev_21eaf0bbac9e029a043bd746f9d40848`
