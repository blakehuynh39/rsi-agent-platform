---
title: "IP Asset On-Chain Metadata Design"
type: "decision"
slug: "decisions/ip-asset-on-chain-metadata-design"
freshness: "2023-09-12T03:54:00Z"
tags:
  - "ip-asset"
  - "metadata"
  - "on-chain"
  - "protocol-design"
owners: []
source_revision_ids:
  - "srcrev_349b17446f9107d63f8872308dc176e8"
conflict_state: "none"
---

# IP Asset On-Chain Metadata Design

## Summary

Proposal to move critical IP asset metadata on-chain, defining core static and dynamic attribute primitives for interoperability and composability.

## Claims

- Current IP asset data structure outsources metadata to off-chain storage like IPFS or Arweave, leaving critical fields off-chain. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-1) `source_document_id=srcdoc_9efa560a979244e0d05ef8ba4a4ce6ef` `source_revision_id=srcrev_349b17446f9107d63f8872308dc176e8` `chunk_id=srcchunk_29af031f9060a6688f928d5ed55025a8` `native_locator=https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-1` `source_timestamp=2023-09-12T03:54:00Z`
- Off-chain metadata offers flexibility with JSON structures but forfeits standard on-chain data format, ease of access, and interoperability. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-1) `source_document_id=srcdoc_9efa560a979244e0d05ef8ba4a4ce6ef` `source_revision_id=srcrev_349b17446f9107d63f8872308dc176e8` `chunk_id=srcchunk_29af031f9060a6688f928d5ed55025a8` `native_locator=https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-1` `source_timestamp=2023-09-12T03:54:00Z`
- Common IP asset fields proposed include IP Type, Author, Title/Name, Description, Owner, Royalty Receiver, IP Account, Date of Creation, and Registration/IP Number. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-1) `source_document_id=srcdoc_9efa560a979244e0d05ef8ba4a4ce6ef` `source_revision_id=srcrev_349b17446f9107d63f8872308dc176e8` `chunk_id=srcchunk_29af031f9060a6688f928d5ed55025a8` `native_locator=https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-1` `source_timestamp=2023-09-12T03:54:00Z`
- Registration/Application Date and Expiration Date were proposed but struck through in the source. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-1) `source_document_id=srcdoc_9efa560a979244e0d05ef8ba4a4ce6ef` `source_revision_id=srcrev_349b17446f9107d63f8872308dc176e8` `chunk_id=srcchunk_29af031f9060a6688f928d5ed55025a8` `native_locator=https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-1` `source_timestamp=2023-09-12T03:54:00Z`
- Core IP attribute primitives should be categorized into Static (fixed source of truth, e.g., licensor, licensee, territory) and Dynamic (may evolve, e.g., terms, royalties, sublicensing rights). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-2) `source_document_id=srcdoc_9efa560a979244e0d05ef8ba4a4ce6ef` `source_revision_id=srcrev_349b17446f9107d63f8872308dc176e8` `chunk_id=srcchunk_16746b336223ed866251c91924e921ca` `native_locator=https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-2` `source_timestamp=2023-09-12T03:54:00Z`
- Parametric values must be on-chain for enforcement and role setting; metadata is for illustrative and legal purposes. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-2) `source_document_id=srcdoc_9efa560a979244e0d05ef8ba4a4ce6ef` `source_revision_id=srcrev_349b17446f9107d63f8872308dc176e8` `chunk_id=srcchunk_16746b336223ed866251c91924e921ca` `native_locator=https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-2` `source_timestamp=2023-09-12T03:54:00Z`
- Proposed term areas include General Terms (exclusive, commercial, sublicensable, active), Payment Terms (contract address, author/parent splits), Granting Terms (approval steps, processor contracts), and Time-related Terms (duration, renewable, renovation oracle). `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-2) `source_document_id=srcdoc_9efa560a979244e0d05ef8ba4a4ce6ef` `source_revision_id=srcrev_349b17446f9107d63f8872308dc176e8` `chunk_id=srcchunk_16746b336223ed866251c91924e921ca` `native_locator=https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697#chunk-2` `source_timestamp=2023-09-12T03:54:00Z`

## Open Questions

- How will dynamic attributes be updated and governed on-chain?
- What is the cost model for storing metadata on-chain versus off-chain?
- What is the final agreed set of on-chain IP attribute primitives?

## Sources

- `source_document_id`: `srcdoc_9efa560a979244e0d05ef8ba4a4ce6ef`
- `source_revision_id`: `srcrev_349b17446f9107d63f8872308dc176e8`
- `source_url`: [Notion source](https://www.notion.so/IPA-IP-Assets-rethinking-fff0f90d32ec40379a294af82bd0d697)
