---
title: "Usage Monitoring Data for Royalty Graph, IP Graph, and POC Modules"
type: "system"
slug: "systems/usage-monitoring-data-royalty-ip-graph-poc"
freshness: "2026-05-15T02:16:39Z"
tags:
  - "data-access"
  - "ip-graph"
  - "monitoring"
  - "poc-modules"
  - "royalty-graph"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
conflict_state: "none"
---

# Usage Monitoring Data for Royalty Graph, IP Graph, and POC Modules

## Summary

Documentation of the current state of access to usage and monitoring data for Royalty Graph v1/v2, IP Graph, and POC Modules (Collection, Licensing, Royalties). Includes available dashboards, database restrictions, and connection issues.

## Claims

- Royalty Graph v2 production workflows can be monitored via Temporal Cloud namespace 'royalty-graph-v2-prod.koyiy' at https://cloud.temporal.io/namespaces/royalty-graph-v2-prod.koyiy/workflows. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
- Grafana provides a Story Indexer dashboard with partial IP Graph coverage, but no royalty or POC dashboards are available. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Database access for SOS and blockchain is not available to RSI, limiting usage data retrieval. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- Multiple Temporal namespaces related to royalty graph and indexer exist but were inaccessible by RSI due to 'connection reset by peer' errors. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- Prometheus does not have any royalty or IP graph metrics. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- DB read requests to 'depin-prod' were denied for listing tables, indicating restricted access even to non-sensitive schema information. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`

## Open Questions

- Is there an internal dashboard for royalty graph usage (royalties and transactions) beyond Temporal? No one in the thread knew of one.
- Who owns the usage metrics or dashboards for POC modules (Collection, Licensing, Royalties)? On-chain data sources like Dune Analytics were suggested.
- Why are Temporal namespace connections failing? Is this a transient issue or a configuration problem?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_8e08632a9a6a3f2fdf2491c07753b20c`
