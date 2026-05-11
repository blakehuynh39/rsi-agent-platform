---
title: "Protocol v1.0.0-rc.2 Indexing"
type: "system"
slug: "systems/protocol-v1-0-0-rc-2-indexing"
freshness: "2024-04-09T04:21:00Z"
tags:
  - "indexing"
  - "protocol"
  - "release-candidate"
owners: []
source_revision_ids:
  - "srcrev_1e55eac18b5bb8e806b3f0e987f10c36"
conflict_state: "none"
---

# Protocol v1.0.0-rc.2 Indexing

## Summary

Contract addresses and entity schema changes for the Protocol v1.0.0-rc.2 release candidate indexing.

## Claims

- The AccessController contract address is 0x7e253Df9b0fC872746877Fa362b2cAf32712d770. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-1) `source_document_id=srcdoc_1b16421b8461a0d9030a8f18d11cbcaf` `source_revision_id=srcrev_1e55eac18b5bb8e806b3f0e987f10c36` `chunk_id=srcchunk_bf9c326839e76f2bc7c4abe7478aeb09` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-1` `source_timestamp=2024-04-09T04:21:00Z`
- The IPAPolicy entity is renamed to IPLicenseTerm. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-1) `source_document_id=srcdoc_1b16421b8461a0d9030a8f18d11cbcaf` `source_revision_id=srcrev_1e55eac18b5bb8e806b3f0e987f10c36` `chunk_id=srcchunk_bf9c326839e76f2bc7c4abe7478aeb09` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-1` `source_timestamp=2024-04-09T04:21:00Z`
- The handlePolicyRegistered handler is replaced by handleLicenseTermsRegistered, which listens to the LicenseTermsRegistered event and calls the licenseTemplate to get JSON data. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-2) `source_document_id=srcdoc_1b16421b8461a0d9030a8f18d11cbcaf` `source_revision_id=srcrev_1e55eac18b5bb8e806b3f0e987f10c36` `chunk_id=srcchunk_fd49ed4fe29e97dd682fda16463a80b5` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-2` `source_timestamp=2024-04-09T04:21:00Z`
- The handleIpIdLinkedToParents handler is replaced by handleDerivativeRegistered, which uses the DerivativeRegistered event from the LicensingModule contract. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-2) `source_document_id=srcdoc_1b16421b8461a0d9030a8f18d11cbcaf` `source_revision_id=srcrev_1e55eac18b5bb8e806b3f0e987f10c36` `chunk_id=srcchunk_fd49ed4fe29e97dd682fda16463a80b5` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-2` `source_timestamp=2024-04-09T04:21:00Z`
- The License entity is renamed to LicenseToken, with fields including licensorIpId, licenseTemplate, licenseTermsId, transferable, owner, mintedAt, expiresAt, burntAt. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-1) `source_document_id=srcdoc_1b16421b8461a0d9030a8f18d11cbcaf` `source_revision_id=srcrev_1e55eac18b5bb8e806b3f0e987f10c36` `chunk_id=srcchunk_bf9c326839e76f2bc7c4abe7478aeb09` `native_locator=https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa#chunk-1` `source_timestamp=2024-04-09T04:21:00Z`

## Sources

- `source_document_id`: `srcdoc_1b16421b8461a0d9030a8f18d11cbcaf`
- `source_revision_id`: `srcrev_1e55eac18b5bb8e806b3f0e987f10c36`
- `source_url`: [Notion source](https://www.notion.so/Protocol-v1-0-0-rc-2-indexing-4997e50147c646289656c93a81ed1ffa)
