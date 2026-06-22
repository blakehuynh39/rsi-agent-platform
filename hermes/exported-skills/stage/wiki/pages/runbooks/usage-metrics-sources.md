---
title: "Usage Metrics Sources for Royalty Graph, IP Graph, and POC Modules"
type: "runbook"
slug: "runbooks/usage-metrics-sources"
freshness: "2026-05-15T02:16:39Z"
tags:
  - "grafana"
  - "ip-graph"
  - "monitoring"
  - "poc-modules"
  - "royalty-graph"
  - "temporal"
owners:
  - "U04L0DD6B6F"
  - "U04VDFP1YQ5"
  - "U0772SH7BRA"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7"
conflict_state: "none"
---

# Usage Metrics Sources for Royalty Graph, IP Graph, and POC Modules

## Summary

Current state and locations of usage metrics for Royalty Graph v1/v2, IP Graph, and POC Modules Collection, Licensing, and Royalties as of May 2026, compiled from a Slack investigation.

## Claims

- Royalty Graph v1/v2 usage data is expected to be available through Temporal Cloud namespaces (e.g., `royalty-graph-v2-prod.koyiy`), Grafana SOS dashboards, or PostgreSQL tables in SOS tracking royalty graph job state. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- IP Graph usage data likely resides in the indexer Temporal namespace and its database tables, with possible Grafana indexer dashboards. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- POC Modules Collection, Licensing, and Royalties usage data is protocol-layer on-chain data, available via Dune Analytics or internal blockchain indexing dashboards. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- As of 2026-05-15, RSI's attempt to access Temporal namespaces (`indexer-production.koyiy`, `royalty-graph-production.koyiy`, `royalty-graph-v2-prod.koyiy`, `royalty-graph-staging-gcp.koyiy`, `royalty-graph-v2-staging.koyiy`) failed with 'connection reset by peer'. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- Grafana currently has a Story Indexer dashboard providing partial IP Graph coverage, but no dedicated royalty or POC dashboards were found. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Database read access is limited to `depin-prod`; SOS and blockchain databases are not accessible. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus metrics do not include royalty or IP graph metrics. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Direct read requests to the `depin-prod` database (e.g., listing tables) were denied by user @U0772SH7BRA. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7` `chunk_id=srcchunk_045ff27753cd8b236824f84388aa8905` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810034.117999` `source_timestamp=2026-05-15T01:54:06Z`
- A direct link to the Temporal Cloud workflows for `royalty-graph-v2-prod.koyiy` is available: https://cloud.temporal.io/namespaces/royalty-graph-v2-prod.koyiy/workflows. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`

## Open Questions

- Is there an internal dashboard for royalty amounts and transactions beyond Temporal workflows?
- Who can grant access to SOS or blockchain databases for metrics queries?
- Why are RSI's connections to Temporal namespaces failing (connection reset by peer)?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_2afce762999c8ea208bf478a3d7bcdcb`
