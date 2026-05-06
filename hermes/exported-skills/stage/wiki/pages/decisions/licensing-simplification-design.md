---
title: "Simplify Licensing Design"
type: "decision"
slug: "decisions/licensing-simplification-design"
freshness: "2026-05-05T06:42:42Z"
tags:
  - "design"
  - "licensing"
  - "mainnet"
  - "simplification"
owners: []
source_revision_ids:
  - "srcrev_9fa42b077e32038c88df480d3bd50ccf"
conflict_state: "none"
---

# Simplify Licensing Design

## Summary

Design to simplify the licensing structure from License → Policy → FrameworkData → Framework to License → LicenseTemplate, update interfaces, and fully support third-party LicenseTemplates.

## Claims

- Simplify licensing concepts from License → Policy → FrameworkData → Framework to License → LicenseTemplate. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-1) `source_document_id=srcdoc_475f7ffc7360ac82b8d3c0f652193a23` `source_revision_id=srcrev_9fa42b077e32038c88df480d3bd50ccf` `chunk_id=srcchunk_6d7591db179bcd5eca9bf0ea86ef61b3` `native_locator=https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-1` `source_timestamp=2026-05-05T06:42:42Z`
- Simplify interfaces to use only license-related functions: attachLicense, mintLicenseToken, registerDerivative, registerDerivativeWithToken. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-1) `source_document_id=srcdoc_475f7ffc7360ac82b8d3c0f652193a23` `source_revision_id=srcrev_9fa42b077e32038c88df480d3bd50ccf` `chunk_id=srcchunk_6d7591db179bcd5eca9bf0ea86ef61b3` `native_locator=https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-1` `source_timestamp=2026-05-05T06:42:42Z`
- No need to encode large policy struct to bytes and pass between contracts. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-1) `source_document_id=srcdoc_475f7ffc7360ac82b8d3c0f652193a23` `source_revision_id=srcrev_9fa42b077e32038c88df480d3bd50ccf` `chunk_id=srcchunk_6d7591db179bcd5eca9bf0ea86ef61b3` `native_locator=https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-1` `source_timestamp=2026-05-05T06:42:42Z`
- Fully support third-party LicenseTemplate with permissionless register and access, and IP owner can choose which LicenseTemplate to use. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-1) `source_document_id=srcdoc_475f7ffc7360ac82b8d3c0f652193a23` `source_revision_id=srcrev_9fa42b077e32038c88df480d3bd50ccf` `chunk_id=srcchunk_6d7591db179bcd5eca9bf0ea86ef61b3` `native_locator=https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-1` `source_timestamp=2026-05-05T06:42:42Z`
- LicenseTemplate interface includes functions: name, getLicenseString, getMetadataURI, exists, isTransferable, getExpireTime, verifyMintLicenseToken, verifyRegisterDerivative, verifyCompatibleLicenses. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-2) `source_document_id=srcdoc_475f7ffc7360ac82b8d3c0f652193a23` `source_revision_id=srcrev_9fa42b077e32038c88df480d3bd50ccf` `chunk_id=srcchunk_e3e49e0ac99731941d6591b68e4404ab` `native_locator=https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-2` `source_timestamp=2026-05-05T06:42:42Z`
- When registerDerivative, license activation time is now, IPA expired timestamp is computed from licenseTemplate.expiredAt and stored in IPA storage. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-3) `source_document_id=srcdoc_475f7ffc7360ac82b8d3c0f652193a23` `source_revision_id=srcrev_9fa42b077e32038c88df480d3bd50ccf` `chunk_id=srcchunk_b6c0314917dd0763414eceb91a5c1aad` `native_locator=https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-3` `source_timestamp=2026-05-05T06:42:42Z`
- When minting license token, store license issue time in NFT, compute expiration via licenseTemplate.expireAt and store LicenseToken expired timestamp. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-3) `source_document_id=srcdoc_475f7ffc7360ac82b8d3c0f652193a23` `source_revision_id=srcrev_9fa42b077e32038c88df480d3bd50ccf` `chunk_id=srcchunk_b6c0314917dd0763414eceb91a5c1aad` `native_locator=https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-3` `source_timestamp=2026-05-05T06:42:42Z`
- StoryProtocol defines a Metadata Standard for License Template that all third-party templates should follow, with fields: title, description, TextURL, authors, version, TemplateTerms. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-3) `source_document_id=srcdoc_475f7ffc7360ac82b8d3c0f652193a23` `source_revision_id=srcrev_9fa42b077e32038c88df480d3bd50ccf` `chunk_id=srcchunk_b6c0314917dd0763414eceb91a5c1aad` `native_locator=https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b#chunk-3` `source_timestamp=2026-05-05T06:42:42Z`

## Sources

- `source_document_id`: `srcdoc_475f7ffc7360ac82b8d3c0f652193a23`
- `source_revision_id`: `srcrev_9fa42b077e32038c88df480d3bd50ccf`
- `source_url`: [Notion source](https://www.notion.so/Simplify-Licensing-Design-fc5513aefc824fe4a48480e82a01450b)
