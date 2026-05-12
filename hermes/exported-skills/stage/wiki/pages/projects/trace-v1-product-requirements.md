---
title: "Trace V1 Product Requirements"
type: "project"
slug: "projects/trace-v1-product-requirements"
freshness: "2026-05-12T21:14:00Z"
tags:
  - "audit"
  - "data-provenance"
  - "prd"
  - "trace"
  - "v1"
owners:
  - "Allen"
  - "Andrea Muttoni"
  - "Avi"
  - "Avneet, Julie"
  - "Blake"
  - "Jacob"
  - "Raul + Lion team"
  - "Romain"
  - "Susan"
source_revision_ids:
  - "srcrev_3e3d65144fcf8ba96739dceceb8a5ac1"
conflict_state: "none"
---

# Trace V1 Product Requirements

## Summary

Trace is the public audit layer for AI training data registered on the Protocol. V1 ships a central public portal at trace.thedatafoundation.ai and whitelabel embeds for contributing apps, targeting a June 15, 2026 mainnet launch.

## Claims

- Trace is the public audit layer for data registered on the Protocol, generating immutable receipts for every contribution from contributing apps. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_ac8e5cca5c4de9c8b4674369babf5812` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T21:14:00Z`
- The target ship date for Trace V1 mainnet rollout and rebrand launch day is June 15, 2026. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_ac8e5cca5c4de9c8b4674369babf5812` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T21:14:00Z`
- V1 includes a central portal at trace.thedatafoundation.ai and whitelabel embeds at partner subdomains (e.g., audit.kled.ai, audit.oto.ai). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_ac8e5cca5c4de9c8b4674369babf5812` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T21:14:00Z`
- Trace is a receipt layer; it does not host data, run fraud detection, or store values of sensitive metadata. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_ac8e5cca5c4de9c8b4674369babf5812` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T21:14:00Z`
- The data model consists of three nested scopes: app-level, user-level, and asset-level, relevant for both central and whitelabel portals. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_8ae06fb1bf33ec6201bfc96ae6c3e2fb` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2` `source_timestamp=2026-05-12T21:14:00Z`
- Trace uses a metadata-presence model that stores whether sensitive signals exist on the asset (e.g., exif: present) but never the actual values. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_8ae06fb1bf33ec6201bfc96ae6c3e2fb` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2` `source_timestamp=2026-05-12T21:14:00Z`
- The canonical content-addressing scheme is SHA-256 with multihash prefix; the API form is `sha256:<64-hex-chars>` and the onchain form is Multihash `0x1220<32-byte-digest>`. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-3) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_64f22727b5f4135b47f063fc072f51d9` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-3` `source_timestamp=2026-05-12T21:14:00Z`
- Receipt ingestion is durable via Temporal workflow with idempotency on `idempotency_key` and automatic retry with exponential backoff. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_74a0ca9f750c7e7bf1f883fe12b1bb08` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5` `source_timestamp=2026-05-12T21:14:00Z`
- Sustained ingestion target is ~60 receipts/sec (5M+ records/day), with peak ingestion of ~250 receipts/sec and indexing latency p50 < 60s, p99 < 5 min. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_74a0ca9f750c7e7bf1f883fe12b1bb08` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5` `source_timestamp=2026-05-12T21:14:00Z`
- Kled's 1.1B record backlog will be migrated over ~1.5-2 months with a hockey-stick schedule to coincide with token unlock timing. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_74a0ca9f750c7e7bf1f883fe12b1bb08` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5` `source_timestamp=2026-05-12T21:14:00Z`
- The product owner is Andrea Muttoni, with oversight by Allen, backend by Blake, smart contracts by Romain, frontend by Jacob, and supporting engineering by Raul + Lion team. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_ac8e5cca5c4de9c8b4674369babf5812` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T21:14:00Z`
- Non-goals for V1 include fraud detection, consumer file lookup, CLI tool, rendered diff view between TOS/PP versions, and hosting actual content. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_3e3d65144fcf8ba96739dceceb8a5ac1` `chunk_id=srcchunk_ac8e5cca5c4de9c8b4674369babf5812` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T21:14:00Z`

## Open Questions

- Aggregate analytics over very large scoped groups (>10M hashes): sync with heavy index or async with job page?
- Compliance frameworks enum governance: who decides when new frameworks get added?
- Contributor anon_id portability across apps: out of scope for V1; revisit if labs request cross-app reputation signals.
- Gas batching window of ~1 min is a placeholder; Romain to size based on simulated Kled volume + measured gas costs.
- Gas limit per block at projected volume — Romain to model.
- Max Temporal retry ceiling before manual escalation — proposal 24h, Blake to confirm.
- Onchain signature for consent: store full signature or just signature_hash?
- Self-serve tenant config dashboard timing: hand-rolled for first 5 partners, dashboard after.

## Related Pages

- `kled-backlog-migration`
- `trace-data-model`

## Sources

- `source_document_id`: `srcdoc_f34407b922c741d72df23780d7864b93`
- `source_revision_id`: `srcrev_3e3d65144fcf8ba96739dceceb8a5ac1`
- `source_url`: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8)
