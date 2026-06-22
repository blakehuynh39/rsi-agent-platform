---
title: "Royalty Graph v1/v2"
type: "system"
slug: "systems/royalty-graph-v1-v2"
freshness: "2026-05-15T02:18:26Z"
tags:
  - "monitoring"
  - "royalty-graph"
  - "temporal"
  - "usage"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_41275761193e96d75e682a02e7e04d16"
  - "srcrev_53c073c35d14caf2817be96bf175fc8b"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
conflict_state: "none"
---

# Royalty Graph v1/v2

## Summary

Royalty Graph component handles royalty calculations and related workflows. It is deployed across multiple Temporal namespaces (v1 and v2), but currently lacks an internal dashboard for transaction/royalty data.

## Claims

- Royalty Graph v2 is accessible through Temporal Cloud namespace royalty-graph-v2-prod.koyiy. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- There is no internal dashboard for royalty graph transaction amounts or royalty data. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_53c073c35d14caf2817be96bf175fc8b` `chunk_id=srcchunk_b5017699187e6434648cc6c59d4abef9` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811506.588819` `source_timestamp=2026-05-15T02:18:26Z`
- Temporal namespaces related to Royalty Graph include: royalty-graph-production.koyiy, royalty-graph-v2-prod.koyiy, royalty-graph-staging-gcp.koyiy, royalty-graph-v2-staging.koyiy. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- RSI investigation on 2026-05-15 found no Grafana dashboard, no DB access, and no Prometheus metrics specifically for royalty graph monitoring. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`

## Open Questions

- How to obtain access to SOS database for royalty graph job state?
- Is there a Dune Analytics dashboard for Royalty Graph on-chain transactions?
- Who is responsible for Royalty Graph monitoring?

## Related Pages

- `ip-graph`
- `poc-modules-collection-licensing-royalties`
- `rsi-investigation-2026-05-15-usage-data`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_c3ccf444cfe45d7ef6bb2a9d5d8c06d3`
