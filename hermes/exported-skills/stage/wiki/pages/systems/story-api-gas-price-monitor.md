---
title: "Story API Gas Price Monitor"
type: "system"
slug: "systems/story-api-gas-price-monitor"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "gas-price"
  - "incident"
  - "monitor"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_8dc7fa2a3a30fde9fe84407453101344"
  - "srcrev_f952827b856f3ab36283a3b4357b665a"
conflict_state: "none"
---

# Story API Gas Price Monitor

## Summary

Monitor component for gas prices in story-api. Experienced an RPC failure, mitigated by holding last-good snapshot and gate vote. Issue resolved.

## Claims

- The gas-price monitor experienced an RPC failure. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab13a9de3a06e4ee653754cfea366fc2` `source_revision_id=srcrev_8dc7fa2a3a30fde9fe84407453101344` `chunk_id=srcchunk_7b76451ede2850eb072e88286603ddb2` `native_locator=slack:C07K3J4JTH6:1781160623.732939:1781160623.732939` `source_timestamp=2026-06-11T06:50:23Z`
- The system held a last-good snapshot and gate vote in response to the RPC failure. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab13a9de3a06e4ee653754cfea366fc2` `source_revision_id=srcrev_8dc7fa2a3a30fde9fe84407453101344` `chunk_id=srcchunk_7b76451ede2850eb072e88286603ddb2` `native_locator=slack:C07K3J4JTH6:1781160623.732939:1781160623.732939` `source_timestamp=2026-06-11T06:50:23Z`
- Blake Huynh marked the related Sentry issue STORY-API-EX as resolved. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab13a9de3a06e4ee653754cfea366fc2` `source_revision_id=srcrev_f952827b856f3ab36283a3b4357b665a` `chunk_id=srcchunk_1c8cc20825c53bd5ede258506b26cb4b` `native_locator=slack:C07K3J4JTH6:1781160623.732939:1781630302.930409` `source_timestamp=2026-06-16T17:18:22Z`

## Open Questions

- What caused the RPC failure?

## Sources

- `source_document_id`: `srcdoc_ab13a9de3a06e4ee653754cfea366fc2`
- `source_revision_id`: `srcrev_8dc7fa2a3a30fde9fe84407453101344`
