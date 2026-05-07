---
title: "EIP-7212 Precompile"
type: "decision"
slug: "decisions/eip-7212-precompile"
freshness: "2024-02-18T06:33:00Z"
tags:
  - "account-abstraction"
  - "cryptography"
  - "EIP-7212"
  - "precompile"
  - "secp256r1"
owners: []
source_revision_ids:
  - "srcrev_8f437a891c2fbf458639daab16872581"
conflict_state: "none"
---

# EIP-7212 Precompile

## Summary

Analysis and recommendation for adding a secp256r1 precompile to the Renaissance chain to enable native support for widely-used secure authentication technologies like Apple Secure Enclave, Android Keystore, and WebAuthn, primarily to supercharge Account Abstraction (AA) user operations.

## Claims

- Ethereum uses the secp256k1 curve for signature verification and key generation, while secp256r1 is more widely used in mainstream technology and recommended by NIST. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-1) `source_document_id=srcdoc_2c35ee3d01c0d187b03a5e9e06489385` `source_revision_id=srcrev_8f437a891c2fbf458639daab16872581` `chunk_id=srcchunk_90b04aa470f3aca1782648b11671f290` `native_locator=https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-1` `source_timestamp=2024-02-18T06:33:00Z`
- secp256r1 is used in Android keystore, Apple's secure enclave, and WebAuthn, enabling native biometric authentication without exposing private keys. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-1) `source_document_id=srcdoc_2c35ee3d01c0d187b03a5e9e06489385` `source_revision_id=srcrev_8f437a891c2fbf458639daab16872581` `chunk_id=srcchunk_90b04aa470f3aca1782648b11671f290` `native_locator=https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-1` `source_timestamp=2024-02-18T06:33:00Z`
- The primary use case for the secp256r1 precompile is supercharging Account Abstraction (AA) by allowing UserOperations to be validated with secp256r1 signatures, enabling biometric auth flows. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-1) `source_document_id=srcdoc_2c35ee3d01c0d187b03a5e9e06489385` `source_revision_id=srcrev_8f437a891c2fbf458639daab16872581` `chunk_id=srcchunk_90b04aa470f3aca1782648b11671f290` `native_locator=https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-1` `source_timestamp=2024-02-18T06:33:00Z`
- Adding the secp256r1 precompile is straightforward due to the mature and well-defined algorithm, and it should be requested from Caldera at a dedicated address with similar gas costs to ecRecover. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-2) `source_document_id=srcdoc_2c35ee3d01c0d187b03a5e9e06489385` `source_revision_id=srcrev_8f437a891c2fbf458639daab16872581` `chunk_id=srcchunk_805ea8a52a466e7310b38cc14c56de1c` `native_locator=https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-2` `source_timestamp=2024-02-18T06:33:00Z`
- Supporting secp256r1 for EOA creation is a secondary option that would allow native secure key generation on Apple/Android devices, but adoption may be hindered by limited library support (e.g., ethers.js). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-2) `source_document_id=srcdoc_2c35ee3d01c0d187b03a5e9e06489385` `source_revision_id=srcrev_8f437a891c2fbf458639daab16872581` `chunk_id=srcchunk_805ea8a52a466e7310b38cc14c56de1c` `native_locator=https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-2` `source_timestamp=2024-02-18T06:33:00Z`
- EIP-7560 (native AA transaction support) is considered a large infrastructure lift and should be pursued later due to early maturity and uncertain adoption. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-2) `source_document_id=srcdoc_2c35ee3d01c0d187b03a5e9e06489385` `source_revision_id=srcrev_8f437a891c2fbf458639daab16872581` `chunk_id=srcchunk_805ea8a52a466e7310b38cc14c56de1c` `native_locator=https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e#chunk-2` `source_timestamp=2024-02-18T06:33:00Z`

## Open Questions

- How big of a lift would it be for Caldera to support secp256r1 EOA generation alongside secp256k1?
- What is Caldera's assessment of the infrastructure lift for EIP-7560 native AA transaction support?

## Sources

- `source_document_id`: `srcdoc_2c35ee3d01c0d187b03a5e9e06489385`
- `source_revision_id`: `srcrev_8f437a891c2fbf458639daab16872581`
- `source_url`: [Notion source](https://www.notion.so/EIP-7212-Precompile-c26ad7d04d634c9bab1d9c5b0b84fb3e)
