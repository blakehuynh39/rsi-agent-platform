---
title: "POC Modules (Collection, Licensing, Royalties)"
type: "system"
slug: "systems/poc-modules"
freshness: "2026-05-15T01:55:24Z"
tags:
  - "licensing"
  - "on-chain"
  - "poc"
  - "royalties"
owners:
  - "U04L0DD6B6F"
  - "U04VDFP1YQ5"
  - "U05A515NBFC"
source_revision_ids:
  - "srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa"
  - "srcrev_71094780a5c936f2948da743f0c75821"
  - "srcrev_c5f44a625d9cfb3a708b34791880757c"
conflict_state: "none"
---

# POC Modules (Collection, Licensing, Royalties)

## Summary

POC Modules encompass Collection, Licensing, and Royalties components. Their usage data is primarily on-chain. No internal dashboards or Prometheus metrics are available. RSI lacks blockchain DB read access, so it cannot query this data.

## Claims

- POC Modules usage data is on-chain (transaction counts, contract call volume). `claim:claim_3_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_71094780a5c936f2948da743f0c75821` `chunk_id=srcchunk_87f758c13e0547398046eafbfbd95e57` `native_locator=slack:C04T5307FNU:1778808153.324069:1778808289.440849` `source_timestamp=2026-05-15T01:24:56Z`
- No Grafana dashboards exist for POC modules; Grafana only has a Story Indexer dashboard (partial IP Graph). `claim:claim_3_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- Prometheus has no metrics for POC/royalty components. `claim:claim_3_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`
- RSI does not have blockchain DB read access, preventing it from querying on-chain POC usage data. `claim:claim_3_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_c5f44a625d9cfb3a708b34791880757c` `chunk_id=srcchunk_3f0854084e329e34f55895c02c72c82c` `native_locator=slack:C04T5307FNU:1778808153.324069:1778809852.245359` `source_timestamp=2026-05-15T01:50:52Z`
  - citation: `source_document_id=srcdoc_7c2c352c7fc64cf13951a268ecc3c379` `source_revision_id=srcrev_04b3a9c83cfd05d6412a5bb8b874e7fa` `chunk_id=srcchunk_9223be07393947bdf9d62c062383608d` `native_locator=slack:C04T5307FNU:1778808153.324069:1778810124.192439` `source_timestamp=2026-05-15T01:55:24Z`

## Open Questions

- Are there Dune Analytics dashboards or internal blockchain indexing dashboards for POC modules?
- Who owns the on-chain data queries for POC modules?

## Related Pages

- `ip-graph`
- `royalty-graph`
- `rsi-agent`

## Sources

- `source_document_id`: `srcdoc_7c2c352c7fc64cf13951a268ecc3c379`
- `source_revision_id`: `srcrev_71d9a06a743d957c08030ae93197ee4a`
