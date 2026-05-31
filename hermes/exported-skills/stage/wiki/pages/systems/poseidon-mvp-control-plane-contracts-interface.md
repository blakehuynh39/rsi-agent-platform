---
title: "Poseidon MVP Control Plane Contracts Interface"
type: "system"
slug: "systems/poseidon-mvp-control-plane-contracts-interface"
freshness: "2025-08-11T19:42:00Z"
tags:
  - "control-plane"
  - "mvp"
  - "poseidon"
  - "smart-contracts"
  - "storage"
owners: []
source_revision_ids:
  - "srcrev_2a64efcb96db10368964076a9bf168c8"
conflict_state: "none"
---

# Poseidon MVP Control Plane Contracts Interface

## Summary

Defines the on-chain control plane contracts for the Poseidon MVP, including terminology, architecture, key workflows (upload object, advance epoch), and the Solidity interface IControlPlane with structs and functions for managing storage reservations, objects, and epochs.

## Claims

- An IP Asset is a Story Protocol IP Asset. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- A Storage reservation describes reserved storage space for a defined period (startEpoch to endEpoch) and users may extend the reservation duration and/or increase capacity size. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- An Object is an individual S3 object (file) stored within a Storage, and a single Storage can contain one or multiple Object instances. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- A Bucket is an S3 bucket used to organize Objects under a folder-like namespace, named with the IP Asset address, and may hold multiple Objects. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- The Storage Network is the decentralized infrastructure providing physical space, triggered by on-chain control plane events. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- All data related to an IP Asset is stored in its dedicated S3 bucket named with the IP Asset address. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- A Storage reservation can host one or more Objects and may span multiple IP Assets if reused. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- Only the Storage owner can upload data into that Storage slot; only the IP Asset owner can upload Objects of the given IP. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- The expiration of an Object is governed by the endEpoch of its parent Storage reservation. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- The data license is managed by the IP Asset. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_186cb21325f4b5a4abd9319d85a7f947` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-1` `source_timestamp=2025-08-11T19:42:00Z`
- One Object must belong to one Storage, and one Storage might store one or multiple Objects. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-2) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_60336b9ca422871a837a130eb482ade1` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-2` `source_timestamp=2025-08-11T19:42:00Z`
- One Storage might store Objects/Files of multiple IPs. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-2) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_60336b9ca422871a837a130eb482ade1` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-2` `source_timestamp=2025-08-11T19:42:00Z`
- Only the Storage owner can upload Objects to the Storage, and only the IP owner can upload Objects associated with the IP. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-2) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_60336b9ca422871a837a130eb482ade1` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-2` `source_timestamp=2025-08-11T19:42:00Z`
- The expire time of an Object is determined by the end epoch of the Storage in which the Object is stored. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-2) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_60336b9ca422871a837a130eb482ade1` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-2` `source_timestamp=2025-08-11T19:42:00Z`
- To advance an epoch, one must prepare an EpochConfig with totalCapacitySize, storagePrice, and storagePubKey, then call advanceEpoch(epochConfig) which increments the epoch counter and applies updated capacity and pricing rules. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-3) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_e98ae5ef632966776a741ec6404fe0b4` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-3` `source_timestamp=2025-08-11T19:42:00Z`
- The IControlPlane interface defines structs for Object, RegisterObjectParams, CertifyObjectParams, and Storage, and includes functions such as deleteObjects and advanceEpoch. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-3) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_e98ae5ef632966776a741ec6404fe0b4` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-3` `source_timestamp=2025-08-11T19:42:00Z`
- The Object struct includes fields: id (Merkle-root hash), registeredEpoch, size, certificatedEpoch, storageId, ipId, and deletable. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-3) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_e98ae5ef632966776a741ec6404fe0b4` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-3` `source_timestamp=2025-08-11T19:42:00Z`
- The Storage struct includes fields: id, startEpoch, endEpoch, capacitySize, usedSize, and owner. `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-3) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_e98ae5ef632966776a741ec6404fe0b4` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-3` `source_timestamp=2025-08-11T19:42:00Z`
- The deleteObjects function allows deletion of registered objects if allowed, taking an ipId and an array of objectIds. `claim:claim_1_19` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-4) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_37ee5f63a9a724e65b261a0601357782` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-4` `source_timestamp=2025-08-11T19:42:00Z`
- The advanceEpoch function takes an EpochConfig and returns the new epoch number. `claim:claim_1_20` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-4) `source_document_id=srcdoc_a6d1054954d90bbec1ee17a5404d07e0` `source_revision_id=srcrev_2a64efcb96db10368964076a9bf168c8` `chunk_id=srcchunk_37ee5f63a9a724e65b261a0601357782` `native_locator=https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97#chunk-4` `source_timestamp=2025-08-11T19:42:00Z`

## Sources

- `source_document_id`: `srcdoc_a6d1054954d90bbec1ee17a5404d07e0`
- `source_revision_id`: `srcrev_2a64efcb96db10368964076a9bf168c8`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-MVP-Control-Plane-Contracts-Interface-1ea051299a5480c1bc96d3add39e6e97)
