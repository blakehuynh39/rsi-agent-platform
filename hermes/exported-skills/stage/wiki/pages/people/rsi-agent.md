---
title: "RSI Agent"
type: "person"
slug: "people/rsi-agent"
freshness: "2026-05-15T01:55:24Z"
tags:
  - "agent"
  - "ai"
  - "data-access"
  - "rsi"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
conflict_state: "none"
---

# RSI Agent

## Summary

RSI is an AI assistant that can answer questions using available data sources. Its current access is limited to the depin database and Grafana dashboards; it cannot read from SOS or blockchain databases, and its Temporal namespace connections are broken.

## Claims

- RSI does not have direct access to usage dashboards or live metrics. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- RSI does not have read access to SOS or blockchain databases yet. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- RSI can access Grafana and found only the Story Indexer dashboard (partial IP Graph coverage); no royalty or POC dashboards exist. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI attempted to query Temporal namespaces (indexer-production, royalty-graph-*, etc.) but all connections failed with 'connection reset by peer'. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- RSI can read from the depin database (limited), but SOS and blockchain DBs remain off‑limits. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`

## Open Questions

- When will RSI gain read access to SOS and blockchain DBs?
- Why is Temporal namespace connection failing for RSI?

## Related Pages

- `royalty-graph-monitoring`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_3c85a9061e5787d8a437266dc596c23f`
