---
title: "RSI Agent"
type: "system"
slug: "systems/rsi-agent"
freshness: "2026-05-15T01:55:24Z"
tags:
  - "ai-agent"
  - "monitoring"
  - "rsi"
owners:
  - "U0ASDQKU3UL"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_08cae80f2873293e78663b739464a798"
  - "srcrev_177669b86527bc5a48233fdfea52cb0d"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_71d9a06a743d957c08030ae93197ee4a"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
  - "srcrev_f3606868517ce9344cf7c5587bd3e82d"
  - "srcrev_f5bcdcbd7c4349bf12a16f9b355cc860"
conflict_state: "none"
---

# RSI Agent

## Summary

RSI is an AI agent integrated into Slack, capable of accessing certain data sources like Grafana, Temporal, and depin database, but lacks access to SOS and blockchain databases. It assists with investigating system usage data but currently has connection issues to Temporal.

## Claims

- RSI does not have read access to SOS or blockchain databases. `claim:claim_4_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI can query Grafana and found the Story Indexer dashboard, confirming no royalty or POC dashboards exist. `claim:claim_4_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI attempted to query Temporal namespaces but currently faces connection reset errors, preventing usage data retrieval. `claim:claim_4_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- RSI can track investigation runs via a platform link. `claim:claim_4_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_08cae80f2873293e78663b739464a798` `chunk_id=srcchunk_7736c846d55d165c982d77c39f8b56b2` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809829.531549` `source_timestamp=2026-05-15T01:50:29Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_f3606868517ce9344cf7c5587bd3e82d` `chunk_id=srcchunk_7f3153ece6485dda74c791e7cdfe13c6` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809975.439779` `source_timestamp=2026-05-15T01:52:55Z`
- RSI can be mentioned in Slack to perform tasks like investigating usage data. `claim:claim_4_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71d9a06a743d957c08030ae93197ee4a` `chunk_id=srcchunk_59df869eedb4d42a446364cd0b7222c0` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809825.643479` `source_timestamp=2026-05-15T01:50:25Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_177669b86527bc5a48233fdfea52cb0d` `chunk_id=srcchunk_bf709db2bbd6b6008f6972a8d1f73b3b` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809934.445769` `source_timestamp=2026-05-15T01:52:14Z`
- Aiwei Agent is a separate bot/app user that may block direct messages. `claim:claim_4_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_f5bcdcbd7c4349bf12a16f9b355cc860` `chunk_id=srcchunk_847642dd6f713e95f89ac10f7d58f043` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808230.793089` `source_timestamp=2026-05-15T01:23:52Z`

## Related Pages

- `ip-graph`
- `poc-modules`
- `royalty-graph`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_71d9a06a743d957c08030ae93197ee4a`
