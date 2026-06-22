---
title: "Royalty Graph and IP Graph Usage Monitoring"
type: "system"
slug: "systems/royalty-graph-ip-graph-usage-monitoring"
freshness: "2026-05-15T02:18:19Z"
tags:
  - "grafana"
  - "ip-graph"
  - "monitoring"
  - "royalty-graph"
  - "rsi"
  - "temporal"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_41275761193e96d75e682a02e7e04d16"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
conflict_state: "none"
---

# Royalty Graph and IP Graph Usage Monitoring

## Summary

No internal dashboard exists for royalty transactions or POC module usage. Royalty Graph v2 has a Temporal namespace with workflow monitoring. Grafana's Story Indexer provides partial IP Graph coverage. RSI cannot access SOS or blockchain databases and Temporal connections are currently failing.

## Claims

- Royalty Graph v2 has a Temporal namespace `royalty-graph-v2-prod.koyiy` with workflow monitoring at https://cloud.temporal.io/namespaces/royalty-graph-v2-prod.koyiy/workflows. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- No internal dashboard exists for royalty graph transaction and royalty amounts; only Temporal workflow monitoring is currently observable. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`
- Grafana hosts a Story Indexer dashboard providing partial IP Graph coverage, but no dashboards exist for royalty or POC modules. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI's database read access is limited to `depin`; SOS and blockchain databases are not accessible. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus does not contain any royalty or IP graph metrics. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI attempted to connect to Temporal namespaces (including `royalty-graph-v2-prod.koyiy`) to retrieve usage data but all connections failed with 'connection reset by peer', preventing direct retrieval. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`

## Open Questions

- Are there any metrics or logging for POC Modules Collection, Licensing, and Royalties beyond on-chain data?
- Is there an internal dashboard for royalty graph transaction amounts or is it planned?
- What is the current coverage of the Story Indexer dashboard for IP Graph and is it sufficient?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_4a6d275eac509decd3f3e11fbbb44b94`
