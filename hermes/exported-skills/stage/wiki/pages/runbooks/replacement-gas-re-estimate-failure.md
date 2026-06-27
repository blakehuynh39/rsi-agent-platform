---
title: "Replacement Gas Re-estimate Failure"
type: "runbook"
slug: "runbooks/replacement-gas-re-estimate-failure"
freshness: "2026-06-16T17:18:24Z"
tags:
  - "error"
  - "gas-estimation"
  - "quarantine"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_096d48760d6d5f3ad89a4627e045f673"
  - "srcrev_21eaf0bbac9e029a043bd746f9d40848"
conflict_state: "none"
---

# Replacement Gas Re-estimate Failure

## Summary

The story-api reported a deterministic failure in replacement gas re-estimate, leading to batch quarantining. The issue was resolved by Blake Huynh (Sentry issue STORY-API-EK).

## Claims

- The story-api reported a deterministic failure in replacement gas re-estimate, resulting in the batch being quarantined. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9d20e04e2cc7ef75da2ae649d16cd825` `source_revision_id=srcrev_096d48760d6d5f3ad89a4627e045f673` `chunk_id=srcchunk_3004b82b176ad24ded13422b5404f4e5` `native_locator=slack:C07K3J4JTH6:1780777693.854829:1780869853.041989` `source_timestamp=2026-06-07T22:04:13Z`
- Blake Huynh marked the Sentry issue STORY-API-EK as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9d20e04e2cc7ef75da2ae649d16cd825` `source_revision_id=srcrev_21eaf0bbac9e029a043bd746f9d40848` `chunk_id=srcchunk_e863c3fbb8c4a368efe08ee9e95044bc` `native_locator=slack:C07K3J4JTH6:1780777693.854829:1781630304.433319` `source_timestamp=2026-06-16T17:18:24Z`

## Open Questions

- How often does this error occur?
- Is there a permanent fix or workaround?
- What is the root cause of the replacement gas re-estimate failure?

## Sources

- `source_document_id`: `srcdoc_9d20e04e2cc7ef75da2ae649d16cd825`
- `source_revision_id`: `srcrev_096d48760d6d5f3ad89a4627e045f673`
