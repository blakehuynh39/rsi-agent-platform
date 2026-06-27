---
title: "Story API Replacement Gas Re-estimate Failure"
type: "system"
slug: "systems/story-api-replacement-gas-failure"
freshness: "2026-06-16T17:18:24Z"
tags:
  - "batch-quarantine"
  - "failure"
  - "gas"
  - "re-estimate"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_096d48760d6d5f3ad89a4627e045f673"
  - "srcrev_21eaf0bbac9e029a043bd746f9d40848"
  - "srcrev_75bbe7f9ed7e482003a639354c5b37da"
conflict_state: "none"
---

# Story API Replacement Gas Re-estimate Failure

## Summary

The story-api experienced a deterministic failure in replacement gas re-estimation, resulting in batch quarantining. The issue (STORY-API-EK) was later resolved.

## Claims

- The story-api's replacement gas re-estimate failed deterministically, leading to the batch being quarantined. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9d20e04e2cc7ef75da2ae649d16cd825` `source_revision_id=srcrev_75bbe7f9ed7e482003a639354c5b37da` `chunk_id=srcchunk_1e47de9762ab44e8016629434de933f4` `native_locator=slack:C07K3J4JTH6:1780777693.854829:1780777693.854829` `source_timestamp=2026-06-06T20:28:13Z`
  - citation: `source_document_id=srcdoc_9d20e04e2cc7ef75da2ae649d16cd825` `source_revision_id=srcrev_096d48760d6d5f3ad89a4627e045f673` `chunk_id=srcchunk_3004b82b176ad24ded13422b5404f4e5` `native_locator=slack:C07K3J4JTH6:1780777693.854829:1780869853.041989` `source_timestamp=2026-06-07T22:04:13Z`
- The Sentry issue STORY-API-EK was resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9d20e04e2cc7ef75da2ae649d16cd825` `source_revision_id=srcrev_21eaf0bbac9e029a043bd746f9d40848` `chunk_id=srcchunk_e863c3fbb8c4a368efe08ee9e95044bc` `native_locator=slack:C07K3J4JTH6:1780777693.854829:1781630304.433319` `source_timestamp=2026-06-16T17:18:24Z`

## Sources

- `source_document_id`: `srcdoc_9d20e04e2cc7ef75da2ae649d16cd825`
- `source_revision_id`: `srcrev_75bbe7f9ed7e482003a639354c5b37da`
