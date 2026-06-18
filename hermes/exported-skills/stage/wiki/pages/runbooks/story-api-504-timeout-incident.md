---
title: "Story API 504 Timeout Incident"
type: "runbook"
slug: "runbooks/story-api-504-timeout-incident"
freshness: "2026-02-28T16:00:43Z"
tags:
  - "incident"
  - "story-api"
  - "timeout"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_5c60cbca531a8f0d37eca4489bc39e42"
  - "srcrev_f9f94b28a1e097a6eb72a4f763f890ae"
conflict_state: "none"
---

# Story API 504 Timeout Incident

## Summary

POST /api/v4/transactions experienced a 504 timeout error, later resolved.

## Claims

- POST /api/v4/transactions failed with 504: Request timeout `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d6480023c8ae0a48bff6479d40b22721` `source_revision_id=srcrev_5c60cbca531a8f0d37eca4489bc39e42` `chunk_id=srcchunk_b877675fcd133b9083e46e16f5937211` `native_locator=slack:C07K3J4JTH6:1772236643.997649:1772236643.997649` `source_timestamp=2026-02-27T23:57:23Z`
- Issue STORY-API-E8 was marked as resolved in Sentry by blake.huynh@storyprotocol.xyz `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d6480023c8ae0a48bff6479d40b22721` `source_revision_id=srcrev_f9f94b28a1e097a6eb72a4f763f890ae` `chunk_id=srcchunk_1f4c55846a41b6cf62c67da7630ea635` `native_locator=slack:C07K3J4JTH6:1772236643.997649:1772294443.029829` `source_timestamp=2026-02-28T16:00:43Z`

## Sources

- `source_document_id`: `srcdoc_d6480023c8ae0a48bff6479d40b22721`
- `source_revision_id`: `srcrev_f9f94b28a1e097a6eb72a4f763f890ae`
