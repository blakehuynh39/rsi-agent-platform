---
title: "License Registry"
type: "system"
slug: "systems/license-registry"
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

# License Registry

## Summary

The License Registry is responsible for attaching license terms to IP and verifying mint license tokens. It performs existence, derivative, expiration, attachment, and minting config checks.

## Claims

- attachLicenseTermsToIp performs _exists (checks LicenseTemplate is registered), _isDerivativeIp (checks length of parent IPs > 0), and _isExpiredNow (gets EXPIRATION_TIME from IPAccount storage, checks expired). `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-1) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_e604d8152f08d021995ccf80973b48eb` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-1` `source_timestamp=2024-04-07T11:07:00Z`
- verifyMintLicenseToken performs _isExpiredNow, _exists, _hasIpAttachedLicenseTerms (checks if license term is default NC social remix or attached to IP), and _getMintingLicenseConfig (checks and returns minting license config). `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_8338ac14f916e6a937797b44d752d1bd` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2` `source_timestamp=2024-04-07T11:07:00Z`
- registerDerivativeIp is only callable by LicensingModule. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_8338ac14f916e6a937797b44d752d1bd` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2` `source_timestamp=2024-04-07T11:07:00Z`

## Related Pages

- `license-token`
- `licensing-module`

## Sources

- `source_document_id`: `srcdoc_c8db57f233b5bf217dd12f6000ea7d2c`
- `source_revision_id`: `srcrev_376bf19160ed5c18e4a14771d31f4d41`
- `source_url`: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc)
