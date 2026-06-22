---
title: "Story Protocol Usage Data Monitoring"
type: "concept"
slug: "concepts/usage-data-monitoring"
freshness: "2026-05-15T02:18:19Z"
tags:
  - "data-access"
  - "ip-graph"
  - "monitoring"
  - "poc-modules"
  - "royalty-graph"
owners:
  - "U04L0DD6B6F"
  - "U04VDFP1YQ5"
  - "U05A515NBFC"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_41275761193e96d75e682a02e7e04d16"
  - "srcrev_61fd869e6f05dd0662ec576a85dfd618"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_986a7cfb5f5a9b4f23031b60a8e8e529"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
conflict_state: "none"
---

# Story Protocol Usage Data Monitoring

## Summary

Current state of monitoring and data availability for Royalty Graph v1/v2, IP Graph, and POC Modules (Collection, Licensing, Royalties). As of the investigation, no internal dashboards exist for royalty or POC metrics; IP Graph has a partial Story Indexer dashboard in Grafana. Temporal namespaces provide some workflow data but are not fully accessible by the RSI agent. On-chain data exists for POC modules but requires external tools like Dune Analytics.

## Claims

- Yao requested actual usage data for Royalty graph v1/v2, IP Graph, and POC Modules Collection, Licensing, and Royalties on 2026-05-15. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_61fd869e6f05dd0662ec576a85dfd618` `chunk_id=srcchunk_785dd8f247308c1b94fb2c7eff851031` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808153.324069` `source_timestamp=2026-05-15T01:22:33Z`
- Aiwei Agent suggested Royalty Graph v1/v2 data could be in Temporal Cloud console (royalty-graph-v2 namespace), Grafana SOS dashboards, and PostgreSQL tables in SOS. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- Aiwei Agent suggested IP Graph data likely resides in the indexer Temporal namespace and its DB tables, possibly with Grafana indexer dashboards. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- Aiwei Agent suggested POC Modules data is on-chain and accessible via Dune Analytics or internal blockchain indexing dashboards. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- RSI agent confirmed it lacks read access to SOS and blockchain databases, limiting its ability to query Royalty Graph or IP Graph data directly. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- RSI agent attempted to query the depin-prod database but the read request was denied by a user. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
- RSI agent found that Temporal namespace connections for indexer and royalty graph namespaces were failing with 'connection reset by peer', preventing usage data retrieval. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- The known Temporal namespaces for royalty and IP graphs include indexer-production.koyiy, royalty-graph-production.koyiy, royalty-graph-v2-prod.koyiy, royalty-graph-staging-gcp.koyiy, royalty-graph-v2-staging.koyiy. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- RSI agent's investigation of Grafana discovered a Story Indexer dashboard providing partial IP Graph coverage but found no dashboards for royalty or POC modules. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus does not contain any royalty or IP graph metrics. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- A direct link to the Temporal Cloud royalty-graph-v2-prod.koyiy workflows is available for inspecting workflow-level data. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- Yao clarified that the desired monitoring is an internal dashboard showing royalties and transactions for the royalty graph, not merely Temporal workflow data; no such internal dashboard is known to exist. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_986a7cfb5f5a9b4f23031b60a8e8e529` `chunk_id=srcchunk_fafe4d5c7c4066db94efd23269ed9939` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811433.646089` `source_timestamp=2026-05-15T02:17:13Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`

## Open Questions

- Are SOS Grafana dashboards actually configured for Royalty Graph? They were suggested but not confirmed.
- Is there an internal dashboard for Royalty Graph royalties and transactions? Current investigation suggests none exists.
- What is the full coverage of the Story Indexer dashboard for IP Graph? It is known to be partial.
- Who owns the metrics mentioned by Aiwei Agent (U04L0DD6B6F, U04VDFP1YQ5, U05A515NBFC) and can they provide direct access?

## Related Pages

- `rsi-agent`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7`
