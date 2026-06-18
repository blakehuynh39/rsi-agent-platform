---
title: "Stripe Payouts Integration"
type: "decision"
slug: "decisions/stripe-payouts-integration"
freshness: "2026-06-09T22:34:26Z"
tags:
  - "country-support"
  - "crypto"
  - "fees"
  - "global-payouts"
  - "payouts"
  - "stablecoin"
  - "stripe"
  - "wire-transfer"
owners: []
source_revision_ids:
  - "srcrev_0437c50152eea01f23564ed755984a44"
  - "srcrev_0ca24e32d379193ec14a6edd2ac3b80d"
  - "srcrev_14921143541819f0178c5f016a04da6c"
  - "srcrev_3429c435834efae380a5463e00ea0476"
  - "srcrev_881f0021b8760ad685f42e220f057faf"
  - "srcrev_907fc2e5fb7cf2ba64b64f356b2d45c6"
  - "srcrev_9bda0c57d3c614c51f578f99aeb2b30e"
  - "srcrev_a2ce25f8185ca7e9e28b92b6bbe6cc89"
  - "srcrev_bcffde51010c15a9d96406e7c1d41b02"
  - "srcrev_bf0ab09bce15e97ee3b5f042e4c0a694"
conflict_state: "none"
---

# Stripe Payouts Integration

## Summary

Decisions and challenges around integrating Stripe for international creator payouts. Initial confusion over Connect vs Global Payouts, discovery that some countries require expensive wire transfers, unsupported countries like Bangladesh, and exploration of stablecoin payouts as a cheaper alternative.

## Claims

- Stripe support initially told us that Connect setup is not needed and that Global Payouts would be sufficient. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_881f0021b8760ad685f42e220f057faf` `chunk_id=srcchunk_5e84d53aa8027c52162a66cf4000cdc8` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1780991668.610319` `source_timestamp=2026-06-09T07:54:28Z`
- Despite support's guidance, errors persisted, and the team started onboarding Connect v1, but encountered identity verification errors. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_3429c435834efae380a5463e00ea0476` `chunk_id=srcchunk_feb27ccb935122fa0bc788bbe2939dab` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1780991896.353139` `source_timestamp=2026-06-09T07:58:16Z`
- Debugging revealed that some countries only support wire transfers (not local bank rails), which caused the payment failures. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_bf0ab09bce15e97ee3b5f042e4c0a694` `chunk_id=srcchunk_b08f98433e658cb14161b71a81b37895` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1780999245.026829` `source_timestamp=2026-06-09T10:00:45Z`
- Wire transfer fees are estimated at $7-$15, making payouts as low as $25 potentially uneconomical. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_14921143541819f0178c5f016a04da6c` `chunk_id=srcchunk_27f5528c0cb1a059107d8f5790e4cf6e` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1780999356.626229` `source_timestamp=2026-06-09T10:02:36Z`
- Vietnam and Pakistan were confirmed to require wire per Stripe’s classification. The team decided to temporarily use wire for these countries while exploring stablecoin payments. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_0437c50152eea01f23564ed755984a44` `chunk_id=srcchunk_ab536baa10e80574455409f2139fa6af` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1781040695.352069` `source_timestamp=2026-06-09T21:31:35Z`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_0ca24e32d379193ec14a6edd2ac3b80d` `chunk_id=srcchunk_187bb83d7e08c19e1a1594e9956bf3f9` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1781040813.071049` `source_timestamp=2026-06-09T21:33:33Z`
- Bangladesh cannot be paid via Stripe at all (neither Global Payouts nor Connect v1). Stablecoin payments are proposed as a solution. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_a2ce25f8185ca7e9e28b92b6bbe6cc89` `chunk_id=srcchunk_71f51619e280834900f7d71776946973` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1781024326.130839` `source_timestamp=2026-06-09T16:58:46Z`
- Stripe has approved stablecoin payouts for our account, and a stablecoin-based flow is being developed to cover unsupported countries and reduce costs. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_bcffde51010c15a9d96406e7c1d41b02` `chunk_id=srcchunk_42d1bfc51857b7d2a783950455975fd9` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1781000004.817609` `source_timestamp=2026-06-09T10:13:43Z`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_907fc2e5fb7cf2ba64b64f356b2d45c6` `chunk_id=srcchunk_9e33c6d6904d54101335b979032c7c0b` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1781026440.048279` `source_timestamp=2026-06-09T17:34:00Z`
- A tester from Pakistan completed payout setup, but the wallet balance did not decrement after approval, indicating a possible balance update issue. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_19b7d24007003cf2a1db88b0a401585c` `source_revision_id=srcrev_9bda0c57d3c614c51f578f99aeb2b30e` `chunk_id=srcchunk_81b480ba5de081643ad400c64a2d75f9` `native_locator=slack:C0AL7EKNHDF:1780991225.749039:1781044466.748219` `source_timestamp=2026-06-09T22:34:26Z`

## Open Questions

- How to support payouts to Korea for the Numo mobile team collaboration?
- How will wallet balance deduction be handled after payout approval?
- What is the timeline for implementing stablecoin payouts?
- What is the user base count per country to assess fee impact?

## Sources

- `source_document_id`: `srcdoc_19b7d24007003cf2a1db88b0a401585c`
- `source_revision_id`: `srcrev_ce1b427e1c871538d8defef768b7ca6e`
