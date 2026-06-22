---
title: "Observability for Royalty Graph, IP Graph, and POC Modules"
type: "system"
slug: "systems/observability-royalty-graph-ip-graph-poc"
freshness: "2026-05-15T02:18:26Z"
tags:
  - "ip-graph"
  - "monitoring"
  - "observability"
  - "poc-modules"
  - "royalty-graph"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_41275761193e96d75e682a02e7e04d16"
  - "srcrev_53c073c35d14caf2817be96bf175fc8b"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
  - "srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7"
conflict_state: "none"
---

# Observability for Royalty Graph, IP Graph, and POC Modules

## Summary

Current state of usage data sources for Royalty Graph v1/v2, IP Graph, and POC Modules (Collection, Licensing, Royalties) as of 2026-05-15. RSI lacks access to most data sources; only partial IP coverage via Grafana Story Indexer dashboard; Temporal namespaces exist but unreachable; no dedicated royalty or POC dashboards; no Prometheus metrics; database access limited.

## Claims

- Aiwei suggested that Royalty Graph v1/v2 usage data could be found in Temporal Cloud console (royalty-graph-v2 namespace), Grafana SOS dashboards, or PostgreSQL tables tracking royalty graph job state in SOS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- Aiwei suggested that IP Graph usage data may be available via the indexer Temporal namespace, its DB tables, and possibly Grafana indexer dashboards. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- Aiwei suggested that POC Modules (Collection, Licensing, Royalties) usage data is on-chain and would be sourced from transaction counts, contract call volume via Dune Analytics or internal blockchain indexing dashboards. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- RSI currently does not have read access to SOS or blockchain databases, preventing retrieval of usage data from those sources. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- Attempted DB read on depin-prod was denied by user U0772SH7BRA. The query was for table names in the public schema. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7` `chunk_id=srcchunk_045ff27753cd8b236824f84388aa8905` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810034.117999` `source_timestamp=2026-05-15T01:54:06Z`
- RSI attempted to access Temporal namespaces (indexer-production.koyiy, royalty-graph-production.koyiy, royalty-graph-v2-prod.koyiy, royalty-graph-staging-gcp.koyiy, royalty-graph-v2-staging.koyiy) but all connections failed with 'connection reset by peer'. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- Grafana has the Story Indexer dashboard, which provides partial IP Graph coverage, but there are no royalty or POC dashboards. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus does not contain any royalty/IP graph metrics. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- A link to the royalty-graph-v2-prod.koyiy workflows on Temporal Cloud was shared, though RSI could not connect to it. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- When asked whether an internal dashboard for royalty graph usage exists (for seeing royalties and transactions), the response was uncertain ('oh, do we have one?'). `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_53c073c35d14caf2817be96bf175fc8b` `chunk_id=srcchunk_b5017699187e6434648cc6c59d4abef9` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811506.588819` `source_timestamp=2026-05-15T02:18:26Z`

## Open Questions

- Are there any plans to create Grafana dashboards for Royalty Graph and POC Modules?
- How to grant RSI read access to SOS and blockchain databases for usage data?
- How to restore RSI access to Temporal namespaces (indexer-production, royalty-graph-production, royalty-graph-v2-*, etc.)?
- Is there an internal dashboard for Royalty Graph usage (royalties + transactions)? If so, where is it located and who owns it?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_61fd869e6f05dd0662ec576a85dfd618`
