---
title: "Licensing Module"
type: "system"
slug: "systems/licensing-module"
freshness: "2024-04-07T11:07:00Z"
tags:
  - "licensing"
  - "protocol-core-v1"
  - "smart-contract"
owners: []
source_revision_ids:
  - "srcrev_376bf19160ed5c18e4a14771d31f4d41"
conflict_state: "none"
---

# Licensing Module

## Summary

The Licensing Module handles attaching license terms to IP, minting license tokens, and registering derivative IP. It enforces permissions, dispute status, and expiration checks.

## Claims

- attachLicenseTerms uses verifyPermission modifier (IPAccountRegistry for isIpAccount, conditional AccessController for checkPermission), _verifyIpNotDisputed (DisputeModule for isIpTagged), LicenseRegistry for attachLicenseTermsToIp (with _exists, _isDerivativeIp, _isExpiredNow checks), sets license template, and adds license terms ID to IP's list. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-1) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_e604d8152f08d021995ccf80973b48eb` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-1` `source_timestamp=2024-04-07T11:07:00Z`
- mintLicenseTokens calls _verifyIpNotDisputed, _hasPermission (IPAccountRegistry isIpAccount, conditional AccessController checkPermission), LicenseRegistry verifyMintLicenseToken (_isExpiredNow, _exists, _hasIpAttachedLicenseTerms, _getMintingLicenseConfig), conditional HookModule verify, _payMintingFee (LicenseTemplate getRoyaltyPolicy, conditional RoyaltyModule onLicenseMinting, conditional _getTotalMintingFee with MintingFeeModule getMintingFee, conditional RoyaltyModule payLicensingMintingFee), LicenseTemplate verifyMintLicenseToken, and LicenseToken mintLicenseTokens (isLicenseTransferable, getExpireTime, _mint ERC-721Enumerable). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-1) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_e604d8152f08d021995ccf80973b48eb` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-1` `source_timestamp=2024-04-07T11:07:00Z`
- registerDerivative uses verifyPermission modifier (IPAccountRegistry isIpAccount, conditional AccessController checkPermission). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-1) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_e604d8152f08d021995ccf80973b48eb` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-1` `source_timestamp=2024-04-07T11:07:00Z`

## Related Pages

- `license-registry`
- `license-token`

## Sources

- `source_document_id`: `srcdoc_c8db57f233b5bf217dd12f6000ea7d2c`
- `source_revision_id`: `srcrev_376bf19160ed5c18e4a14771d31f4d41`
- `source_url`: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc)
