---
title: "Tax Forms (TaxBandits) Integration"
type: "system"
slug: "systems/tax-forms-taxbandits-integration"
freshness: "2026-06-09T00:40:12Z"
tags:
  - "payouts"
  - "pdf-download"
  - "tax-forms"
  - "taxbandits"
owners:
  - "U05A515NBFC"
  - "U083MMT1771"
  - "U08951K4SRY"
  - "U09QGMMUDPC"
source_revision_ids:
  - "srcrev_1d71060613726992d9947a3b5698400d"
  - "srcrev_4d4dd7850aad43e4efff573a598c8d60"
  - "srcrev_5c68ac4e276c8a129a90268793394b8a"
  - "srcrev_6bebd192e8640ae1f66c53c83432d08f"
  - "srcrev_98860fcc535c6e4a6963889dde473c58"
  - "srcrev_a2888824ca7ef645209cc07bc69466af"
conflict_state: "none"
---

# Tax Forms (TaxBandits) Integration

## Summary

Integration of TaxBandits for W-8/W-9 tax form collection within the withdrawal flow, including admin resync, design improvements, and configuration for encrypted PDF downloads.

## Claims

- The current UI allows submitting tax forms independently of Stripe setup, but both must be completed for withdrawal. Clicking 'complete tax form' opens TaxBandits in a new tab. Automatic update after completion; forms can be downloaded via admin dashboard. If status not updating, admin can trigger a 'resync'. Completed forms are 'verified', replaced ones become 'superseded'. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_85f0affbd897f71a84eda9181ae58578` `source_revision_id=srcrev_5c68ac4e276c8a129a90268793394b8a` `chunk_id=srcchunk_a667c35f17aa8591d13cd98ffc84cdbf` `native_locator=slack:C0AL7EKNHDF:1780929501.598549:1780929501.598549` `source_timestamp=2026-06-08T14:38:21Z`
- PDF downloading from TaxBandits requires additional environment variables in the backend. Currently set up with TaxBandits sandbox credentials; production credentials need to be replaced. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_85f0affbd897f71a84eda9181ae58578` `source_revision_id=srcrev_1d71060613726992d9947a3b5698400d` `chunk_id=srcchunk_15986a213d889b511b23f9807e8b2eb6` `native_locator=slack:C0AL7EKNHDF:1780929501.598549:1780929560.516469` `source_timestamp=2026-06-08T14:39:20Z`
  - citation: `source_document_id=srcdoc_85f0affbd897f71a84eda9181ae58578` `source_revision_id=srcrev_98860fcc535c6e4a6963889dde473c58` `chunk_id=srcchunk_060e4d4d08b6a2d780743b5687b1bbb7` `native_locator=slack:C0AL7EKNHDF:1780929501.598549:1780929593.598319` `source_timestamp=2026-06-08T14:39:53Z`
- A design proposal exists to improve the tax form UI copy and clarify the independence of steps, addressing false dependency and explaining purpose. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_85f0affbd897f71a84eda9181ae58578` `source_revision_id=srcrev_4d4dd7850aad43e4efff573a598c8d60` `chunk_id=srcchunk_6019091c97daf8763869b8e1f0560ae2` `native_locator=slack:C0AL7EKNHDF:1780929501.598549:1780964079.473099` `source_timestamp=2026-06-09T00:14:39Z`
- To enable tax form PDF downloading, a specific set of environment variables must be configured, including TaxBandits OAuth credentials, webhook reference, and optional PDF encryption keys (SSE-C). Minimum set: TAX_COLLECTION_ENABLED=true, TAXBANDITS_CLIENT_ID, TAXBANDITS_CLIENT_SECRET, TAXBANDITS_USER_TOKEN, TAXBANDITS_BUSINESS_ID, TAXBANDITS_WEBHOOK_REF, TAX_FORM_PDF_STORAGE_ENABLED=false, TAX_PAYOUT_GATING_ENABLED=true. Production requires OAuth and API host overrides. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_85f0affbd897f71a84eda9181ae58578` `source_revision_id=srcrev_6bebd192e8640ae1f66c53c83432d08f` `chunk_id=srcchunk_07cd6b7efbecc35c4dc2c17467e0d83a` `native_locator=slack:C0AL7EKNHDF:1780929501.598549:1780964905.206739` `source_timestamp=2026-06-09T00:28:25Z`
- PR #435 adds TAXBANDITS_PDF_* environment variables for encrypted TaxBandits W-9/W-8 PDF download, with staging active and production commented-out. Review passed all 7 checks, including correct Vault refs, no plaintext secrets, and Helm configuration. Non-blocking: staging has stale S3-related config with storage disabled by default. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_85f0affbd897f71a84eda9181ae58578` `source_revision_id=srcrev_a2888824ca7ef645209cc07bc69466af` `chunk_id=srcchunk_e694f6f71c14181ce710b53c91949833` `native_locator=slack:C0AL7EKNHDF:1780929501.598549:1780965612.142889` `source_timestamp=2026-06-09T00:40:12Z`

## Open Questions

- Are the production TaxBandits credentials ready to be deployed?
- Should the design proposal for improved UI copy be implemented?

## Related Pages

- `payments-withdrawal-flow`
- `story-deployments`
- `stripe-integration`

## Sources

- `source_document_id`: `srcdoc_85f0affbd897f71a84eda9181ae58578`
- `source_revision_id`: `srcrev_a2888824ca7ef645209cc07bc69466af`
