---
title: "Data Audit Store 503 Error"
type: "runbook"
slug: "runbooks/data-audit-store-not-configured-503-error"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "api"
  - "data-audit"
  - "incident"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_6763606d53f08bbf001f2cbf10b8326b"
  - "srcrev_fa23e5e820aa060162f9cd27a5ca668a"
conflict_state: "none"
---

# Data Audit Store 503 Error

## Summary

The /api/v1/data-audit/stats endpoint returns a 503 error when the data audit store is not configured.

## Claims

- GET /api/v1/data-audit/stats failed with 503 error: data audit store is not configured. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b89783489f3a57ee563b37deb359e6a0` `source_revision_id=srcrev_6763606d53f08bbf001f2cbf10b8326b` `chunk_id=srcchunk_b78a9e50abf036569b06e04f9ab27686` `native_locator=slack:C07K3J4JTH6:1781138983.948289` `source_timestamp=2026-06-11T00:49:43Z`
- Blake Huynh marked the Sentry issue STORY-API-EW as resolved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b89783489f3a57ee563b37deb359e6a0` `source_revision_id=srcrev_fa23e5e820aa060162f9cd27a5ca668a` `chunk_id=srcchunk_4bbdc88235971ff2ccb75fc65cbe2bd2` `native_locator=slack:C07K3J4JTH6:1781630302.988559` `source_timestamp=2026-06-16T17:18:22Z`

## Open Questions

- Why was the data audit store not configured?

## Sources

- `source_document_id`: `srcdoc_b89783489f3a57ee563b37deb359e6a0`
- `source_revision_id`: `srcrev_fa23e5e820aa060162f9cd27a5ca668a`
