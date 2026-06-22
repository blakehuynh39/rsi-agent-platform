---
title: "Usage Data Sources for Royalty Graph, IP Graph, and POC Modules"
type: "concept"
slug: "concepts/royalty-graph-ip-graph-poc-usage-data-sources"
freshness: "2026-05-15T02:16:39Z"
tags:
  - "data-sources"
  - "ip-graph"
  - "monitoring"
  - "poc"
  - "royalty-graph"
  - "usage"
owners:
  - "@U04L0DD6B6F"
  - "@U04VDFP1YQ5"
  - "@U0772SH7BRA"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
  - "srcrev_f5bcdcbd7c4349bf12a16f9b355cc860"
conflict_state: "none"
---

# Usage Data Sources for Royalty Graph, IP Graph, and POC Modules

## Summary

Information gathered about available data sources for Royalty Graph v1/v2, IP Graph, and POC Modules usage data. As of 2026-05-15, no dedicated dashboards or database access exist for royalty and POC; only a partial IP Graph dashboard exists via Grafana Story Indexer. Temporal namespaces are identified but currently inaccessible due to connection issues.

## Claims

- The AI agent does not have direct access to usage dashboards or live metrics for Royalty Graph, IP Graph, or POC modules. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- Grafana dashboard 'Story Indexer' provides partial IP Graph coverage. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- No Grafana dashboards exist for Royalty Graph or POC modules. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus has no royalty/IP graph metrics. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Database access is limited to depin; SOS and blockchain DBs are unavailable. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- Identified Temporal namespaces for royalty graph and IP indexer: indexer-production.koyiy, royalty-graph-production.koyiy, royalty-graph-v2-prod.koyiy, royalty-graph-staging-gcp.koyiy, royalty-graph-v2-staging.koyiy. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- Connection to Temporal namespaces failed with 'connection reset by peer', so usage data could not be retrieved. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- Production Temporal Cloud URL for royalty-graph-v2 workflows is https://cloud.temporal.io/namespaces/royalty-graph-v2-prod.koyiy/workflows. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- The AI agent (Aiwei Agent) is a bot/app user, and Slack may block normal DMs, so asking in thread is preferred. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_f5bcdcbd7c4349bf12a16f9b355cc860` `chunk_id=srcchunk_847642dd6f713e95f89ac10f7d58f043` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808230.793089` `source_timestamp=2026-05-15T01:23:52Z`

## Open Questions

- Is there an internal dashboard for Royalty Graph transactions and royalties beyond Temporal?
- Who owns the metrics and dashboards for Royalty Graph and POC modules?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa`
