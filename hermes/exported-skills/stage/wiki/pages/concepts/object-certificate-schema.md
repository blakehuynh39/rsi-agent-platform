---
title: "Object Certificate Schema"
type: "concept"
slug: "concepts/object-certificate-schema"
freshness: "2025-05-14T01:27:00Z"
tags:
  - "certificate"
  - "data-availability"
  - "storage-network"
owners: []
source_revision_ids:
  - "srcrev_15c0ca638218a6c8d679d5ca09ddac18"
conflict_state: "none"
---

# Object Certificate Schema

## Summary

Defines the schema, canonical encoding, signing, and on-chain verification of object certificates for proof of storage on Story Network.

## Claims

- The Object Certificate schema proves that a storage node holds a specific object until a given epoch, signed by the Storage Network private key and validated on-chain by control plane contracts. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`
- The signed fields are: ObjectID ([32]byte, e.g., Merkle-root hash), IpId (string, IP address), Size (uint64, bytes), and EndEpoch (uint64, expiry epoch). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`
- An optional JSON Schema requires objectId, ipId, size, endEpoch, with pattern constraints (objectId: 0x + 64 hex chars, ipId: 0x + 40 hex chars) and size minimum 1, endEpoch minimum 0. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`
- Canonical encoding packs fields in strict order: ObjectID (32 bytes), IpId address (20 bytes), Size (8 bytes big-endian), EndEpoch (8 bytes big-endian). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`
- The certificate hash is computed by keccak256 of the packed encoding (Go: crypto.Keccak256(oc.packForHash())). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`
- On-chain digest matches the Go hash using Solidity: msgHash = keccak256(abi.encodePacked(objectId, ipId, size, endEpoch)). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`
- Signing uses secp256k1: SignCertificate(privKey, oc) returns a 65-byte signature (R||S||V) after hashing the certificate. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`
- On-chain verification extracts (r, s, v) from the signature and calls ecrecover(msgHash, v, r, s). `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`
- The CertifyObjectParams struct defines on-chain certification parameters: objectId (bytes32), ipId (address), signature (bytes). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b) `source_document_id=srcdoc_78d234a362734196f0f9fc031eff4410` `source_revision_id=srcrev_15c0ca638218a6c8d679d5ca09ddac18` `chunk_id=srcchunk_915f34d8e26655a6f266824a3a846d55` `native_locator=https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b` `source_timestamp=2025-05-14T01:27:00Z`

## Sources

- `source_document_id`: `srcdoc_78d234a362734196f0f9fc031eff4410`
- `source_revision_id`: `srcrev_15c0ca638218a6c8d679d5ca09ddac18`
- `source_url`: [Notion source](https://www.notion.so/Object-Certificate-Schema-1f1051299a54803389dec3803339186b)
