---
title: "Usage Data for Royalty Graph, IP Graph, and POC Modules"
type: "concept"
slug: "concepts/royalty-ip-poc-usage-data"
freshness: "2026-05-15T02:16:39Z"
tags:
  - "ip-graph"
  - "monitoring"
  - "observability"
  - "poc-modules"
  - "royalty-graph"
owners:
  - "U04L0DD6B6F"
  - "U04VDFP1YQ5"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
  - "srcrev_f9a26ebbd3baf467fef114060a2f2d5a"
conflict_state: "none"
---

# Usage Data for Royalty Graph, IP Graph, and POC Modules

## Summary

Current state of available dashboards, metrics, and data access for Royalty Graph, IP Graph, and POC Modules.

## Claims

- No internal dashboards exist for Royalty Graph transaction/royalty volume or POC Modules usage data. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Grafana has a Story Indexer dashboard that provides partial coverage of IP Graph metrics. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus does not currently collect royalty or IP Graph metrics. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI's database access is limited to the 'depin' database; SOS and blockchain DBs are unavailable. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- Temporal namespaces exist for royalty graph and indexer production environments, but connectivity from RSI is currently failing with 'connection reset by peer'. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- The royalty-graph-v2-prod Temporal namespace is accessible via https://cloud.temporal.io/namespaces/royalty-graph-v2-prod.koyiy/workflows, but this provides workflow execution data, not business metrics like royalty amounts. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- SOS repository has read access for some users, and may contain emitted metrics or queryable data that could help answer usage questions. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_f9a26ebbd3baf467fef114060a2f2d5a` `chunk_id=srcchunk_f24316a2c55750f8f77f2da2f3fa639d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810137.026989` `source_timestamp=2026-05-15T01:55:39Z`

## Open Questions

- How can we obtain actual usage data for Royalty Graph (v1/v2), IP Graph, and POC Modules?
- Who should build or provide dashboards/metrics for these components?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_71094780a5c936f2948da743f0c75821`
