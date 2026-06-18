---
title: "504 Timeout on POST /api/v4/licenses/tokens"
type: "runbook"
slug: "runbooks/post-api-v4-licenses-tokens-504-timeout"
freshness: "2026-02-28T16:00:43Z"
tags:
  - "504"
  - "incident"
  - "resolved"
  - "story-api"
  - "timeout"
owners: []
source_revision_ids:
  - "srcrev_18f0617856f1c661b0ff9dd6a47c955e"
  - "srcrev_e08e083bca31e32d5159b99fbe50cccd"
conflict_state: "none"
---

# 504 Timeout on POST /api/v4/licenses/tokens

## Summary

The story-api endpoint POST /api/v4/licenses/tokens experienced a 504 Request timeout. Incident tracked as STORY-API-E7 and resolved.

## Claims

- The story-api endpoint POST /api/v4/licenses/tokens returned a 504 Request timeout error. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a718e26269bc36b61f397e98752d46c5` `source_revision_id=srcrev_18f0617856f1c661b0ff9dd6a47c955e` `chunk_id=srcchunk_04929d1542934825c8323e38fe698e94` `native_locator=slack:C07K3J4JTH6:1772235535.496989:1772235535.496989` `source_timestamp=2026-02-27T23:38:55Z`
- Blake Huynh marked Sentry issue STORY-API-E7 as resolved, indicating the incident was resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a718e26269bc36b61f397e98752d46c5` `source_revision_id=srcrev_e08e083bca31e32d5159b99fbe50cccd` `chunk_id=srcchunk_70ff3389ced6f9a6b4c8ec287a4e9590` `native_locator=slack:C07K3J4JTH6:1772235535.496989:1772294443.007659` `source_timestamp=2026-02-28T16:00:43Z`

## Sources

- `source_document_id`: `srcdoc_a718e26269bc36b61f397e98752d46c5`
- `source_revision_id`: `srcrev_e08e083bca31e32d5159b99fbe50cccd`
