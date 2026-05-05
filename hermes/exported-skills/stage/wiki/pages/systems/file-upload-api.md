---
title: "File Upload API"
type: "system"
slug: "systems/file-upload-api"
freshness: "2026-05-05T06:27:20Z"
tags:
  - "api"
  - "storyprotocol"
  - "upload"
owners: []
source_revision_ids:
  - "srcrev_30ad49a78626b6f998abe93ca76dde4a"
conflict_state: "none"
---

# File Upload API

## Summary

Specification for the Story Protocol file upload endpoint, including request format, parameters, and response.

## Claims

- The file upload endpoint is a POST request to https://stag.api.storyprotocol.net/protocol/v2/files/upload. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b#chunk-1) `source_document_id=srcdoc_2321dc61cde7bde2e81b8c2713e8c1df` `source_revision_id=srcrev_30ad49a78626b6f998abe93ca76dde4a` `chunk_id=srcchunk_e19f5673de8787cef69891db146206e4` `native_locator=https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b#chunk-1` `source_timestamp=2026-05-05T06:27:20Z`
- The request body must contain a base64-encoded file string and a contentType string (e.g., 'image/png'). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b#chunk-1) `source_document_id=srcdoc_2321dc61cde7bde2e81b8c2713e8c1df` `source_revision_id=srcrev_30ad49a78626b6f998abe93ca76dde4a` `chunk_id=srcchunk_e19f5673de8787cef69891db146206e4` `native_locator=https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b#chunk-1` `source_timestamp=2026-05-05T06:27:20Z`
- The request body may optionally include a metadata field (string, markdown or JSON containing markdown) and an owner field. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b#chunk-2) `source_document_id=srcdoc_2321dc61cde7bde2e81b8c2713e8c1df` `source_revision_id=srcrev_30ad49a78626b6f998abe93ca76dde4a` `chunk_id=srcchunk_276aca78283aca099309fbb6650cc9f2` `native_locator=https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b#chunk-2` `source_timestamp=2026-05-05T06:27:20Z`
- The response is a JSON object containing a txHash string. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b#chunk-2) `source_document_id=srcdoc_2321dc61cde7bde2e81b8c2713e8c1df` `source_revision_id=srcrev_30ad49a78626b6f998abe93ca76dde4a` `chunk_id=srcchunk_276aca78283aca099309fbb6650cc9f2` `native_locator=https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b#chunk-2` `source_timestamp=2026-05-05T06:27:20Z`

## Sources

- `source_document_id`: `srcdoc_2321dc61cde7bde2e81b8c2713e8c1df`
- `source_revision_id`: `srcrev_30ad49a78626b6f998abe93ca76dde4a`
- `source_url`: [Notion source](https://www.notion.so/Uploader-Request-Response-example-8ad9cc98b2f14d11bf29c633584efd0b)
