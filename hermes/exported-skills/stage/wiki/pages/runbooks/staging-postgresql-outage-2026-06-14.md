---
title: "Staging PostgreSQL Outage (2026-06-14)"
type: "runbook"
slug: "runbooks/staging-postgresql-outage-2026-06-14"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "incident"
  - "postgresql"
  - "rsi"
  - "staging"
owners: []
source_revision_ids:
  - "srcrev_2a9d63ce9c2b552bac8edb539bbe4ec1"
  - "srcrev_e346cd593776bb33cea16a601904c746"
  - "srcrev_f204c7b5c129f1f99211858cdc9ffaf2"
  - "srcrev_fbd786b7890711fb13d57a2458d62ee1"
conflict_state: "none"
---

# Staging PostgreSQL Outage (2026-06-14)

## Summary

A ~20-second outage of staging PostgreSQL (RDS restart) on 2026-06-14 at 04:32 UTC caused story-api 500 errors. The incident resolved automatically and was tracked in Sentry as STORY-API-EY.

## Claims

- story-api GET /api/v1/data-audit/search failed with 500: EOF `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_fbd786b7890711fb13d57a2458d62ee1` `chunk_id=srcchunk_7cc595f9cbd7967e100490a0f813dc14` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411569.669329` `source_timestamp=2026-06-14T04:32:49Z`
- Root cause: Transient staging PostgreSQL outage (RDS restart). Both analytics DB (10.64.201.95) and protocol DB (10.32.100.216) were unreachable for ~20s at 04:32 UTC. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_e346cd593776bb33cea16a601904c746` `chunk_id=srcchunk_d3b8d6f78d93030801199fb3d4b7e936` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411783.978009` `source_timestamp=2026-06-14T04:36:23Z`
- The outage fully resolved, all pods are healthy and 5xx error rates are back to 0. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_e346cd593776bb33cea16a601904c746` `chunk_id=srcchunk_d3b8d6f78d93030801199fb3d4b7e936` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411783.978009` `source_timestamp=2026-06-14T04:36:23Z`
- Blake Huynh marked Sentry issue STORY-API-EY as resolved. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_f204c7b5c129f1f99211858cdc9ffaf2` `chunk_id=srcchunk_0e74a216ca9de629c78a678caf8ad2a5` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781630302.891229` `source_timestamp=2026-06-16T17:18:22Z`
- The incident was tracked in an RSI run with trace link: https://staging-rsi-platform.storyprotocol.net/sessions?conversation=conv-0f3b5fe5cbc24d89a15c8a046d5a4cfc&tab=conversations&trace=trace-70159f59ed204d7c8725d9317c42caa1. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc3dde93c7db94f921c29a9b6624b2c0` `source_revision_id=srcrev_2a9d63ce9c2b552bac8edb539bbe4ec1` `chunk_id=srcchunk_cc984ae0f8de15b3d765351bf45e9988` `native_locator=slack:C07K3J4JTH6:1781411569.669329:1781411608.570789` `source_timestamp=2026-06-14T04:33:28Z`

## Sources

- `source_document_id`: `srcdoc_fc3dde93c7db94f921c29a9b6624b2c0`
- `source_revision_id`: `srcrev_e346cd593776bb33cea16a601904c746`
