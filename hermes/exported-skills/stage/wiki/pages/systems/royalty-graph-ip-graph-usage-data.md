---
title: "Royalty Graph and IP Graph Usage Data Sources"
type: "system"
slug: "systems/royalty-graph-ip-graph-usage-data"
freshness: "2026-05-15T02:18:19Z"
tags:
  - "data-access"
  - "database"
  - "grafana"
  - "monitoring"
  - "temporal"
  - "usage-data"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_41275761193e96d75e682a02e7e04d16"
  - "srcrev_61fd869e6f05dd0662ec576a85dfd618"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_986a7cfb5f5a9b4f23031b60a8e8e529"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
  - "srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7"
conflict_state: "none"
---

# Royalty Graph and IP Graph Usage Data Sources

## Summary

Documents the available and unavailable sources for usage data of Royalty Graph v1/v2, IP Graph, and POC Modules (Collection, Licensing, Royalties). No internal dashboards were found for royalty or POC modules; Grafana provides a Story Indexer dashboard for partial IP Graph coverage. Database access to SOS and blockchain is denied. Temporal namespace connections are currently failing. The path to usage data remains unclear.

## Claims

- A request was made for actual usage data of Royalty Graph v1/v2, IP Graph, and POC Modules (Collection, Licensing, Royalties). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_61fd869e6f05dd0662ec576a85dfd618` `chunk_id=srcchunk_785dd8f247308c1b94fb2c7eff851031` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808153.324069` `source_timestamp=2026-05-15T01:22:33Z`
- Aiwei bot suggested data sources: for Royalty Graph, Temporal Cloud `royalty-graph-v2` namespace, Grafana SOS dashboards, PostgreSQL SOS tables; for IP Graph, indexer Temporal namespace and DB tables; for POC Modules, on-chain metrics via Dune Analytics or internal blockchain indexing dashboards. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- Aiwei does not have direct access to usage dashboards or live metrics. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- RSI initially lacked read access to SOS or blockchain databases. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- Database read requests to the `depin-prod` database were denied by user U0772SH7BRA. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7` `chunk_id=srcchunk_045ff27753cd8b236824f84388aa8905` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810034.117999` `source_timestamp=2026-05-15T01:54:06Z`
- RSI attempted to check Temporal namespaces for usage data but all namespace connections failed with 'connection reset by peer'. Namespaces checked: `indexer-production.koyiy`, `royalty-graph-production.koyiy`, `royalty-graph-v2-prod.koyiy`, `royalty-graph-staging-gcp.koyiy`, `royalty-graph-v2-staging.koyiy`. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- Grafana has a Story Indexer dashboard providing partial IP Graph coverage, but lacks dashboards for royalty or POC modules. Prometheus also has no royalty/IP graph metrics. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI's database access is limited to `depin`; SOS and blockchain DBs remain unavailable. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- There is confusion about whether an internal dashboard for royalty graph usage (royalties + transactions) exists, distinct from Temporal. Some team members were unaware of such a dashboard. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_986a7cfb5f5a9b4f23031b60a8e8e529` `chunk_id=srcchunk_fafe4d5c7c4066db94efd23269ed9939` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811433.646089` `source_timestamp=2026-05-15T02:17:13Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`

## Open Questions

- How can database read access be obtained for SOS and blockchain databases?
- What internal dashboards (if any) exist for royalty graph and POC module usage data?
- Who owns the metrics and data pipelines for royalty graph, IP graph, and POC modules?
- Why are Temporal namespace connections failing, and when will they be fixed?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_c5f44a625d9cfb3a708b34791880757c`
