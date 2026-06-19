---
title: "Payment Method Prioritization: Fiat vs Stablecoin"
type: "decision"
slug: "decisions/payment-method-prioritization-decision"
freshness: "2026-04-29T20:28:36Z"
tags:
  - "crypto"
  - "payments"
  - "stablecoin"
  - "strategy"
owners: []
source_revision_ids:
  - "srcrev_a1a2a522894bf560dd52a1684eb307ab"
  - "srcrev_bb90697db35eac44a4ddc680eee701e2"
  - "srcrev_df9a5efcdc45814500d6052e2537d3c2"
  - "srcrev_e75c689c7fea6ef0df457f91990a31fe"
  - "srcrev_ebbd22bcb3d9ec42246a16f1e13b6d2d"
conflict_state: "none"
---

# Payment Method Prioritization: Fiat vs Stablecoin

## Summary

Decision on whether to prioritize fiat or stablecoin for payouts, considering Kled intel on stablecoin usage and the upcoming launch of IPUSDC on Story chain.

## Claims

- Kled intel reports that 90% of payouts are in stablecoin. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9f8cd6a8340cf05c07dc81eff8a7791b` `source_revision_id=srcrev_bb90697db35eac44a4ddc680eee701e2` `chunk_id=srcchunk_8a22cd2d0d9c04e1b290eb7385b803fa` `native_locator=slack:C0AL7EKNHDF:1777494350.428769:1777494350.428769` `source_timestamp=2026-04-29T20:25:50Z`
- Option to offer crypto payments as a cheaper/faster option with weekly batch payouts and use payment provider as fallback. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9f8cd6a8340cf05c07dc81eff8a7791b` `source_revision_id=srcrev_bb90697db35eac44a4ddc680eee701e2` `chunk_id=srcchunk_8a22cd2d0d9c04e1b290eb7385b803fa` `native_locator=slack:C0AL7EKNHDF:1777494350.428769:1777494350.428769` `source_timestamp=2026-04-29T20:25:50Z`
- KYC may not be required for very small amounts, e.g., Whop only requires KYC for amounts over $5000. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9f8cd6a8340cf05c07dc81eff8a7791b` `source_revision_id=srcrev_bb90697db35eac44a4ddc680eee701e2` `chunk_id=srcchunk_8a22cd2d0d9c04e1b290eb7385b803fa` `native_locator=slack:C0AL7EKNHDF:1777494350.428769:1777494350.428769` `source_timestamp=2026-04-29T20:25:50Z`
- Q3 planned launch of IPUSDC stablecoin. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9f8cd6a8340cf05c07dc81eff8a7791b` `source_revision_id=srcrev_ebbd22bcb3d9ec42246a16f1e13b6d2d` `chunk_id=srcchunk_cd1b0f3a40efe0b13b0bf3097a7d780c` `native_locator=slack:C0AL7EKNHDF:1777494387.628209:1777494387.628209` `source_timestamp=2026-04-29T20:26:27Z`
  - citation: `source_document_id=srcdoc_9f8cd6a8340cf05c07dc81eff8a7791b` `source_revision_id=srcrev_a1a2a522894bf560dd52a1684eb307ab` `chunk_id=srcchunk_bb35bcdf9753564e20e17f5f76d7b07d` `native_locator=slack:C0AL7EKNHDF:1777494392.154509:1777494392.154509` `source_timestamp=2026-04-29T20:26:32Z`
- Use of stablecoin is considered simplest, and distribution can be forced on Story chain using USDC.e to increase chain usage. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9f8cd6a8340cf05c07dc81eff8a7791b` `source_revision_id=srcrev_df9a5efcdc45814500d6052e2537d3c2` `chunk_id=srcchunk_77cb1a924437d171662e86cad98d0e04` `native_locator=slack:C0AL7EKNHDF:1777494476.652369:1777494476.652369` `source_timestamp=2026-04-29T20:27:56Z`
- Decision needed on whether to prioritize fiat payment then stables or stablecoin then fiat. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9f8cd6a8340cf05c07dc81eff8a7791b` `source_revision_id=srcrev_e75c689c7fea6ef0df457f91990a31fe` `chunk_id=srcchunk_1ff8428e84b71a461f18e18e6d032c0a` `native_locator=slack:C0AL7EKNHDF:1777494516.850739:1777494516.850739` `source_timestamp=2026-04-29T20:28:36Z`

## Sources

- `source_document_id`: `srcdoc_9f8cd6a8340cf05c07dc81eff8a7791b`
- `source_revision_id`: `srcrev_e75c689c7fea6ef0df457f91990a31fe`
