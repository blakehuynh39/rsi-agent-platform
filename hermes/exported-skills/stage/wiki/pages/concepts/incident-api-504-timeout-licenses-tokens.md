---
title: "API 504 Timeout on /api/v4/licenses/tokens"
type: "concept"
slug: "concepts/incident-api-504-timeout-licenses-tokens"
freshness: "2026-02-28T16:00:43Z"
tags:
  - "504"
  - "incident"
  - "story-api"
  - "timeout"
owners: []
source_revision_ids:
  - "srcrev_18f0617856f1c661b0ff9dd6a47c955e"
  - "srcrev_e08e083bca31e32d5159b99fbe50cccd"
conflict_state: "none"
---

# API 504 Timeout on /api/v4/licenses/tokens

## Summary

A 504 Gateway Timeout occurred on story-api POST /api/v4/licenses/tokens. The issue (STORY-API-E7) was subsequently resolved.

## Claims

- The story-api POST /api/v4/licenses/tokens failed with a 504 Request timeout. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a718e26269bc36b61f397e98752d46c5` `source_revision_id=srcrev_18f0617856f1c661b0ff9dd6a47c955e` `chunk_id=srcchunk_04929d1542934825c8323e38fe698e94` `native_locator=slack:C07K3J4JTH6:1772235535.496989:1772235535.496989` `source_timestamp=2026-02-27T23:38:55Z`
- The issue was tracked as STORY-API-E7 and marked as resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a718e26269bc36b61f397e98752d46c5` `source_revision_id=srcrev_e08e083bca31e32d5159b99fbe50cccd` `chunk_id=srcchunk_70ff3389ced6f9a6b4c8ec287a4e9590` `native_locator=slack:C07K3J4JTH6:1772235535.496989:1772294443.007659` `source_timestamp=2026-02-28T16:00:43Z`

## Open Questions

- What caused the 504 timeout?

## Related Pages

- `story-api`

## Sources

- `source_document_id`: `srcdoc_a718e26269bc36b61f397e98752d46c5`
- `source_revision_id`: `srcrev_18f0617856f1c661b0ff9dd6a47c955e`
