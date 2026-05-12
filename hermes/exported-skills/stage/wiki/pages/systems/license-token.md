---
title: "License Token"
type: "system"
slug: "systems/license-token"
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

# License Token

## Summary

The License Token contract handles minting, burning, and validation of license tokens for derivatives. It is only callable by the LicensingModule and checks expiration, ownership, and revocation status.

## Claims

- mintLicenseTokens is only callable by LicensingModule. `claim:claim_3_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_8338ac14f916e6a937797b44d752d1bd` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2` `source_timestamp=2024-04-07T11:07:00Z`
- burnLicenseTokens is only callable by LicensingModule. `claim:claim_3_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_8338ac14f916e6a937797b44d752d1bd` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2` `source_timestamp=2024-04-07T11:07:00Z`
- validateLicenseTokensForDerivative performs _isExpiredNow, ownerOf, and isLicenseTokenRevoked (DisputeModule for isIpTagged). `claim:claim_3_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2) `source_document_id=srcdoc_c8db57f233b5bf217dd12f6000ea7d2c` `source_revision_id=srcrev_376bf19160ed5c18e4a14771d31f4d41` `chunk_id=srcchunk_8338ac14f916e6a937797b44d752d1bd` `native_locator=https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc#chunk-2` `source_timestamp=2024-04-07T11:07:00Z`

## Related Pages

- `license-registry`
- `licensing-module`

## Sources

- `source_document_id`: `srcdoc_c8db57f233b5bf217dd12f6000ea7d2c`
- `source_revision_id`: `srcrev_376bf19160ed5c18e4a14771d31f4d41`
- `source_url`: [Notion source](https://www.notion.so/License-PR-33-56e13a8f4cfc4530915fb3d30b717edc)
