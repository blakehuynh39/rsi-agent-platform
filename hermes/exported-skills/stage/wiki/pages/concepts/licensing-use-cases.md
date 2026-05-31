---
title: "Licensing Use Cases"
type: "concept"
slug: "concepts/licensing-use-cases"
freshness: "2023-12-09T05:40:00Z"
tags:
  - "IPOrg"
  - "licensing"
  - "NFT"
  - "protocol"
  - "UDL"
owners: []
source_revision_ids:
  - "srcrev_04973cc764190ca322cbe264f6cdb203"
conflict_state: "none"
---

# Licensing Use Cases

## Summary

Overview of licensing use cases for the protocol, including licensing frameworks, UDL v1.0, license NFT metadata, IPOrg configuration, derivative IPA creation, and license activation.

## Claims

- Initially, only UDL v1.0 will be available as a licensing framework. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- TermsRepository is renamed to LicensingFrameworkRepo for consistency. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- A Licensing framework is composed of a link to the off-chain licensing text and a set of Parameters with their values. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- Parameters have a Tag (previously named “Category”), example “Derivatives”, and accept values of types: boolean, number (treated like amount in wei), multiple choices (set). The exact way to indicate accepted values is TBD. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- The licensing text lists a set of Licensing Parameters (previously named Terms). Parameters are considered “tagged” in a License if a term of the corresponding category is in its array of terms. The text for each Parameter establishes default behavior in absence of a “tag”, so not all terms need to be present initially. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- A License will be an NFT with metadata fields: licensor, licensee, revoker, textUrl, terms (listing each term Tag and Value), status, isCommercial, parentLicenseId, ipaId. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- Through the licensing module, an IPOrg can select an available licensing framework by configuring a UDL link and an array of available terms. When UDL is selected, all license parameters are selected automatically; the IPOrg can then set some parameters (e.g., Territory, Transferrable). If a parameter is not set, the licensor has freedom to configure it. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- After configuration, Licenses can be created by the LicensingModule, which will mint License NFTs. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- The IPOrg can choose who can create licenses: always the IPOrg owner, or the owner of an IPA (requires IPA id, assumes no parent license), or the licensee of a parent License (for sublicenses). This choice is a step before UDL in the Licensing Module. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- When creating a Derivative IPA, if the parent License had Share Alike activated and the creator has an existing IPA, the new LNFT is linked to that IPA (previously the LNFT was burned and the data structure became bound; now only linking occurs). After linking, the license status changes to Used. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- A Licensee with an inactive License can activate it. Some tagged Parameters like Allowed-With-Approval cause the license to start deactivated. Activation methods exist and the LNFT has status fields; adding the Parameter may be needed. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_84d1d31832b9f95ef27921fa0ca0c23e` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-1` `source_timestamp=2023-12-09T05:40:00Z`
- When a new license is minted, it copies the parent’s terms, except that the licensor of the new license is the licensee (owner) of the previous one. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-2) `source_document_id=srcdoc_1df10f8eb0208d159b4aa524605c7ede` `source_revision_id=srcrev_04973cc764190ca322cbe264f6cdb203` `chunk_id=srcchunk_7e95b83102cd8b75af3bd9dde62ae502` `native_locator=https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6#chunk-2` `source_timestamp=2023-12-09T05:40:00Z`

## Open Questions

- How to indicate which values a Licensing Parameter accepts? TBD, probably a library.

## Sources

- `source_document_id`: `srcdoc_1df10f8eb0208d159b4aa524605c7ede`
- `source_revision_id`: `srcrev_04973cc764190ca322cbe264f6cdb203`
- `source_url`: [Notion source](https://www.notion.so/Licensing-use-cases-2ffe73008f494753bf2e3a70108330c6)
