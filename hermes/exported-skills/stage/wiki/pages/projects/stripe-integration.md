---
title: "Stripe Integration"
type: "project"
slug: "projects/stripe-integration"
freshness: "2026-05-13T23:42:51Z"
tags:
  - "fiat-payouts"
  - "onboarding"
  - "payments"
  - "stripe"
owners:
  - "U083MMT1771"
  - "U09QGMMUDPC"
source_revision_ids:
  - "srcrev_1cf07627e42daef07a9f96ee0b8b4ebe"
  - "srcrev_4970f2527fed30bb41dd71d274977044"
  - "srcrev_575c23bbfd772ed55927b2d649ed237d"
  - "srcrev_6e734fcfab88427275d07d771a9d65c8"
  - "srcrev_726239f8d4c1d0baaba96e15096a5b0e"
  - "srcrev_85491324fbd000bf7aebc16bb9ca065c"
  - "srcrev_86e236cddd089b9c0d776833e013278a"
  - "srcrev_a0094724bc0a454cd2114f15f463b257"
  - "srcrev_d3de2dafec43c1e934cfb80b6f9e8a68"
  - "srcrev_e6dfddcee1c1ebca2a818300c7f0d991"
conflict_state: "none"
---

# Stripe Integration

## Summary

Integration of Stripe for fiat payouts, including live account setup, onboarding flow, live testing with Indian bank, compliance (W-8BEN), and exploration of Wise/Whop for cost reduction.

## Claims

- Stripe integration has a working proof of concept on testnet. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_1cf07627e42daef07a9f96ee0b8b4ebe` `chunk_id=srcchunk_7bd4c5eacc3e7f45db11c8aa8a6d9db0` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778625274.022519` `source_timestamp=2026-05-12T22:34:34Z`
- A live Stripe account is already set up with sandbox and production accounts. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_a0094724bc0a454cd2114f15f463b257` `chunk_id=srcchunk_79a37f73fa0700c68525ea06c5bccebf` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778625158.032099` `source_timestamp=2026-05-12T22:32:38Z`
- Funds have been initiated on the Stripe production account via ACH, awaiting availability. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_4970f2527fed30bb41dd71d274977044` `chunk_id=srcchunk_67663e6d506be67b81051873967b1727` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778625230.639489` `source_timestamp=2026-05-12T22:33:50Z`
- Live fiat payout test sent $25 to an Indian bank: ₹2,389.15 credited, fees $1.94 (7.76%), scales to ~3% at $100+. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_86e236cddd089b9c0d776833e013278a` `chunk_id=srcchunk_170dbc9ecc4d67c34cdf481e1b7b3ba9` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778652506.727779` `source_timestamp=2026-05-13T06:08:26Z`
- Stripe hosted flow handles user onboarding, bank detail collection, and PII without our systems touching sensitive data. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_86e236cddd089b9c0d776833e013278a` `chunk_id=srcchunk_170dbc9ecc4d67c34cdf481e1b7b3ba9` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778652506.727779` `source_timestamp=2026-05-13T06:08:26Z`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_85491324fbd000bf7aebc16bb9ca065c` `chunk_id=srcchunk_507134840b6a3d33c556f6dafc069e57` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778654397.438869` `source_timestamp=2026-05-13T06:39:57Z`
- W-8BEN form collection integration is required for legal/compliance and remains in progress. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_86e236cddd089b9c0d776833e013278a` `chunk_id=srcchunk_170dbc9ecc4d67c34cdf481e1b7b3ba9` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778652506.727779` `source_timestamp=2026-05-13T06:08:26Z`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_6e734fcfab88427275d07d771a9d65c8` `chunk_id=srcchunk_010567ec79932c8bc806a38b5b83a621` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778704637.242189` `source_timestamp=2026-05-13T20:37:29Z`
- Wise integration and Whop integration are being evaluated for lower costs (Wise under 3% vs Stripe 6-7%). `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_575c23bbfd772ed55927b2d649ed237d` `chunk_id=srcchunk_5d5a10f2da7bc09597dd82362b98ec13` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778652720.131829` `source_timestamp=2026-05-13T06:12:00Z`
- Stripe restricts US platforms from creating Express accounts in India without explicit approval; staging tests hit this error. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_e6dfddcee1c1ebca2a818300c7f0d991` `chunk_id=srcchunk_76131568bcc2845e5fcbcf53a3f4fb72` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778715771.454739` `source_timestamp=2026-05-13T23:42:51Z`
- Branding and other configuration on Stripe still needs to be completed. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_d3de2dafec43c1e934cfb80b6f9e8a68` `chunk_id=srcchunk_0dd005177078eea2cbeb4b3777072164` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778654465.194019` `source_timestamp=2026-05-13T06:41:05Z`
- Stripe is expected to be fully set up by end of week so test payments can begin the following Monday. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f` `source_revision_id=srcrev_726239f8d4c1d0baaba96e15096a5b0e` `chunk_id=srcchunk_fec20fd9cad02094d36d46f8462dc646` `native_locator=slack:C0AL7EKNHDF:1778624319.726139:1778626389.636919` `source_timestamp=2026-05-12T22:53:09Z`

## Open Questions

- Branding and configuration finalization
- Decision on short-term vs. long-term payout provider (Stripe vs. Wise/Whop)
- Status of India account creation approval from Stripe
- W-8BEN integration completion and legal sign-off

## Related Pages

- `wise-integration`

## Sources

- `source_document_id`: `srcdoc_0c35cba9c7a3c5c078d632c7a3d8551f`
- `source_revision_id`: `srcrev_c4c7835d9107d080c6eb6fd76fb4dc4c`
