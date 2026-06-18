---
title: "Story API Staging Outage on 2026-06-16"
type: "system"
slug: "systems/story-api-staging-outage-2026-06-16"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "incident"
  - "postgres"
  - "postmortem"
  - "staging"
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

# Story API Staging Outage on 2026-06-16

## Summary

At 04:32 UTC on 2026-06-16, the staging story-api experienced 500 errors on GET /api/v1/data-audit/search due to a transient PostgreSQL outage. The outage lasted ~20s and resolved automatically; all services returned to healthy state.

## Claims

- GET /api/v1/data-audit/search failed with HTTP 500 and an EOF error. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_fbd786b7890711fb13d57a2458d62ee1` `chunk_id=srcchunk_7cc595f9cbd7967e100490a0f813dc14` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411569.669329` `source_timestamp=2026-06-14T04:32:49Z`
- Root cause was a transient staging PostgreSQL outage due to an RDS restart. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_e346cd593776bb33cea16a601904c746` `chunk_id=srcchunk_d3b8d6f78d93030801199fb3d4b7e936` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411783.978009` `source_timestamp=2026-06-14T04:36:23Z`
- The PostgreSQL outage affected both the analytics DB (10.64.201.95) and protocol DB (10.32.100.216), which were unreachable for approximately 20 seconds. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_e346cd593776bb33cea16a601904c746` `chunk_id=srcchunk_d3b8d6f78d93030801199fb3d4b7e936` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411783.978009` `source_timestamp=2026-06-14T04:36:23Z`
- The outage has fully resolved; all pods are healthy and 5xx error rates returned to zero. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_e346cd593776bb33cea16a601904c746` `chunk_id=srcchunk_d3b8d6f78d93030801199fb3d4b7e936` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411783.978009` `source_timestamp=2026-06-14T04:36:23Z`
- A Sentry issue (STORY-API-EY) was created for the error and marked resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_f204c7b5c129f1f99211858cdc9ffaf2` `chunk_id=srcchunk_0e74a216ca9de629c78a678caf8ad2a5` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781630302.891229` `source_timestamp=2026-06-16T17:18:22Z`
- The incident was tracked in an RSI run session. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_2a9d63ce9c2b552bac8edb539bbe4ec1` `chunk_id=srcchunk_cc984ae0f8de15b3d765351bf45e9988` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411608.570789` `source_timestamp=2026-06-14T04:33:28Z`

## Sources

- `source_document_id`: `srcdoc_fc3dde93c7db94f921c29a9b6624b2c0`
- `source_revision_id`: `srcrev_f204c7b5c129f1f99211858cdc9ffaf2`
