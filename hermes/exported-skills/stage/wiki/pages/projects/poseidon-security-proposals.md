---
title: "Poseidon Security Proposals"
type: "project"
slug: "projects/poseidon-security-proposals"
freshness: "2025-08-11T18:36:00Z"
tags:
  - "ip-protection"
  - "lit-protocol"
  - "mcp"
  - "security"
  - "strifed"
  - "threat-modeling"
  - "walrus"
owners: []
source_revision_ids:
  - "srcrev_917b726baebb671738d13d4aec8a91e0"
conflict_state: "none"
---

# Poseidon Security Proposals

## Summary

A collection of security proposals and threat models for the Poseidon project, covering storage, key management, compliance, MCP oracles, and IP protection.

## Claims

- The STRIFED framework (Spoofing, Tampering, Repudiation, Information Disclosure, Financial Theft/Loss, Elevation of Privilege, Denial of Service) is used for threat modeling. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_30934dc676fdb6408daf902dd2a6d80e` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1` `source_timestamp=2025-08-11T18:36:00Z`
- Walrus mitigates spoofing by requiring nodes to stake WAL tokens and registering identities on-chain. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_30934dc676fdb6408daf902dd2a6d80e` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1` `source_timestamp=2025-08-11T18:36:00Z`
- Walrus uses Merkle roots stored on-chain to detect tampering of erasure-coded slivers. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_30934dc676fdb6408daf902dd2a6d80e` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1` `source_timestamp=2025-08-11T18:36:00Z`
- Walrus provides slashing conditions and Proof-of-Availability mechanisms to prevent repudiation by storage nodes. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_30934dc676fdb6408daf902dd2a6d80e` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1` `source_timestamp=2025-08-11T18:36:00Z`
- Walrus blobs are public by default; confidentiality requires client-side encryption before upload. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_30934dc676fdb6408daf902dd2a6d80e` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-1` `source_timestamp=2025-08-11T18:36:00Z`
- Lit Protocol nodes operate within Trusted Execution Environments (TEEs) to protect key shares and computations. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-2) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_737db8ee571d1c1136753a3f32bf7a34` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-2` `source_timestamp=2025-08-11T18:36:00Z`
- Lit Protocol's backup and recovery uses a Recovery Party requiring a quorum of more than two-thirds to decrypt backups, with an additional Blinder encryption layer per node. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-2) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_737db8ee571d1c1136753a3f32bf7a34` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-2` `source_timestamp=2025-08-11T18:36:00Z`
- A proposed MCP Oracle subnet security measure includes using agent-controlled multisigs with a main driver agent and additional signers running different LLMs sandboxed from external inputs. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-5) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_1492a6425a40bed7f6a308a7aac517d6` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-5` `source_timestamp=2025-08-11T18:36:00Z`
- IP protection proposals include adding 'poison pills' to data, such as using HarmonyCloak to embed imperceptible noise that confounds AI model training. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-5) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_1492a6425a40bed7f6a308a7aac517d6` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-5` `source_timestamp=2025-08-11T18:36:00Z`
- Economic security evaluation must consider tokenomics and value flows to identify risks such as black swans, price death spirals, hyperinflation, excessive sell pressure, governance attacks, excessive farming, and MEV attacks. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-3) `source_document_id=srcdoc_70c8068ac7509089b9d37a81857a84de` `source_revision_id=srcrev_917b726baebb671738d13d4aec8a91e0` `chunk_id=srcchunk_c32ca8eeca5853ccd316f3a644ce0eca` `native_locator=https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e#chunk-3` `source_timestamp=2025-08-11T18:36:00Z`

## Open Questions

- Can expiring licenses produce expiring keys?
- Can we use this system to encrypt data in other storage layers? (private IPA metadata in L1 stored in IPFS)

## Related Pages

- `depin-and-iot-threat-modeling`
- `web2-infra-threat-modeling`

## Sources

- `source_document_id`: `srcdoc_70c8068ac7509089b9d37a81857a84de`
- `source_revision_id`: `srcrev_917b726baebb671738d13d4aec8a91e0`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-Security-Proposals-1d8051299a54801cbd22fdfd2cf9ad2e)
