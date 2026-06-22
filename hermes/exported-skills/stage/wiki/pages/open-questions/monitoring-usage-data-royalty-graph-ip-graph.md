---
title: "Monitoring Usage Data for Royalty Graph and IP Graph"
type: "open_question"
slug: "open-questions/monitoring-usage-data-royalty-graph-ip-graph"
freshness: "2026-05-15T02:18:26Z"
tags:
  - "ip-graph"
  - "monitoring"
  - "poc-modules"
  - "royalty-graph"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_41275761193e96d75e682a02e7e04d16"
  - "srcrev_53c073c35d14caf2817be96bf175fc8b"
  - "srcrev_986a7cfb5f5a9b4f23031b60a8e8e529"
  - "srcrev_9d61bcc4440d2a39c7e0b3093dc01292"
  - "srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7"
conflict_state: "none"
---

# Monitoring Usage Data for Royalty Graph and IP Graph

## Summary

As of 2026-05-15, there is no internal business dashboard for Royalty Graph (royalties/transactions), only a Temporal Cloud namespace link. IP Graph has partial coverage via Grafana Story Indexer dashboard. POC Modules (Collection, Licensing, Royalties) lack dashboards. Temporal access is broken, DB access restricted, Prometheus missing metrics.

## Claims

- There is no dedicated internal dashboard for Royalty Graph royalties/transactions; the only available link is to Temporal Cloud namespace royalty-graph-v2-prod.koyiy. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_9d61bcc4440d2a39c7e0b3093dc01292` `chunk_id=srcchunk_f020c23f0726d7e113846ad8ed062aa2` `native_locator=slack:C04T5307FNU:1778811399.075409` `source_timestamp=2026-05-15T02:16:39Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_986a7cfb5f5a9b4f23031b60a8e8e529` `chunk_id=srcchunk_fafe4d5c7c4066db94efd23269ed9939` `native_locator=slack:C04T5307FNU:1778811433.646089` `source_timestamp=2026-05-15T02:17:13Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_41275761193e96d75e682a02e7e04d16` `chunk_id=srcchunk_23857ef547e14155e84234f810ed1a78` `native_locator=slack:C04T5307FNU:1778811499.765429` `source_timestamp=2026-05-15T02:18:19Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_53c073c35d14caf2817be96bf175fc8b` `chunk_id=srcchunk_b5017699187e6434648cc6c59d4abef9` `native_locator=slack:C04T5307FNU:1778811506.588819` `source_timestamp=2026-05-15T02:18:26Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- IP Graph usage is partially covered by the Grafana Story Indexer dashboard. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- POC Modules (Collection, Licensing, Royalties) have no dashboards or Prometheus metrics. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI attempted to access Temporal Cloud namespaces (royalty-graph-v2-prod, indexer-production, etc.) but connections failed with 'connection reset by peer'. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- RSI attempted to read depin-prod public schema tables but was denied by @U0772SH7BRA. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7` `chunk_id=srcchunk_045ff27753cd8b236824f84388aa8905` `native_locator=slack:C04T5307FNU:1778810034.117999` `source_timestamp=2026-05-15T01:54:06Z`
- Prometheus has no royalty/IP graph metrics. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`

## Open Questions

- Can SOS/blockchain DB read access be granted for monitoring purposes?
- Can Temporal access be restored for RSI to query namespace metrics?
- Is there a plan to create an internal dashboard for royalty graph and POC modules?
- Where can actual usage data (royalties, transactions) for Royalty Graph be obtained?

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_f3606868517ce9344cf7c5587bd3e82d`
