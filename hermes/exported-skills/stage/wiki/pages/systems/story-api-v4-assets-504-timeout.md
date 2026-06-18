---
title: "Story API POST /api/v4/assets 504 Timeout Incident"
type: "system"
slug: "systems/story-api-v4-assets-504-timeout"
freshness: "2026-02-28T16:00:43Z"
tags:
  - "504"
  - "api"
  - "incident"
  - "story-api"
  - "timeout"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_b47705ef571b3a2757fbad08dce68924"
  - "srcrev_ef83f444d405237da80f15e989c62846"
conflict_state: "none"
---

# Story API POST /api/v4/assets 504 Timeout Incident

## Summary

On approximately 2026-02-26, the Story API endpoint POST /api/v4/assets experienced a 504 Gateway Timeout, causing request failures. The issue was later resolved by Blake Huynh marking the associated Sentry issue STORY-API-E2 as resolved.

## Claims

- POST /api/v4/assets endpoint returned a 504 Gateway Timeout. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_68c139891ce1b049e928df36324e310f` `source_revision_id=srcrev_ef83f444d405237da80f15e989c62846` `chunk_id=srcchunk_4e796fc77284ce29333f94888a7ee103` `native_locator=slack:C07K3J4JTH6:1772050530.748759:1772050530.748759` `source_timestamp=2026-02-25T20:15:30Z`
- Blake Huynh marked the associated Sentry issue STORY-API-E2 (issue #7292027469) as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_68c139891ce1b049e928df36324e310f` `source_revision_id=srcrev_b47705ef571b3a2757fbad08dce68924` `chunk_id=srcchunk_0094f667b8761d64a0b9655a9ab4297d` `native_locator=slack:C07K3J4JTH6:1772050530.748759:1772294443.011259` `source_timestamp=2026-02-28T16:00:43Z`

## Sources

- `source_document_id`: `srcdoc_68c139891ce1b049e928df36324e310f`
- `source_revision_id`: `srcrev_b47705ef571b3a2757fbad08dce68924`
