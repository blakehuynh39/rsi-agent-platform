---
title: "Implement and Test Payment Solution for P0 countries - Stripe PoC"
type: "project"
slug: "projects/stripe-poc-payment-solution"
freshness: "2026-05-05T17:00:00Z"
tags:
  - "global-payouts"
  - "p0-countries"
  - "payments"
  - "poc"
  - "stripe"
owners: []
source_revision_ids:
  - "srcrev_cfe94b01f4be61add34cd6cd938e21b5"
conflict_state: "none"
---

# Implement and Test Payment Solution for P0 countries - Stripe PoC

## Summary

Project to implement and test a Stripe-based payment solution for P0 countries, including a proof of concept for global payouts and safeguards against spam and malicious withdrawals.

## Claims

- A P1 safeguard mechanism on withdrawal is planned to prevent spam and malicious users. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4) `source_document_id=srcdoc_056ec0e3d256e88b47958ac3dab627af` `source_revision_id=srcrev_cfe94b01f4be61add34cd6cd938e21b5` `chunk_id=srcchunk_187fe5fcae743a0cc77f590999caec22` `native_locator=https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4` `source_timestamp=2026-05-05T17:00:00Z`
- Users can only withdraw rewards of processed tasks (approved or rejected); they cannot withdraw rewards of unprocessed tasks (e.g., submission reward). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4) `source_document_id=srcdoc_056ec0e3d256e88b47958ac3dab627af` `source_revision_id=srcrev_cfe94b01f4be61add34cd6cd938e21b5` `chunk_id=srcchunk_187fe5fcae743a0cc77f590999caec22` `native_locator=https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4` `source_timestamp=2026-05-05T17:00:00Z`
- Users must have X amount of approved tasks to withdraw; otherwise they are flagged as spam accounts and may be banned or shown spam warnings. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4) `source_document_id=srcdoc_056ec0e3d256e88b47958ac3dab627af` `source_revision_id=srcrev_cfe94b01f4be61add34cd6cd938e21b5` `chunk_id=srcchunk_187fe5fcae743a0cc77f590999caec22` `native_locator=https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4` `source_timestamp=2026-05-05T17:00:00Z`
- Access to the Stripe dashboard can be requested from specific team members. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4) `source_document_id=srcdoc_056ec0e3d256e88b47958ac3dab627af` `source_revision_id=srcrev_cfe94b01f4be61add34cd6cd938e21b5` `chunk_id=srcchunk_187fe5fcae743a0cc77f590999caec22` `native_locator=https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4` `source_timestamp=2026-05-05T17:00:00Z`
- The Stripe dashboard URL for global payouts is https://dashboard.stripe.com/acct_1TP3QFBJqeLC13SA/global-payouts/overview. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4) `source_document_id=srcdoc_056ec0e3d256e88b47958ac3dab627af` `source_revision_id=srcrev_cfe94b01f4be61add34cd6cd938e21b5` `chunk_id=srcchunk_187fe5fcae743a0cc77f590999caec22` `native_locator=https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4` `source_timestamp=2026-05-05T17:00:00Z`

## Sources

- `source_document_id`: `srcdoc_056ec0e3d256e88b47958ac3dab627af`
- `source_revision_id`: `srcrev_81aff07399ccd2d77a48a2511342508b`
- `source_url`: [Notion source](https://www.notion.so/Implement-and-Test-Payment-Solution-for-P0-countries-Stripe-PoC-34f051299a5480508170ea006981f7a4)
