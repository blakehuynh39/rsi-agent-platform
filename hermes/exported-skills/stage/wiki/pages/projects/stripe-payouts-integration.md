---
title: "Stripe Payouts Integration"
type: "project"
slug: "projects/stripe-payouts-integration"
freshness: "2026-04-30T19:54:25Z"
tags:
  - "compliance"
  - "integration"
  - "payouts"
  - "stripe"
owners: []
source_revision_ids:
  - "srcrev_1a41ad3c249863d9c96aa3d95ee357d2"
  - "srcrev_2e8d6ef9a9cc1ef10581a0f19a1193c1"
  - "srcrev_99fb1cb9bac63bb9360c79751229fdbb"
  - "srcrev_bfbc95c18151d5e66ddba48b499e7691"
conflict_state: "none"
---

# Stripe Payouts Integration

## Summary

Integration of Stripe payouts to enable sending USD to Indian bank accounts in INR, with Stripe handling KYC and compliance.

## Claims

- A full end-to-end payouts flow was successfully tested in Stripe sandbox, sending USD to an Indian bank account in INR. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9387ba5a6485fb3bff185174c4d7be52` `source_revision_id=srcrev_1a41ad3c249863d9c96aa3d95ee357d2` `chunk_id=srcchunk_875f7c3a05812854671a4e0796543426` `native_locator=slack:C0AL7EKNHDF:1777501727.413219:1777501727.413219` `source_timestamp=2026-04-29T22:28:47Z`
  - citation: `source_document_id=srcdoc_9387ba5a6485fb3bff185174c4d7be52` `source_revision_id=srcrev_bfbc95c18151d5e66ddba48b499e7691` `chunk_id=srcchunk_8ccfc70af71b9615d8dde5cb04eadbaa` `native_locator=slack:C0AL7EKNHDF:1777501727.413219:1777501786.706229` `source_timestamp=2026-04-29T22:29:46Z`
- UX flow: user is created with only an email; payment initiated without bank details; users are redirected to a custom landing page that generates a Stripe payments link (short-lived 10 min URL) to collect bank info and PII; Stripe handles the payment initiation directly. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9387ba5a6485fb3bff185174c4d7be52` `source_revision_id=srcrev_1a41ad3c249863d9c96aa3d95ee357d2` `chunk_id=srcchunk_875f7c3a05812854671a4e0796543426` `native_locator=slack:C0AL7EKNHDF:1777501727.413219:1777501727.413219` `source_timestamp=2026-04-29T22:28:47Z`
- Benefits: low integration effort for RSI (only email required), safer because RSI does not handle PII, and user has flexibility to collect payment when they want. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9387ba5a6485fb3bff185174c4d7be52` `source_revision_id=srcrev_1a41ad3c249863d9c96aa3d95ee357d2` `chunk_id=srcchunk_875f7c3a05812854671a4e0796543426` `native_locator=slack:C0AL7EKNHDF:1777501727.413219:1777501727.413219` `source_timestamp=2026-04-29T22:28:47Z`
- Sandbox cost estimate: ₹1,000 INR delivered to recipient cost $10.55 USD debited from RSI financial account; fees $1.69 ($1.50 standard + $0.08 cross-border + $0.11 FX). Sandbox FX is illustrative; real costs need live validation. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9387ba5a6485fb3bff185174c4d7be52` `source_revision_id=srcrev_1a41ad3c249863d9c96aa3d95ee357d2` `chunk_id=srcchunk_875f7c3a05812854671a4e0796543426` `native_locator=slack:C0AL7EKNHDF:1777501727.413219:1777501727.413219` `source_timestamp=2026-04-29T22:28:47Z`
- Compliance: Stripe handles tax form collection via W-8BEN product; no US withholding required for foreign-source payouts; KYC managed through Stripe hosted onboarding form; OFAC screening may need additional service integration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9387ba5a6485fb3bff185174c4d7be52` `source_revision_id=srcrev_99fb1cb9bac63bb9360c79751229fdbb` `chunk_id=srcchunk_c08628685686928bed6c7fb6934cb9ef` `native_locator=slack:C0AL7EKNHDF:1777501727.413219:1777564813.236389` `source_timestamp=2026-04-30T16:07:20Z`
- The team views Stripe's handling of most compliance (tax and KYC) as crucial because RSI has no dedicated compliance team. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9387ba5a6485fb3bff185174c4d7be52` `source_revision_id=srcrev_2e8d6ef9a9cc1ef10581a0f19a1193c1` `chunk_id=srcchunk_6dfc7d332c0a421c824d6c8b6da55b68` `native_locator=slack:C0AL7EKNHDF:1777501727.413219:1777578865.051079` `source_timestamp=2026-04-30T19:54:25Z`

## Open Questions

- Crypto payments approval
- Custom landing page server hosting details
- KYC/tax compliance finalization
- Live account testing and confirmation
- OFAC screening integration details

## Sources

- `source_document_id`: `srcdoc_9387ba5a6485fb3bff185174c4d7be52`
- `source_revision_id`: `srcrev_2e8d6ef9a9cc1ef10581a0f19a1193c1`
