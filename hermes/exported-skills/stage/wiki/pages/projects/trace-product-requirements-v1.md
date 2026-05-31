---
title: "Trace V1 Product Requirements"
type: "project"
slug: "projects/trace-product-requirements-v1"
freshness: "2026-05-12T18:55:00Z"
tags:
  - "audit-portal"
  - "product-requirements"
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
  - "srcrev_0a09dd219125919f30170c377326c65f"
conflict_state: "none"
---

# Trace V1 Product Requirements

## Summary

Trace is the public audit layer for data registered on the Protocol. It generates immutable receipts for contributions from apps like Kled, Numo, Oto, and Miso, enabling labs to verify dataset legitimacy, contributors to confirm consent terms, and regulators to audit AI training data provenance. V1 ships a central portal at trace.thedatafoundation.ai with whitelabel embeds for partners, ingesting receipts via webhook, pinning them onchain through Story Protocol, and surfacing them in app-level, user-level, and asset-level views.

## Claims

- Trace V1 targets a mainnet rollout and rebrand launch on June 15, 2026. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_c349edc422eaef1420dbaa7f07affbef` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T18:55:00Z`
- Trace is a public audit layer for data registered on the Protocol, generating immutable receipts for every contribution from contributing apps. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_c349edc422eaef1420dbaa7f07affbef` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T18:55:00Z`
- Receipts surface in three views: app-level (contributing app compliance posture), user-level (aggregate contributor metadata), and asset-level (individual receipt). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_c349edc422eaef1420dbaa7f07affbef` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T18:55:00Z`
- Trace standardizes on SHA-256 with multihash prefix as the canonical content-addressing scheme. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-3) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_233c9f7bfd88fcc82cd268e3e345a6f2` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-3` `source_timestamp=2026-05-12T18:55:00Z`
- Trace uses a metadata-presence model that stores whether sensitive metadata exists on an asset, not the actual values. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_34175c4edb6132d8426d34f3a825fce5` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2` `source_timestamp=2026-05-12T18:55:00Z`
- Trace V1 non-goals include fraud detection, originality verification, consumer-facing file lookup, CLI tools for batch hash lookups, and rendered diff view between TOS/PP versions. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_c349edc422eaef1420dbaa7f07affbef` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-12T18:55:00Z`
- The whitelabel portal for partners is hosted at subdomains like audit.kled.ai, with branded chrome and required 'Powered by Trace / The Data Foundation' attribution. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-4) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_d14c25ed5fe14e08bc18844dfcee1b4f` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-4` `source_timestamp=2026-05-12T18:55:00Z`
- Ingestion throughput targets sustained ~60 receipts/sec (5M+ records/day) and peak ~250 receipts/sec across all partners. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_25cca6a481fc9a0fcbd9018d76c5f6f2` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5` `source_timestamp=2026-05-12T18:55:00Z`
- Kled's 1.1 billion record backlog will be migrated over a ~1.5 to 2 month window using an exponential hockey-stick schedule, targeting completion around end of July/early August for token unlock alignment. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_25cca6a481fc9a0fcbd9018d76c5f6f2` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5` `source_timestamp=2026-05-12T18:55:00Z`
- Receipt ingestion is durable via Temporal workflow, with idempotency by idempotency_key, automatic retries, and a max retry ceiling proposal of 24 hours with exponential backoff before manual escalation. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_25cca6a481fc9a0fcbd9018d76c5f6f2` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-5` `source_timestamp=2026-05-12T18:55:00Z`
- The lifecycle of a receipt is append-only; apps post updates to the same asset.hash with new lifecycle events, and Trace appends without overwriting. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-3) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_0a09dd219125919f30170c377326c65f` `chunk_id=srcchunk_233c9f7bfd88fcc82cd268e3e345a6f2` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-3` `source_timestamp=2026-05-12T18:55:00Z`

## Open Questions

- Aggregate analytics over very large scoped groups (>10M hashes): sync with a heavy index or async with a job page? Blake to call after looking at infrastructure costs.
- Compliance frameworks enum governance: who decides when eu-ai-act or a new framework gets added? Proposal: Avneet + Julie + Avi own that decision.
- Contributor anon_id portability across apps: out of scope for V1, worth revisiting if labs ask for cross-app contributor reputation signals.
- Gas batching window of ~1 min is a placeholder — Romain to size based on simulated Kled volume + measured gas costs.
- Gas limit per block at projected volume — Romain to model.
- Max Temporal retry ceiling before manual escalation — proposal 24h, Blake to confirm.
- Onchain signature for consent: store consent signatures onchain or just the signature hash? Flag if legal wants full signature recoverable.
- Self-serve tenant config dashboard timing: hand-rolled for first 5 partners, dashboard after?

## Related Pages

- `kled-backlog-migration`
- `trace-data-model`
- `trace-hash-function`
- `trace-webhook-api`

## Sources

- `source_document_id`: `srcdoc_f34407b922c741d72df23780d7864b93`
- `source_revision_id`: `srcrev_0a09dd219125919f30170c377326c65f`
- `source_url`: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8)
