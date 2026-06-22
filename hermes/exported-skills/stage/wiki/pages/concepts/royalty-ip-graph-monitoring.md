---
title: "Royalty and IP Graph Monitoring"
type: "concept"
slug: "concepts/royalty-ip-graph-monitoring"
freshness: "2026-05-15T02:17:13Z"
tags:
  - "ip-graph"
  - "monitoring"
  - "poc-modules"
  - "royalty-graph"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_986a7cfb5f5a9b4f23031b60a8e8e529"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
conflict_state: "none"
---

# Royalty and IP Graph Monitoring

## Summary

Current state of monitoring and data access for Royalty Graph v1/v2, IP Graph, and POC Modules (Collection, Licensing, Royalties) as of 2026-05-15.

## Claims

- There is no internal Grafana dashboard for Royalty Graph v1/v2 or POC Modules. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_986a7cfb5f5a9b4f23031b60a8e8e529` `chunk_id=srcchunk_fafe4d5c7c4066db94efd23269ed9939` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811433.646089` `source_timestamp=2026-05-15T02:17:13Z`
- Royalty Graph workflow execution data is accessible via Temporal Cloud at the 'royalty-graph-v2-prod.koyiy' namespace. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- RSI's connection to Temporal namespaces (royalty-graph, indexer) is failing with 'connection reset by peer' errors as of 2026-05-15, preventing usage data retrieval. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- RSI database read access is limited to 'depin' target; attempts to query SOS or blockchain databases were denied. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
- No Prometheus metrics are available for royalty or IP graph data. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Partial IP Graph monitoring is available via the Grafana Story Indexer dashboard. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`

## Open Questions

- How can RSI's Temporal namespace connection be restored?
- What are the exact metrics emitted by SOS for royalty graph?
- What is the link to the Grafana Story Indexer dashboard?
- Where can internal dashboards for royalty and transaction counts be found?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_f5bcdcbd7c4349bf12a16f9b355cc860`
