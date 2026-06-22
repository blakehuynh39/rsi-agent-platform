---
title: "Usage Data Inventory for Royalty Graph, IP Graph, and POC Modules"
type: "system"
slug: "systems/usage-data-inventory-royalty-ip-graph-poc"
freshness: "2026-05-15T02:18:19Z"
tags:
  - "data-access"
  - "ip-graph"
  - "poc-modules"
  - "royalty-graph"
owners:
  - "Aiwei Agent (U0ANHPXBBDM)"
  - "RSI (U0ASDQKU3UL)"
  - "Yao (U083MMT1771)"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_41275761193e96d75e682a02e7e04d16"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_986a7cfb5f5a9b4f23031b60a8e8e529"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7"
conflict_state: "none"
---

# Usage Data Inventory for Royalty Graph, IP Graph, and POC Modules

## Summary

Assessment of data sources available and access status for usage metrics of Royalty Graph, IP Graph, and POC Modules, based on a conversation on 2026-05-15.

## Claims

- Royalty Graph v2 production namespace is `royalty-graph-v2-prod.koyiy`, accessible via Temporal Cloud console. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- Aiwei Agent pointed to Temporal Cloud console (royalty-graph-v2), Grafana SOS dashboards, PostgreSQL tables for Royalty Graph; indexer Temporal namespace and DB tables for IP Graph; on-chain data via Dune Analytics or internal indexing for POC Modules. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- RSI investigation found Grafana has a Story Indexer dashboard providing partial IP Graph coverage, but no royalty or POC dashboards exist. DB access is limited to depin; SOS and blockchain databases are unavailable. Prometheus has no royalty/IP graph metrics. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Two DB read requests against the `depin-prod` target were denied by user U0772SH7BRA, blocking read access to potential data sources. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7` `chunk_id=srcchunk_045ff27753cd8b236824f84388aa8905` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810034.117999` `source_timestamp=2026-05-15T01:54:06Z`
- RSI attempted to query Temporal namespaces directly but all connections failed with "connection reset by peer". Namespaces checked: `indexer-production.koyiy`, `royalty-graph-production.koyiy`, `royalty-graph-v2-prod.koyiy`, `royalty-graph-staging-gcp.koyiy`, `royalty-graph-v2-staging.koyiy`. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- No dedicated internal dashboard for Royalty Graph royalties and transactions was identified by the thread participants. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_986a7cfb5f5a9b4f23031b60a8e8e529` `chunk_id=srcchunk_fafe4d5c7c4066db94efd23269ed9939` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811433.646089` `source_timestamp=2026-05-15T02:17:13Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`

## Open Questions

- Are there existing Grafana dashboards for IP Graph beyond the Story Indexer partial coverage?
- Who internally owns dashboards or metrics for Royalty Graph and POC Modules?
- Why does the Temporal connection fail for the agent while the console is accessible?
- Why is DB read access to SOS not granted?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_986a7cfb5f5a9b4f23031b60a8e8e529`
