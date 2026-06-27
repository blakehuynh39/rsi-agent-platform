---
title: "Dispute API 500 Error Incident"
type: "runbook"
slug: "runbooks/dispute-api-500-error-incident"
freshness: "2026-06-16T17:18:23Z"
tags:
  - "api"
  - "disputes"
  - "error"
  - "incident"
  - "resolved"
owners: []
source_revision_ids:
  - "srcrev_8658552cba5079f2848dace5e1dd800e"
  - "srcrev_d2252e01d6a49d2fa30a836da394304e"
conflict_state: "none"
---

# Dispute API 500 Error Incident

## Summary

An incident where POST /api/v4/disputes returned a 500 Internal Server Error, which was later resolved by blake.huynh@storyprotocol.xyz.

## Claims

- POST /api/v4/disputes failed with 500: Internal server error `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c6122f7d06f0601db14b70c562c40e14` `source_revision_id=srcrev_d2252e01d6a49d2fa30a836da394304e` `chunk_id=srcchunk_45f84e6b238d4d123ae261438e27df49` `native_locator=slack:C07K3J4JTH6:1781411697.987919:1781411697.987919` `source_timestamp=2026-06-14T04:34:57Z`
- Blake Huynh marked the issue as resolved `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c6122f7d06f0601db14b70c562c40e14` `source_revision_id=srcrev_8658552cba5079f2848dace5e1dd800e` `chunk_id=srcchunk_eba17477f29af56792beec303427a659` `native_locator=slack:C07K3J4JTH6:1781411697.987919:1781630303.109329` `source_timestamp=2026-06-16T17:18:23Z`

## Sources

- `source_document_id`: `srcdoc_c6122f7d06f0601db14b70c562c40e14`
- `source_revision_id`: `srcrev_d2252e01d6a49d2fa30a836da394304e`
