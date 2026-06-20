---
title: "Decision to remove circulation-supply.storyapis.com from Cloudflare"
type: "decision"
slug: "decisions/remove-deprecated-circulation-supply-domain"
freshness: "2026-04-13T06:42:44Z"
tags:
  - "cloudflare"
  - "deprecation"
  - "dns"
  - "domain"
  - "storyapis"
owners:
  - "S083BDZ4FTM"
source_revision_ids:
  - "srcrev_01fa5aeff6abb9f08bfcfe2bf5315913"
  - "srcrev_239bc30529d060f9744dde7bad0fe8d6"
  - "srcrev_3f8a042237021a03658a75e1004f6e78"
  - "srcrev_5521264a7d6bab71857e8689c9b876ff"
  - "srcrev_6ea483a36c02410c64a948d70d93f887"
  - "srcrev_9a79e9e25db76d682d26be3f5b8b97ae"
conflict_state: "none"
---

# Decision to remove circulation-supply.storyapis.com from Cloudflare

## Summary

Approved removal of the deprecated domain circulation-supply.storyapis.com from Cloudflare configuration (PR #176) after confirming successful migration to mainnet-circulation-supply.storyapis.com.

## Claims

- A pull request (https://github.com/piplabs/cloudflare/pull/176) was opened to remove domains from Cloudflare configuration. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_75e489fe59d32ebe875f36aefa01b986` `source_revision_id=srcrev_01fa5aeff6abb9f08bfcfe2bf5315913` `chunk_id=srcchunk_a72da02e1f08c644db9f328d0b6b40c0` `native_locator=slack:C0547N89JUB:1775807898.788999:1775807898.788999` `source_timestamp=2026-04-10T07:58:18Z`
- During review, a question was raised about whether the domain circulation-supply.storyapis.com is still in use. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_75e489fe59d32ebe875f36aefa01b986` `source_revision_id=srcrev_9a79e9e25db76d682d26be3f5b8b97ae` `chunk_id=srcchunk_0d116c91807c336b337d70dc426194a6` `native_locator=slack:C0547N89JUB:1775807898.788999:1775839880.860639` `source_timestamp=2026-04-10T16:51:20Z`
- It was suggested that the correct domain is mainnet-circulation-supply.storyapis.com. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_75e489fe59d32ebe875f36aefa01b986` `source_revision_id=srcrev_6ea483a36c02410c64a948d70d93f887` `chunk_id=srcchunk_1ab17dcd7956adc6d5bec2faa25daca7` `native_locator=slack:C0547N89JUB:1775807898.788999:1775864671.418719` `source_timestamp=2026-04-10T23:44:31Z`
- Confirmed that the migration to mainnet-circulation-supply.storyapis.com was already completed. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_75e489fe59d32ebe875f36aefa01b986` `source_revision_id=srcrev_5521264a7d6bab71857e8689c9b876ff` `chunk_id=srcchunk_0037684fe74957105b3c17be17069214` `native_locator=slack:C0547N89JUB:1775807898.788999:1775869858.965909` `source_timestamp=2026-04-11T01:10:58Z`
- The team approved proceeding with the pull request to remove the deprecated domain, with no objections. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_75e489fe59d32ebe875f36aefa01b986` `source_revision_id=srcrev_3f8a042237021a03658a75e1004f6e78` `chunk_id=srcchunk_7d55b6d9a082d1e2a93dfea6c0921dd0` `native_locator=slack:C0547N89JUB:1775807898.788999:1776041917.312669` `source_timestamp=2026-04-13T00:58:37Z`
  - citation: `source_document_id=srcdoc_75e489fe59d32ebe875f36aefa01b986` `source_revision_id=srcrev_239bc30529d060f9744dde7bad0fe8d6` `chunk_id=srcchunk_d47eda5361406862031174aeb93fdd6c` `native_locator=slack:C0547N89JUB:1775807898.788999:1776062564.914499` `source_timestamp=2026-04-13T06:42:44Z`

## Sources

- `source_document_id`: `srcdoc_75e489fe59d32ebe875f36aefa01b986`
- `source_revision_id`: `srcrev_239bc30529d060f9744dde7bad0fe8d6`
