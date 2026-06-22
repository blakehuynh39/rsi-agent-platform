---
title: "RSI Agent"
type: "system"
slug: "systems/rsi-agent"
freshness: "2026-05-15T01:55:24Z"
tags:
  - "bot"
  - "database"
  - "monitoring"
  - "rsi"
owners:
  - "U0ASDQKU3UL"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
conflict_state: "none"
---

# RSI Agent

## Summary

RSI is a bot that investigated usage data availability. It has read access to `depin-prod` database, can query Grafana, but lacks access to SOS, blockchain DBs, and Temporal namespaces.

## Claims

- RSI investigated Grafana, DB, and Prometheus to answer usage data questions. `claim:claim_5_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI has read access to `depin-prod` database, but not to SOS or blockchain databases. `claim:claim_5_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI found no royalty/IP graph metrics in Prometheus. `claim:claim_5_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`

## Open Questions

- What additional accesses can be granted to RSI to enable comprehensive monitoring?

## Related Pages

- `ip-graph`
- `monitoring-usage-data`
- `poc-modules`
- `royalty-graph`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_177669b86527bc5a48233fdfea52cb0d`
