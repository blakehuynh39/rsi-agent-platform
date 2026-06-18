---
title: "Stripe Connect Setup for Vietnam and Pakistan Payouts"
type: "project"
slug: "projects/stripe-connect-vietnam-pakistan-payouts"
freshness: "2026-06-09T01:34:30Z"
tags:
  - "pakistan"
  - "payouts"
  - "poseidon-ai"
  - "stripe-connect"
  - "vietnam"
owners:
  - "U083MMT1771"
  - "U08951K4SRY"
  - "U09QGMMUDPC"
source_revision_ids:
  - "srcrev_47f4d97ea6dcc492c753644171982b28"
  - "srcrev_74c2719cfc6a042fe19d9f6601c113c5"
  - "srcrev_8ea459cf3f89456300481a196d854bc1"
  - "srcrev_a84c31571a9000c954c871f1ded001fb"
  - "srcrev_f76a82b1c58eccfa06ba1905514281c1"
conflict_state: "none"
---

# Stripe Connect Setup for Vietnam and Pakistan Payouts

## Summary

Integration of Stripe Connect to enable payouts to contributors in Vietnam and Pakistan, evaluating v2 Global Payouts versus v1 classic Express accounts.

## Claims

- Stripe Connect must be set up in the Poseidon AI Dashboard to enable payouts in Vietnam and Pakistan. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_8ea459cf3f89456300481a196d854bc1` `chunk_id=srcchunk_321d47b286b4685fd55694928cf655a0` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780946192.224779` `source_timestamp=2026-06-08T19:16:32Z`
- Errors during testing are due to Stripe platform Connect configuration defaults (card-processing capabilities) not being adjusted, not app bugs. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_8ea459cf3f89456300481a196d854bc1` `chunk_id=srcchunk_321d47b286b4685fd55694928cf655a0` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780946192.224779` `source_timestamp=2026-06-08T19:16:32Z`
- The v1 classic Express path (POST /v1/accounts) returns a 400 error 'You can only create new accounts if you've signed up for Connect.' for VN, PK, BD, ID. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_f76a82b1c58eccfa06ba1905514281c1` `chunk_id=srcchunk_285bcd8f7811af47dfe106ca9b25d99b` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780957160.163429` `source_timestamp=2026-06-08T22:19:20Z`
- India works end-to-end using the v2 Global Payouts path (POST /v2/core/accounts). `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_f76a82b1c58eccfa06ba1905514281c1` `chunk_id=srcchunk_285bcd8f7811af47dfe106ca9b25d99b` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780957160.163429` `source_timestamp=2026-06-08T22:19:20Z`
- Stripe support confirmed that Pakistan and Vietnam are supported recipients of Global Payouts on the Poseidon AI account. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_a84c31571a9000c954c871f1ded001fb` `chunk_id=srcchunk_3fe0008b7878401b859b8288165cdb53` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780952229.075199` `source_timestamp=2026-06-08T20:57:09Z`
- The v2 Global Payouts path previously showed 'features not available for your account' for Vietnam when requesting the bank_accounts.local capability. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_f76a82b1c58eccfa06ba1905514281c1` `chunk_id=srcchunk_285bcd8f7811af47dfe106ca9b25d99b` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780957160.163429` `source_timestamp=2026-06-08T22:19:20Z`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_74c2719cfc6a042fe19d9f6601c113c5` `chunk_id=srcchunk_0d1be43a71da3275564f55ed0c07ed35` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780968870.712719` `source_timestamp=2026-06-09T01:34:30Z`
- The preferred outcome is to standardize on one payment path for all recipient countries, pending advice on v2 vs. v1 and the correct capability. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_f76a82b1c58eccfa06ba1905514281c1` `chunk_id=srcchunk_285bcd8f7811af47dfe106ca9b25d99b` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780957160.163429` `source_timestamp=2026-06-08T22:19:20Z`
- The Connect onboarding flow is being completed by the team. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_47f4d97ea6dcc492c753644171982b28` `chunk_id=srcchunk_750c52d022b7e54901dcf263a232a81c` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780952247.390909` `source_timestamp=2026-06-08T20:57:38Z`
- Stripe support is requesting API call details for further debugging. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b02240e50e66f4348ac7d1f187297ade` `source_revision_id=srcrev_a84c31571a9000c954c871f1ded001fb` `chunk_id=srcchunk_3fe0008b7878401b859b8288165cdb53` `native_locator=slack:C0AL7EKNHDF:1780946192.224779:1780952229.075199` `source_timestamp=2026-06-08T20:57:09Z`

## Open Questions

- Has the 'features not available' error for Vietnam in v2 been resolved?
- Should the team standardize on v2 Global Payouts or keep v1 Express for cross-border markets?
- Which recipient capability should be requested for Vietnam and Pakistan in v2 Global Payouts (bank_accounts.local or alternative)?

## Sources

- `source_document_id`: `srcdoc_b02240e50e66f4348ac7d1f187297ade`
- `source_revision_id`: `srcrev_74c2719cfc6a042fe19d9f6601c113c5`
