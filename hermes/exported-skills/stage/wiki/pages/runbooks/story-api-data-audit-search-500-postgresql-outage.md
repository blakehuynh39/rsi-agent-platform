---
title: "story-api: data-audit/search 500 due to PostgreSQL outage"
type: "runbook"
slug: "runbooks/story-api-data-audit-search-500-postgresql-outage"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "data-audit"
  - "incident"
  - "postgresql"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_2a9d63ce9c2b552bac8edb539bbe4ec1"
  - "srcrev_e346cd593776bb33cea16a601904c746"
  - "srcrev_f204c7b5c129f1f99211858cdc9ffaf2"
  - "srcrev_fbd786b7890711fb13d57a2458d62ee1"
conflict_state: "none"
---

# story-api: data-audit/search 500 due to PostgreSQL outage

## Summary

On 2026-06-14 at ~04:32 UTC, story-api GET /api/v1/data-audit/search returned 500 errors due to a transient staging PostgreSQL RDS restart causing both analytics and protocol DBs to be unreachable for ~20s. The outage resolved automatically, 5xx error rates returned to 0, and the Sentry issue was marked resolved.

## Claims

- story-api GET /api/v1/data-audit/search failed with 500: EOF `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_fbd786b7890711fb13d57a2458d62ee1` `chunk_id=srcchunk_7cc595f9cbd7967e100490a0f813dc14` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411569.669329` `source_timestamp=2026-06-14T04:32:49Z`
- Root cause: Transient staging PostgreSQL outage (RDS restart). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_e346cd593776bb33cea16a601904c746` `chunk_id=srcchunk_d3b8d6f78d93030801199fb3d4b7e936` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411783.978009` `source_timestamp=2026-06-14T04:36:23Z`
- Both analytics DB (10.64.201.95) and protocol DB (10.32.100.216) were unreachable for ~20s at 04:32 UTC. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_e346cd593776bb33cea16a601904c746` `chunk_id=srcchunk_d3b8d6f78d93030801199fb3d4b7e936` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411783.978009` `source_timestamp=2026-06-14T04:36:23Z`
- The outage has fully resolved — all pods are healthy and 5xx error rates are back to 0. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_e346cd593776bb33cea16a601904c746` `chunk_id=srcchunk_d3b8d6f78d93030801199fb3d4b7e936` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411783.978009` `source_timestamp=2026-06-14T04:36:23Z`
- RSI run trace available at https://staging-rsi-platform.storyprotocol.net/sessions?conversation=conv-0f3b5fe5cbc24d89a15c8a046d5a4cfc&tab=conversations&trace=trace-70159f59ed204d7c8725d9317c42caa1 `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_2a9d63ce9c2b552bac8edb539bbe4ec1` `chunk_id=srcchunk_cc984ae0f8de15b3d765351bf45e9988` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411608.570789` `source_timestamp=2026-06-14T04:33:28Z`
- blake.huynh@storyprotocol.xyz marked Sentry issue STORY-API-EY as resolved. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_f204c7b5c129f1f99211858cdc9ffaf2` `chunk_id=srcchunk_0e74a216ca9de629c78a678caf8ad2a5` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781630302.891229` `source_timestamp=2026-06-16T17:18:22Z`

## Related Pages

- `rsi-platform`
- `staging-postgresql`
- `story-api`

## Sources

- `source_document_id`: `srcdoc_fc3dde93c7db94f921c29a9b6624b2c0`
- `source_revision_id`: `srcrev_f4c907980a9307352cc250effdfbf793`
