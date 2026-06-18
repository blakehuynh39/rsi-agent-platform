---
title: "Tax Integration Provider Selection"
type: "decision"
slug: "decisions/tax-integration-provider-selection"
freshness: "2026-06-01T14:09:23Z"
tags:
  - "integration"
  - "korea"
  - "payments"
  - "tax"
owners:
  - "@U09QGMMUDPC"
source_revision_ids:
  - "srcrev_0983fbf3bb7fabe7742b94788cf10be2"
  - "srcrev_49231a7b054fa7d78fef1aa011f1912c"
  - "srcrev_fa240eb417c68e86739fd902620afc14"
  - "srcrev_fd7f0bfe20ee15129f20c748944a5865"
conflict_state: "none"
---

# Tax Integration Provider Selection

## Summary

Evaluation of tax integration providers (Avalara vs Taxbandits) for Numo's payments feature, driven by DRI @U09QGMMUDPC, with target launch by June 10 at risk.

## Claims

- The DRI for payment and tax form integrations is @U09QGMMUDPC. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_fd7f0bfe20ee15129f20c748944a5865` `chunk_id=srcchunk_6031928f68f070b3389eb3e0dca3ec03` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780302597.222739` `source_timestamp=2026-06-01T08:31:34Z`
- Payments launch was targeted for Wednesday, June 10 (presumed June 10, 2026). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_fd7f0bfe20ee15129f20c748944a5865` `chunk_id=srcchunk_6031928f68f070b3389eb3e0dca3ec03` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780302597.222739` `source_timestamp=2026-06-01T08:31:34Z`
- The launch is at risk due to pending third-party integration confirmation. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_fd7f0bfe20ee15129f20c748944a5865` `chunk_id=srcchunk_6031928f68f070b3389eb3e0dca3ec03` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780302597.222739` `source_timestamp=2026-06-01T08:31:34Z`
- The DRI was asked to provide an updated estimate after confirming provider and price. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_fd7f0bfe20ee15129f20c748944a5865` `chunk_id=srcchunk_6031928f68f070b3389eb3e0dca3ec03` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780302597.222739` `source_timestamp=2026-06-01T08:31:34Z`
- Avalara was considered for tax collection. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_0983fbf3bb7fabe7742b94788cf10be2` `chunk_id=srcchunk_55ae2bb041ceb51488a195886917e75b` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780320617.520979` `source_timestamp=2026-06-01T13:30:17Z`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_49231a7b054fa7d78fef1aa011f1912c` `chunk_id=srcchunk_7d8694a4a783a1531e9779fcbf01c548` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780320927.547869` `source_timestamp=2026-06-01T13:35:27Z`
- A reference repository (github.com/modrinth/code) uses a similar stack with Stripe and Avalara. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_49231a7b054fa7d78fef1aa011f1912c` `chunk_id=srcchunk_7d8694a4a783a1531e9779fcbf01c548` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780320927.547869` `source_timestamp=2026-06-01T13:35:27Z`
- Taxbandits is being evaluated; it is considered better for W8/W9 forms and cheaper. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_fa240eb417c68e86739fd902620afc14` `chunk_id=srcchunk_5543ad50a2b82ad7b610593d93c19d5b` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780322963.366649` `source_timestamp=2026-06-01T14:09:23Z`
- A proof of concept for Taxbandits is underway, with a report expected by end of day. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_77e62990b2792b1e64ed0bf9e81874e4` `source_revision_id=srcrev_fa240eb417c68e86739fd902620afc14` `chunk_id=srcchunk_5543ad50a2b82ad7b610593d93c19d5b` `native_locator=slack:C0AL7EKNHDF:1780302597.222739:1780322963.366649` `source_timestamp=2026-06-01T14:09:23Z`

## Open Questions

- Can the Payments launch happen by June 10?
- Will Taxbandits be selected as the tax provider?

## Sources

- `source_document_id`: `srcdoc_77e62990b2792b1e64ed0bf9e81874e4`
- `source_revision_id`: `srcrev_fa240eb417c68e86739fd902620afc14`
