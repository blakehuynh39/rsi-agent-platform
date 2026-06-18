---
title: "Tax Form Intermittent Disappearance and Error (June 2026)"
type: "system"
slug: "systems/tax-form-intermittent-disappearance-june-2026"
freshness: "2026-06-12T00:20:07Z"
tags:
  - "bug"
  - "tax-form"
  - "w-8"
  - "withdrawal"
owners:
  - "U083MMT1771"
  - "U08951K4SRY"
source_revision_ids:
  - "srcrev_6f30bb27ff6b460c7247f0077fc9d0c3"
  - "srcrev_7da146b3b5e9de225c3980aaf51455cd"
  - "srcrev_c74672035ea55122dd70ce3865dffaec"
  - "srcrev_d1397bb7a8e4cc2200b4a7bae4da5630"
conflict_state: "none"
---

# Tax Form Intermittent Disappearance and Error (June 2026)

## Summary

On June 11, 2026, a bug was reported where the tax form setup button intermittently disappeared and clicking the tax button resulted in an empty page and error. Spamming buttons led to payment processing without showing the tax form. A PR was filed to fix the issue. Staging appeared unaffected. The W-8 form was later successfully set up in production. The fix was deployed at 16:13:10 UTC on June 11, 2026.

## Claims

- Tax form info/button sometimes disappears and only shows typical withdrawal option. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bc7a7c361bc02e27a4c0b3415f22c13e` `source_revision_id=srcrev_7da146b3b5e9de225c3980aaf51455cd` `chunk_id=srcchunk_833cf95f2964849840bef8a9111f1c1d` `native_locator=slack:C0AL7EKNHDF:1781219392.813279:1781219392.813279` `source_timestamp=2026-06-11T23:09:52Z`
- Tax button opens and closes an empty page, then shows an error. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bc7a7c361bc02e27a4c0b3415f22c13e` `source_revision_id=srcrev_7da146b3b5e9de225c3980aaf51455cd` `chunk_id=srcchunk_833cf95f2964849840bef8a9111f1c1d` `native_locator=slack:C0AL7EKNHDF:1781219392.813279:1781219392.813279` `source_timestamp=2026-06-11T23:09:52Z`
- Spamming the buttons processed a payment without the tax form ever being shown, and the payment went into review. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bc7a7c361bc02e27a4c0b3415f22c13e` `source_revision_id=srcrev_7da146b3b5e9de225c3980aaf51455cd` `chunk_id=srcchunk_833cf95f2964849840bef8a9111f1c1d` `native_locator=slack:C0AL7EKNHDF:1781219392.813279:1781219392.813279` `source_timestamp=2026-06-11T23:09:52Z`
- A pull request was filed to fix the issue. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bc7a7c361bc02e27a4c0b3415f22c13e` `source_revision_id=srcrev_6f30bb27ff6b460c7247f0077fc9d0c3` `chunk_id=srcchunk_964439342f8082ebab68329a9100afe7` `native_locator=slack:C0AL7EKNHDF:1781219392.813279:1781223185.690509` `source_timestamp=2026-06-12T00:13:05Z`
- Staging environment did not exhibit the bug during previous testing. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bc7a7c361bc02e27a4c0b3415f22c13e` `source_revision_id=srcrev_c74672035ea55122dd70ce3865dffaec` `chunk_id=srcchunk_ee2d5ba0704f9437b7e1dca911ca8af7` `native_locator=slack:C0AL7EKNHDF:1781219392.813279:1781223404.844679` `source_timestamp=2026-06-12T00:16:44Z`
- W-8 form was successfully set up in production and was under manual review. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bc7a7c361bc02e27a4c0b3415f22c13e` `source_revision_id=srcrev_c74672035ea55122dd70ce3865dffaec` `chunk_id=srcchunk_ee2d5ba0704f9437b7e1dca911ca8af7` `native_locator=slack:C0AL7EKNHDF:1781219392.813279:1781223404.844679` `source_timestamp=2026-06-12T00:16:44Z`
- The fix was deployed on June 11, 2026 at 16:13:10 UTC. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bc7a7c361bc02e27a4c0b3415f22c13e` `source_revision_id=srcrev_d1397bb7a8e4cc2200b4a7bae4da5630` `chunk_id=srcchunk_28bad3f4c0be8dfb4506abbb6f4ac624` `native_locator=slack:C0AL7EKNHDF:1781219392.813279:1781223607.265549` `source_timestamp=2026-06-12T00:20:07Z`

## Sources

- `source_document_id`: `srcdoc_bc7a7c361bc02e27a4c0b3415f22c13e`
- `source_revision_id`: `srcrev_d1397bb7a8e4cc2200b4a7bae4da5630`
