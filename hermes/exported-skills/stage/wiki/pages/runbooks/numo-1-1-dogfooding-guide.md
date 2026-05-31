---
title: "Numo 1.1 Dogfooding Guide"
type: "runbook"
slug: "runbooks/numo-1-1-dogfooding-guide"
freshness: "2026-05-22T22:41:00Z"
tags:
  - "dogfooding"
  - "numo"
  - "payments"
  - "staging"
  - "stripe"
owners: []
source_revision_ids:
  - "srcrev_b31b0dcfafb2abf416ba68e00b384df5"
  - "srcrev_bc4c526b69fe5ca1e2c4dad655f3b5d0"
  - "srcrev_dd87e6979337019368f30b443e560f1e"
conflict_state: "none"
---

# Numo 1.1 Dogfooding Guide

## Summary

Internal guide for dogfooding the Numo 1.1 release, covering staging access, resume upload, and Stripe payments preview.

## Claims

- The staging environment for Numo is accessible at staging.numolabs.ai with the password 'numopip' for both access and login. `claim:claim_staging` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_bc4c526b69fe5ca1e2c4dad655f3b5d0` `chunk_id=srcchunk_c5a404dddb3a8a2df330ac988bf46588` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1` `source_timestamp=2026-05-22T22:41:00Z`
- Upon accessing the staging environment, users should see a banner prompting them to upload their resume. `claim:claim_resume_banner` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_bc4c526b69fe5ca1e2c4dad655f3b5d0` `chunk_id=srcchunk_c5a404dddb3a8a2df330ac988bf46588` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1` `source_timestamp=2026-05-22T22:41:00Z`
- Stripe payments setup is not part of the current release but is included as a preview for feedback; the process may change due to pending discussions on KYC and W-8BEN forms for foreign citizens. `claim:claim_stripe_preview` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_bc4c526b69fe5ca1e2c4dad655f3b5d0` `chunk_id=srcchunk_c21f9316791d424dbd5c7d893e91a845` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2` `source_timestamp=2026-05-22T22:41:00Z`
- Dogfooding participants are asked to complete the Stripe setup flow to align on the payout features. `claim:claim_stripe_setup_request` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_bc4c526b69fe5ca1e2c4dad655f3b5d0` `chunk_id=srcchunk_c21f9316791d424dbd5c7d893e91a845` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2` `source_timestamp=2026-05-22T22:41:00Z`
- The staging environment for Numo is accessible at staging.numolabs.ai with the password 'numopip' for both access and login. `claim:claim_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_dd87e6979337019368f30b443e560f1e` `chunk_id=srcchunk_063283b9769ecf163d261b3a77973462` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1` `source_timestamp=2026-05-22T20:49:00Z`
- The staging environment for Numo is accessible at staging.numolabs.ai with the password 'numopip' for both access and login. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_b31b0dcfafb2abf416ba68e00b384df5` `chunk_id=srcchunk_80706552e79b1111405e3d3353db219a` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1` `source_timestamp=2026-05-22T21:26:00Z`
- Upon accessing the staging environment, users should see a banner prompting them to upload their resume. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_b31b0dcfafb2abf416ba68e00b384df5` `chunk_id=srcchunk_80706552e79b1111405e3d3353db219a` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1` `source_timestamp=2026-05-22T21:26:00Z`
- Stripe payments setup is not part of the current release but is included as a preview for feedback; the process may change due to pending discussions on KYC and W-8BEN forms for foreign citizens. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_b31b0dcfafb2abf416ba68e00b384df5` `chunk_id=srcchunk_34a5eb6323cb56a9126d267636ba7745` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2` `source_timestamp=2026-05-22T21:26:00Z`
- Dogfooding participants are asked to complete the Stripe setup flow to align on the payout features. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_b31b0dcfafb2abf416ba68e00b384df5` `chunk_id=srcchunk_34a5eb6323cb56a9126d267636ba7745` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2` `source_timestamp=2026-05-22T21:26:00Z`
- Upon accessing the staging environment, users should see a banner prompting them to upload their resume. `claim:claim_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_dd87e6979337019368f30b443e560f1e` `chunk_id=srcchunk_063283b9769ecf163d261b3a77973462` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-1` `source_timestamp=2026-05-22T20:49:00Z`
- Stripe payments setup is not part of the current release but is included as a preview for feedback; the process may change due to pending discussions on KYC and W-8BEN forms for foreign citizens. `claim:claim_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_dd87e6979337019368f30b443e560f1e` `chunk_id=srcchunk_4c981d0e3d67468984ff3df730e439f5` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2` `source_timestamp=2026-05-22T20:49:00Z`
- Dogfooding participants are asked to complete the Stripe setup flow to align on the payout features. `claim:claim_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2) `source_document_id=srcdoc_c89b865ceeeab41ee41b6cde44392313` `source_revision_id=srcrev_dd87e6979337019368f30b443e560f1e` `chunk_id=srcchunk_4c981d0e3d67468984ff3df730e439f5` `native_locator=https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac#chunk-2` `source_timestamp=2026-05-22T20:49:00Z`

## Sources

- `source_document_id`: `srcdoc_c89b865ceeeab41ee41b6cde44392313`
- `source_revision_id`: `srcrev_bc4c526b69fe5ca1e2c4dad655f3b5d0`
- `source_url`: [Notion source](https://www.notion.so/Numo-1-1-Dogfooding-Guide-a28051299a5483aea76e01ba0a8d0fac)
