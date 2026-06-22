---
title: "IP Graph"
type: "system"
slug: "systems/ip-graph"
freshness: "2026-05-15T01:55:24Z"
tags:
  - "graph"
  - "indexer"
  - "monitoring"
owners:
  - "U04L0DD6B6F"
  - "U04VDFP1YQ5"
  - "U05A515NBFC"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
  - "srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7"
conflict_state: "none"
---

# IP Graph

## Summary

IP Graph is a system likely built on top of an indexer Temporal namespace and database. Partial monitoring coverage exists via the Story Indexer Grafana dashboard, but no Prometheus metrics are available. RSI does not have access to supporting SOS or blockchain databases.

## Claims

- Story Indexer dashboard in Grafana provides partial IP Graph coverage. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus lacks IP Graph metrics. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- IP Graph likely uses an indexer Temporal namespace and its DB tables, and may have Grafana indexer dashboards. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- RSI cannot query SOS or blockchain databases, which would be needed for deeper IP Graph usage data. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- DB read attempts on depin-prod by RSI or others were denied, indicating restricted access. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7` `chunk_id=srcchunk_045ff27753cd8b236824f84388aa8905` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810034.117999` `source_timestamp=2026-05-15T01:54:06Z`

## Open Questions

- Are there Grafana indexer dashboards specifically for IP Graph? (Availability was described as conditional.)
- Does IP Graph have dedicated monitoring beyond the Story Indexer dashboard?

## Related Pages

- `poc-modules`
- `royalty-graph`
- `rsi-agent`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_71d9a06a743d957c08030ae93197ee4a`
