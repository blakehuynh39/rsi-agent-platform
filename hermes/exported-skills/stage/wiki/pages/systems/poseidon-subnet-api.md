---
title: "Poseidon Subnet API"
type: "system"
slug: "systems/poseidon-subnet-api"
freshness: "2025-04-24T22:16:00Z"
tags:
  - "api"
  - "blob-access"
  - "license"
  - "metadata"
  - "subnet"
owners: []
source_revision_ids:
  - "srcrev_12b829e8ee5a792deb125dfa07c2804f"
conflict_state: "none"
---

# Poseidon Subnet API

## Summary

The Poseidon Subnet API provides services for metadata access, blob access, and license management within a subnet.

## Claims

- The Metadata Access service provides search/filtering capabilities (text search, tag filtering, etc.) and fetching metadata for a single dataset or blob_id. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2) `source_document_id=srcdoc_4a714d3273952111fbf0ce4a505e77aa` `source_revision_id=srcrev_12b829e8ee5a792deb125dfa07c2804f` `chunk_id=srcchunk_2838ca68f9b07463377aaf07a745e8bc` `native_locator=https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2` `source_timestamp=2025-04-24T22:16:00Z`
- A S3 compatible service like MinIO allows storing of metadata for each file, and dataset metadata can be stored in a DB used by the Subnet API, which can expose an API to allow dataset owners to update dataset metadata. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2) `source_document_id=srcdoc_4a714d3273952111fbf0ce4a505e77aa` `source_revision_id=srcrev_12b829e8ee5a792deb125dfa07c2804f` `chunk_id=srcchunk_2838ca68f9b07463377aaf07a745e8bc` `native_locator=https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2` `source_timestamp=2025-04-24T22:16:00Z`
- The Blob Access service processes data upload and retrieval requests, routes storage to the appropriate node in the subnet for upload, and for data retrieval, validates the license key and returns the correct s3-compatible endpoint to retrieve the blob. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2) `source_document_id=srcdoc_4a714d3273952111fbf0ce4a505e77aa` `source_revision_id=srcrev_12b829e8ee5a792deb125dfa07c2804f` `chunk_id=srcchunk_2838ca68f9b07463377aaf07a745e8bc` `native_locator=https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2` `source_timestamp=2025-04-24T22:16:00Z`
- The License service generates read and upload licenses to blobs, validates licenses for reading and uploading blobs, and implements caching (e.g., Redis, Memcached) for license checks to minimize latency and blockchain load. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2) `source_document_id=srcdoc_4a714d3273952111fbf0ce4a505e77aa` `source_revision_id=srcrev_12b829e8ee5a792deb125dfa07c2804f` `chunk_id=srcchunk_2838ca68f9b07463377aaf07a745e8bc` `native_locator=https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2` `source_timestamp=2025-04-24T22:16:00Z`

## Related Pages

- `ai-data-service`
- `poseidon-sdk`

## Sources

- `source_document_id`: `srcdoc_4a714d3273952111fbf0ce4a505e77aa`
- `source_revision_id`: `srcrev_12b829e8ee5a792deb125dfa07c2804f`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba)
