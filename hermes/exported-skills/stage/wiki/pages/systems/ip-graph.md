---
title: "IP Graph"
type: "system"
slug: "systems/ip-graph"
freshness: "2026-05-15T01:55:24Z"
tags:
  - "grafana"
  - "indexer"
  - "ip-graph"
  - "monitoring"
  - "usage"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_fc678feaf2a5de2bd549a627adb491fe"
conflict_state: "none"
---

# IP Graph

## Summary

IP Graph is the component that indexes IP assets. Monitoring is partially covered by the Story Indexer Grafana dashboard and a Temporal namespace, but no dedicated dashboard or Prometheus metrics exist.

## Claims

- A Grafana ‘Story Indexer’ dashboard provides partial IP Graph coverage. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- IP Graph usage is tracked in Temporal namespace indexer-production.koyiy. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_fc678feaf2a5de2bd549a627adb491fe` `chunk_id=srcchunk_f9c4fb134e1b75c75c65cf633d0155bb` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808277.576839` `source_timestamp=2026-05-15T01:24:37Z`
- No dedicated Grafana dashboard or Prometheus metrics exist for IP Graph beyond the Story Indexer dashboard. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`

## Open Questions

- What exact metrics are available for IP Graph in the Grafana Story Indexer dashboard?
- Who owns IP Graph monitoring?

## Related Pages

- `poc-modules-collection-licensing-royalties`
- `royalty-graph-v1-v2`
- `rsi-investigation-2026-05-15-usage-data`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_c3ccf444cfe45d7ef6bb2a9d5d8c06d3`
