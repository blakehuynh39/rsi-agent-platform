---
title: "Royalty Graph and IP Graph Usage Data Availability"
type: "open_question"
slug: "open-questions/royalty-graph-usage-data-availability"
freshness: "2026-05-15T02:18:19Z"
tags:
  - "dashboards"
  - "ip-graph"
  - "monitoring"
  - "royalty-graph"
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
conflict_state: "none"
---

# Royalty Graph and IP Graph Usage Data Availability

## Summary

Current state of usage data and dashboards for Royalty Graph, IP Graph, and POC Modules as of 2026-05-15. No dedicated internal dashboards exist; partial IP Graph coverage via Story Indexer; Temporal namespace access broken; DB access denied; Prometheus lacks metrics. The existence of an internal royalty graph dashboard is unknown.

## Claims

- Yao requested actual usage data for Royalty Graph v1/v2, current usage of IP Graph, and POC Modules (Collection, Licensing, Royalties). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_61fd869e6f05dd0662ec576a85dfd618` `chunk_id=srcchunk_785dd8f247308c1b94fb2c7eff851031` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808153.324069` `source_timestamp=2026-05-15T01:22:33Z`
- Aiwei Agent suggested looking at Temporal Cloud namespace royalty-graph-v2, Grafana SOS dashboards, PostgreSQL tables, and on-chain data for POC modules. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- RSI investigation found that Grafana has the Story Indexer dashboard providing partial IP Graph coverage, but no dashboards for Royalty or POC modules. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- DB access to SOS and blockchain databases is unavailable; a read request to depin-prod was denied. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Temporal namespace access is broken with 'connection reset by peer' errors; namespaces for royalty graph and indexer exist but cannot be queried. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- Prometheus did not have any metrics for royalty or IP graph. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- It is unclear whether an internal dashboard for royalty graph usage data exists; Yao asked for one beyond Temporal, and the question remains open. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_986a7cfb5f5a9b4f23031b60a8e8e529` `chunk_id=srcchunk_fafe4d5c7c4066db94efd23269ed9939` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811433.646089` `source_timestamp=2026-05-15T02:17:13Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`

## Open Questions

- How can technical staff gain read access to SOS and blockchain databases?
- Is there an internal dashboard for Royalty Graph usage data and transactions?
- What is the status of Temporal namespace connectivity and when will it be restored?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_98bf163135665bd9846a051a069385a5`
