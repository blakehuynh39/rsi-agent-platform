---
title: "RSI Agent"
type: "system"
slug: "systems/rsi-agent"
freshness: "2026-05-15T01:55:39Z"
tags:
  - "ai-agent"
  - "data-access"
  - "internal-tool"
owners: []
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_3bb1978a214eaea1884e884550a7c07f"
  - "srcrev_3c85a9061e5787d8a437266dc596c23f"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
  - "srcrev_f5bcdcbd7c4349bf12a16f9b355cc860"
  - "srcrev_f9a26ebbd3baf467fef114060a2f2d5a"
conflict_state: "none"
---

# RSI Agent

## Summary

The RSI (Research & Support Intelligence) agent is an internal bot that assists with investigations. As of this thread, its capabilities include Slack interaction, Temporal namespace probing (though connections are currently broken), Grafana dashboard inspection, and DB reads (limited to depin; SOS and blockchain DBs are blocked). It cannot initiate direct DMs with some bot users.

## Claims

- RSI agent cannot send direct messages to the Aiwei Agent bot user due to Slack configuration. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_f5bcdcbd7c4349bf12a16f9b355cc860` `chunk_id=srcchunk_847642dd6f713e95f89ac10f7d58f043` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808230.793089` `source_timestamp=2026-05-15T01:23:52Z`
- RSI agent lacks read access to SOS and blockchain databases. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
- A DB read request by RSI to the depin-prod database was denied by a user (U0772SH7BRA). `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3c85a9061e5787d8a437266dc596c23f` `chunk_id=srcchunk_8ee4e08621e93dcdd726baa990ee4201` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809953.747589` `source_timestamp=2026-05-15T01:52:42Z`
- RSI agent's attempts to connect to Temporal namespaces (including royalty and indexer namespaces) failed with 'connection reset by peer'. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_3bb1978a214eaea1884e884550a7c07f` `chunk_id=srcchunk_4f9bcc302b14a1f2be28e8a991c1adb3` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809963.941999` `source_timestamp=2026-05-15T01:52:46Z`
- RSI agent successfully investigated Grafana and found a Story Indexer dashboard but no royalty or POC dashboards. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI agent can inspect Prometheus and reported no royalty/IP graph metrics. `claim:claim_2_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI agent offered to dig through the SOS codebase for metrics, indicating potential code-level access. `claim:claim_2_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_f9a26ebbd3baf467fef114060a2f2d5a` `chunk_id=srcchunk_f24316a2c55750f8f77f2da2f3fa639d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810137.026989` `source_timestamp=2026-05-15T01:55:39Z`

## Open Questions

- What additional capabilities does RSI agent have beyond what was demonstrated in this thread?
- Who can grant RSI agent read access to SOS and blockchain DBs?
- Why are Temporal namespace connections failing for RSI agent?

## Related Pages

- `usage-data-monitoring`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_ebd5bd890ebdf2b0ac04b0f66becacf7`
