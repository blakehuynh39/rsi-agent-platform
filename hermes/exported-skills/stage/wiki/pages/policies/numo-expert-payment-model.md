---
title: "Numo Expert Payment Model"
type: "policy"
slug: "policies/numo-expert-payment-model"
freshness: "2026-06-23T19:33:00Z"
tags:
  - "annotation"
  - "compensation"
  - "contributors"
  - "payment"
  - "per-task"
owners:
  - "Nick Ippolito"
source_revision_ids:
  - "srcrev_762b66916d85eb7989e67f2a050e024f"
conflict_state: "none"
---

# Numo Expert Payment Model

## Summary

Defines per-task pay for contributors based on real task time and a target skilled hourly wage. Recommends $10/hr for Bengali annotation, yielding per-task rates from $0.26 to $0.70, a 55-65% reduction from current hourly contracts. Advises grandfathering existing contractors during migration.

## Claims

- The payment model is owned by Nick Ippolito and is in draft for review. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_5692806bd05574f33fdd4f35a6e1d2a6` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1` `source_timestamp=2026-06-23T19:33:00Z`
- The target hourly wage for skilled bilingual Bengali annotation is $8-12/hr, with a recommended midpoint of $10/hr. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_5692806bd05574f33fdd4f35a6e1d2a6` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1` `source_timestamp=2026-06-23T19:33:00Z`
- Per‑task pay is calculated as target hourly wage × (real seconds per task / 3600). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_5692806bd05574f33fdd4f35a6e1d2a6` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1` `source_timestamp=2026-06-23T19:33:00Z`
- Real time for TRANSCRIPT_CORRECTION is 252 seconds, derived from contractor ground truth of 50 transcripts in 3.5 hours (midpoint). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-2) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_872cd3c81dc35c69834edd1a559dd0ae` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-2` `source_timestamp=2026-06-23T19:33:00Z`
- The annotation tool timer undercounts real effort by ~20‑30% (logged mean 197s vs real 252s, factor 1.28). `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_5692806bd05574f33fdd4f35a6e1d2a6` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1` `source_timestamp=2026-06-23T19:33:00Z`
- Recommended per‑task rates at $10/hr: Transcript Correction $0.70, Audio Match $0.53, Correction Validation $0.46 (provisional), Pre‑sub Sanity $0.26 (provisional). `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-2) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_872cd3c81dc35c69834edd1a559dd0ae` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-2` `source_timestamp=2026-06-23T19:33:00Z`
- The current effective rate for 50‑transcript batches is ~$25/hr ($75‑100 per batch), considered overpay. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_db393cc14fc97e92145e8040f233f035` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3` `source_timestamp=2026-06-23T19:33:00Z`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_5692806bd05574f33fdd4f35a6e1d2a6` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1` `source_timestamp=2026-06-23T19:33:00Z`
- The new model reduces per‑batch cost by 55‑65% relative to current pay (e.g., $87.50 → $35 for 50 Transcript Corrections at $10/hr). `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_db393cc14fc97e92145e8040f233f035` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3` `source_timestamp=2026-06-23T19:33:00Z`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_5692806bd05574f33fdd4f35a6e1d2a6` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-1` `source_timestamp=2026-06-23T19:33:00Z`
- The recommended migration path is to grandfather current contractors at existing rates through Phase 1/2, then converge to the new band, to avoid losing experienced contributors. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_db393cc14fc97e92145e8040f233f035` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3` `source_timestamp=2026-06-23T19:33:00Z`
- Bengali consensus annotators per item is set at 5 (per PRD Section 10.1). `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-2) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_872cd3c81dc35c69834edd1a559dd0ae` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-2` `source_timestamp=2026-06-23T19:33:00Z`
- QA overhead (honeypots ~10% of queue + reviewer escalation) adds 10‑15% on fully‑loaded item cost. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-2) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_872cd3c81dc35c69834edd1a559dd0ae` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-2` `source_timestamp=2026-06-23T19:33:00Z`
- For Bengali, the skilled talent market (India and Bangladesh) is treated as one at $8‑12/hr, with no major geo‑differentiation recommended. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_db393cc14fc97e92145e8040f233f035` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3` `source_timestamp=2026-06-23T19:33:00Z`
- The local‑competitive floor for generic microwork (Transcript Correction) is estimated at $0.25‑0.35 per task across India, SEA, East Asia, and Egypt, but these are not recommended for skilled Bengali work. `claim:claim_1_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3) `source_document_id=srcdoc_570a89beb86d727201f015129ef5bd03` `source_revision_id=srcrev_762b66916d85eb7989e67f2a050e024f` `chunk_id=srcchunk_db393cc14fc97e92145e8040f233f035` `native_locator=https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205#chunk-3` `source_timestamp=2026-06-23T19:33:00Z`

## Open Questions

- Correction Validation and Pre-sub Sanity time estimates are provisional and need to be replaced with real production data once submissions are available.
- The decision between grandfathering versus immediate tiered pay is a business decision to confirm, not a modelling output.
- The transcript correction time estimate carries ±14% uncertainty (3h→216s vs 4h→288s); a stopwatch study during contractor migration is recommended to pin it down.
- The ‘50 transcripts’ batch is assumed to be TRANSCRIPT_CORRECTION; if it is a different task type, the 252s anchor shifts and all rates must be re-derived.

## Sources

- `source_document_id`: `srcdoc_570a89beb86d727201f015129ef5bd03`
- `source_revision_id`: `srcrev_762b66916d85eb7989e67f2a050e024f`
- `source_url`: [source](https://app.notion.com/p/Numo-Expert-Payment-Model-388051299a5480fe99f0eb7d17e72205)
