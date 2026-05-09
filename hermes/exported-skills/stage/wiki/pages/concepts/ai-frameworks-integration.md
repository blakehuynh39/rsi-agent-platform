---
title: "AI Frameworks Integration"
type: "concept"
slug: "concepts/ai-frameworks-integration"
freshness: "2025-04-24T22:16:00Z"
tags:
  - "ai"
  - "frameworks"
  - "s3"
  - "streaming"
owners: []
source_revision_ids:
  - "srcrev_12b829e8ee5a792deb125dfa07c2804f"
conflict_state: "none"
---

# AI Frameworks Integration

## Summary

Guidance on integrating Poseidon's S3 endpoint with popular AI frameworks like PyTorch, TensorFlow, Hugging Face Datasets, and Ray Datasets.

## Claims

- For PyTorch with WebDataset, pass Poseidon endpoint_url and credentials to WebDataset, using a direct URL like f"{url}/{bucket}/{blob_id}". `claim:claim_4_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2) `source_document_id=srcdoc_4a714d3273952111fbf0ce4a505e77aa` `source_revision_id=srcrev_12b829e8ee5a792deb125dfa07c2804f` `chunk_id=srcchunk_2838ca68f9b07463377aaf07a745e8bc` `native_locator=https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2` `source_timestamp=2025-04-24T22:16:00Z`
- For TensorFlow with smart_open and tf.data.Dataset, wrap smart_open.open() with Poseidon’s boto3.client using custom transport_params and feed the stream to a generator. `claim:claim_4_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2) `source_document_id=srcdoc_4a714d3273952111fbf0ce4a505e77aa` `source_revision_id=srcrev_12b829e8ee5a792deb125dfa07c2804f` `chunk_id=srcchunk_2838ca68f9b07463377aaf07a745e8bc` `native_locator=https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2` `source_timestamp=2025-04-24T22:16:00Z`
- For Hugging Face Datasets with load_dataset(..., streaming=True), use s3://{bucket}/{blob_id} and set hf_s3_endpoint_url and AWS credentials from Poseidon via env vars or config. `claim:claim_4_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2) `source_document_id=srcdoc_4a714d3273952111fbf0ce4a505e77aa` `source_revision_id=srcrev_12b829e8ee5a792deb125dfa07c2804f` `chunk_id=srcchunk_2838ca68f9b07463377aaf07a745e8bc` `native_locator=https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2` `source_timestamp=2025-04-24T22:16:00Z`
- For Ray Datasets, initialize with ray.init(_system_config={"s3.endpoint_override": data.url}) and use the standard S3 URI. `claim:claim_4_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2) `source_document_id=srcdoc_4a714d3273952111fbf0ce4a505e77aa` `source_revision_id=srcrev_12b829e8ee5a792deb125dfa07c2804f` `chunk_id=srcchunk_2838ca68f9b07463377aaf07a745e8bc` `native_locator=https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba#chunk-2` `source_timestamp=2025-04-24T22:16:00Z`

## Related Pages

- `poseidon-sdk`

## Sources

- `source_document_id`: `srcdoc_4a714d3273952111fbf0ce4a505e77aa`
- `source_revision_id`: `srcrev_12b829e8ee5a792deb125dfa07c2804f`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-SDK-1dc051299a5480959b9dfc50604e89ba)
