---
title: "RSI Investigation: Usage Data for Royalty Graph, IP Graph, POC Modules (2026-05-15)"
type: "runbook"
slug: "runbooks/rsi-investigation-2026-05-15-usage-data"
freshness: "2026-05-15T02:18:26Z"
tags:
  - "2026-05-15"
  - "investigation"
  - "monitoring"
  - "rsi"
  - "usage"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_41275761193e96d75e682a02e7e04d16"
  - "srcrev_53c073c35d14caf2817be96bf175fc8b"
  - "srcrev_61fd869e6f05dd0662ec576a85dfd618"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
  - "srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7"
conflict_state: "none"
---

# RSI Investigation: Usage Data for Royalty Graph, IP Graph, POC Modules (2026-05-15)

## Summary

On 2026-05-15, Yao requested usage data for Royalty Graph v1/v2, IP Graph, and POC Modules. RSI investigated available monitoring and data sources, revealing severe limitations in dashboards, database access, and Prometheus metrics.

## Claims

- On 2026-05-15, Yao (U083MMT1771) requested actual usage data for Royalty Graph v1/v2, IP Graph, and POC Modules Collection, Licensing, and Royalties. `claim:claim_4_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_61fd869e6f05dd0662ec576a85dfd618` `chunk_id=srcchunk_785dd8f247308c1b94fb2c7eff851031` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808153.324069` `source_timestamp=2026-05-15T01:22:33Z`
- RSI (bot) lacked read access to SOS or blockchain databases at the time of investigation. `claim:claim_4_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- RSI attempted to query Temporal namespaces but encountered connection reset errors, confirming the existence of royalty-graph-production.koyiy, royalty-graph-v2-prod.koyiy, royalty-graph-staging-gcp.koyiy, royalty-graph-v2-staging.koyiy, and indexer-production.koyiy. `claim:claim_4_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- Two DB read attempts to depin-prod were denied by U0772SH7BRA. Requesters were U067QP5PD6J and U083MMT1771. `claim:claim_4_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7` `chunk_id=srcchunk_045ff27753cd8b236824f84388aa8905` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810034.117999` `source_timestamp=2026-05-15T01:54:06Z`
- Grafana's Story Indexer dashboard provides partial IP Graph coverage, but no dashboards exist for royalty or POC modules. `claim:claim_4_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus has no royalty or IP graph metrics. `claim:claim_4_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- There is no internal dashboard for royalty graph transactions and royalties. `claim:claim_4_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_53c073c35d14caf2817be96bf175fc8b` `chunk_id=srcchunk_b5017699187e6434648cc6c59d4abef9` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811506.588819` `source_timestamp=2026-05-15T02:18:26Z`
- Royalty Graph v2 is accessible via Temporal Cloud at https://cloud.temporal.io/namespaces/royalty-graph-v2-prod.koyiy/workflows. `claim:claim_4_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`

## Open Questions

- How can RSI or other team members obtain read access to SOS/blockchain DBs for usage data?
- Who can provide Dune Analytics or internal blockchain dashboards for POC modules?
- Who owns the SOS repo and can extract royalty/IP graph metrics from its codebase?
- Why are Temporal namespaces unreachable from RSI?

## Related Pages

- `ip-graph`
- `poc-modules-collection-licensing-royalties`
- `royalty-graph-v1-v2`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_c3ccf444cfe45d7ef6bb2a9d5d8c06d3`
