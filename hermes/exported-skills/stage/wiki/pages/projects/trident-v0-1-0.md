---
title: "Trident v0.1.0"
type: "project"
slug: "projects/trident-v0-1-0"
freshness: "2025-08-11T18:37:00Z"
tags:
  - "release"
  - "v0.1.0"
owners:
  - "user://5d2965c9-40b2-44ff-9f1f-ae930770080e"
source_revision_ids:
  - "srcrev_0cdbd21278ee52008571abf6a4ca3588"
  - "srcrev_509996f6ec9f5d494184aa0988321c4c"
conflict_state: "none"
---

# Trident v0.1.0

## Summary

Trident v0.1.0 release details including dependencies, devnet configuration, and platform APIs.

## Claims

- Trident v0.1.0 is a release of the Trident project. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trident-Releases-247051299a5480aa9e08edd82c95562e) `source_document_id=srcdoc_1c0bc80061d6f5a5121b397c2b844560` `source_revision_id=srcrev_509996f6ec9f5d494184aa0988321c4c` `chunk_id=srcchunk_a2de2be682c41668c6bd14fae84ed84c` `native_locator=https://www.notion.so/Trident-Releases-247051299a5480aa9e08edd82c95562e` `source_timestamp=2025-08-11T18:37:00Z`
- Trident v0.1.0 depends on poseidon v0.1.0 (commit db80ad2666950ca8ea8d4103ab71e0f6c322c180) for storage node/KMS, poseidon-app alpha v0.1.0 for frontend, poseidon-benchmarks v0.1.0 for benchmark tool, poseidon-contracts v0.1.0 (commit 5f6fcffec29f0eb18305a0201e4e3368bfdee822) for smart contracts, poseidon-devnet v0.1.0 (commit 7b8f4e3fca20a44b7137026fdad735c3e7508059) for devnet, and trident-platform v0.1.0 (commit bfe2de24aa4b9abab08547e00875b9515f2cb4e6) for backend APIs. `claim:claim_deps` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-1) `source_document_id=srcdoc_61a92d64d3415fa33afd84da190289da` `source_revision_id=srcrev_0cdbd21278ee52008571abf6a4ca3588` `chunk_id=srcchunk_72c38ecc780e9394d1d01689a43621eb` `native_locator=https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-1` `source_timestamp=2025-08-06T00:25:00Z`
- The Trident MVP website is accessible at https://poseidon-app.vercel.app with password 'poseidonCrunch'. `claim:claim_website` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-1) `source_document_id=srcdoc_61a92d64d3415fa33afd84da190289da` `source_revision_id=srcrev_0cdbd21278ee52008571abf6a4ca3588` `chunk_id=srcchunk_72c38ecc780e9394d1d01689a43621eb` `native_locator=https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-1` `source_timestamp=2025-08-06T00:25:00Z`
- The devnet for Trident v0.1.0 has Chain ID 1518, RPC endpoint https://rpc.poseidon.storyrpc.io, and explorer https://poseidon.storyscan.io. `claim:claim_devnet` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-1) `source_document_id=srcdoc_61a92d64d3415fa33afd84da190289da` `source_revision_id=srcrev_0cdbd21278ee52008571abf6a4ca3588` `chunk_id=srcchunk_72c38ecc780e9394d1d01689a43621eb` `native_locator=https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-1` `source_timestamp=2025-08-06T00:25:00Z`
- Trident v0.1.0 platform APIs include Bucket operations (CreateBucket, ListBuckets, HeadBucket, DeleteBuckets) and Object operations (ListObjectsV2, HeadObject, GetObject). `claim:claim_platform_apis` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-1) `source_document_id=srcdoc_61a92d64d3415fa33afd84da190289da` `source_revision_id=srcrev_0cdbd21278ee52008571abf6a4ca3588` `chunk_id=srcchunk_72c38ecc780e9394d1d01689a43621eb` `native_locator=https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-1` `source_timestamp=2025-08-06T00:25:00Z`
- Object endpoints: GET /:object_id retrieves object stream; POST / stores new object (multipart/form-data with object_id and file); DELETE /:object_id deletes object. Detailed error responses documented. `claim:claim_object_endpoints` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-2) `source_document_id=srcdoc_61a92d64d3415fa33afd84da190289da` `source_revision_id=srcrev_0cdbd21278ee52008571abf6a4ca3588` `chunk_id=srcchunk_099a4db6744ff706d0a979a9ed58d293` `native_locator=https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-2` `source_timestamp=2025-08-06T00:25:00Z`
- KMS endpoint GET /kms/:bucket_id returns SSS decryption key for the bucket. Error responses for missing bucket_id or not found. `claim:claim_kms_endpoint` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-2) `source_document_id=srcdoc_61a92d64d3415fa33afd84da190289da` `source_revision_id=srcrev_0cdbd21278ee52008571abf6a4ca3588` `chunk_id=srcchunk_099a4db6744ff706d0a979a9ed58d293` `native_locator=https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246#chunk-2` `source_timestamp=2025-08-06T00:25:00Z`

## Sources

- `source_document_id`: `srcdoc_61a92d64d3415fa33afd84da190289da`
- `source_revision_id`: `srcrev_0cdbd21278ee52008571abf6a4ca3588`
- `source_url`: [Notion source](https://www.notion.so/Trident-v0-1-0-1f9051299a5480fa8fd6c4242fa48246)
