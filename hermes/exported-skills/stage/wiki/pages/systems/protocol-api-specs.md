---
title: "Protocol API Specs"
type: "system"
slug: "systems/protocol-api-specs"
freshness: "2026-05-05T06:26:25Z"
tags:
  - "api"
  - "protocol"
  - "specification"
owners: []
source_revision_ids:
  - "srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e"
conflict_state: "none"
---

# Protocol API Specs

## Summary

REST API specifications for the protocol, covering franchises, IP assets, licenses, collections, and transactions. Relationship operations are handled entirely via smart contracts with no backend API dependency.

## Claims

- The Franchise Get endpoint is `GET {baseUrl}/franchise/:franchiseId` and returns `franchiseId`, `franchiseName`, `ownerAdderss`, `tokenUri`, `txHash`, and in the next version `arweaveUrl`. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The List Franchises endpoint is `GET {baseUrl}/franchise` and returns an array of objects with `franchiseId`, `franchiseName`, `ownerAdderss`, `tokenUri`. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The Get IP Asset endpoint is `GET {baseUrl}/ipasset/:ipAssetId?franchiseId={franchiseId}` and returns `ipAssetId`, `franchiseId`, `ipAssetName`, `ipAssetType`, `ownerAdderss`, `tokenUri`, `metadata` (JSON format), `txHash`. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The List IP Assets endpoint is `GET {baseUrl}/ipasset?franchiseId=(franchiseId)` and returns an array of objects with `ipAssetId`, `franchiseId`, `ipAssetName`, `ipAssetType`, `ownerAdderss`, `tokenUri`, `metadata` (JSON), `txHash`. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The License data model is defined as `interface License { licenseId: string, ipAssetId: string, franchiseId: string, parentLicenseId: string, licenseOwnerAddress: string, uri: string }`. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The Get License endpoint is `GET {baseUrl}/license/:licenseId` and returns `licenseId`, `ipAssetId`, `franchiseId`, `parentLicenseId`, `ownerAddress`, `uri`. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The List Licenses endpoint is `GET {baseUrl}/license?franchiseId={franchiseId}&ipAssetId={ipAssetId}` and returns an array of license objects. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- Relationship operations are all smart contract calls with no backend API dependency. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The List Collections endpoint is `GET {baseUrl}/collection?franchiseId={franchiseId}` and returns an array with `totalCollected`, `totalCollectors` (next step), `ipAssetId`, `franchiseId`. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The Transaction interface includes `txId`, `txHash`, `createdAt` (ISO 8601), `creatorAddress`, `type` (ResourceType), `resourceId`, `franchiseId`. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The ResourceType enum includes FRANCHISE, IP_ASSET, LICENSE, RELATIONSHIP, COLLECTION. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`
- The Get Transaction endpoint is `GET {baseUrl}/transaction/:transactionId` and returns `txId`, `txHash`, `createdAt` (epoch in seconds), `creatorAddress`, `resourceType`. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8) `source_document_id=srcdoc_09ae78c940e9765f368825da86b5efc3` `source_revision_id=srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e` `chunk_id=srcchunk_9d4b149225e8137fb2705b5b966dab70` `native_locator=https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8` `source_timestamp=2026-05-05T06:26:25Z`

## Open Questions

- How will `totalCollectors` be implemented in the List Collections endpoint?
- What is the exact response structure for the Get Transaction endpoint beyond the fields listed?
- What is the purpose of the commented `arweaveTxHash` and `status` fields in the Transaction interface?
- When will the `arweaveUrl` field be added to the Franchise Get response?

## Sources

- `source_document_id`: `srcdoc_09ae78c940e9765f368825da86b5efc3`
- `source_revision_id`: `srcrev_b7e15f7a4c446b5a71c1e664fb8e7f9e`
- `source_url`: [Notion source](https://www.notion.so/Protocol-API-Specs-d89805de5f59405b88bc5436727ea8a8)
