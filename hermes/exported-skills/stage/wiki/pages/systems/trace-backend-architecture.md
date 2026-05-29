---
title: "Trace Backend Architecture"
type: "system"
slug: "systems/trace-backend-architecture"
freshness: "2026-05-29T14:22:00Z"
tags: []
owners: []
source_revision_ids:
  - "srcrev_51ed2f05c5df01838126408740121818"
conflict_state: "none"
---

# Trace Backend Architecture

## Summary

Overview of the Trace backend system architecture, including staging deployment, write/read paths, SQS-based ingestion, and current limitations.

## Claims

- Staging V1 of Trace backend is deployed and load-tested, last updated May 21, 2026. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_51ed2f05c5df01838126408740121818` `chunk_id=srcchunk_26eaf12c4cdb5ba9fb4b3d3c63cac37b` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-29T14:22:00Z`
- Story owns the global data_id UUID and normalized Trace Schema v1.0. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_51ed2f05c5df01838126408740121818` `chunk_id=srcchunk_26eaf12c4cdb5ba9fb4b3d3c63cac37b` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-29T14:22:00Z`
- The write path flows: provider clients → Cloudflare (staging-api.storyprotocol.net) → story-api webhook → SQS Standard → data-audit-ingestor → durable data audit store with explicit index rows. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_51ed2f05c5df01838126408740121818` `chunk_id=srcchunk_26eaf12c4cdb5ba9fb4b3d3c63cac37b` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-29T14:22:00Z`
- SQS is at-least-once; duplicates are handled idempotently, and messages are deleted only after successful persistence. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_51ed2f05c5df01838126408740121818` `chunk_id=srcchunk_26eaf12c4cdb5ba9fb4b3d3c63cac37b` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-29T14:22:00Z`
- The staging queue has a DLQ for conflict and poison messages, but no automatic replay is implemented. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_51ed2f05c5df01838126408740121818` `chunk_id=srcchunk_26eaf12c4cdb5ba9fb4b3d3c63cac37b` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-29T14:22:00Z`
- Kled should retry transient 502, 503, 504, 429, and network failures with the same request body and X-Batch-Id. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_51ed2f05c5df01838126408740121818` `chunk_id=srcchunk_26eaf12c4cdb5ba9fb4b3d3c63cac37b` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-29T14:22:00Z`
- 202 Accepted response means story-api validated the request and SQS accepted the message, but persistence is asynchronous. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1) `source_document_id=srcdoc_f33f716b82984e27937f90590ba0afd6` `source_revision_id=srcrev_51ed2f05c5df01838126408740121818` `chunk_id=srcchunk_26eaf12c4cdb5ba9fb4b3d3c63cac37b` `native_locator=https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914#chunk-1` `source_timestamp=2026-05-29T14:22:00Z`

## Sources

- `source_document_id`: `srcdoc_f33f716b82984e27937f90590ba0afd6`
- `source_revision_id`: `srcrev_51ed2f05c5df01838126408740121818`
- `source_url`: [Notion source](https://www.notion.so/Trace-Backend-Architecture-35e051299a5480a3864be5b963962914)
