---
title: "Ethereum Rollups"
type: "concept"
slug: "concepts/ethereum-rollups"
freshness: "2024-09-11T04:23:00Z"
tags:
  - "ethereum"
  - "l2"
  - "rollups"
  - "scaling"
owners: []
source_revision_ids:
  - "srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598"
conflict_state: "none"
---

# Ethereum Rollups

## Summary

Ethereum rollups are a Layer 2 scaling solution that batches transactions off-chain and submits them to the Ethereum L1, reducing fees and increasing throughput.

## Claims

- Ethereum rollups batch transactions to submit to L1 through Calldata to a smart contract (before EIP4844, Cancun update). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1) `source_document_id=srcdoc_54c4859a9299d4382243db02be15bf49` `source_revision_id=srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598` `chunk_id=srcchunk_8099b327a427aa247debbf8711b01842` `native_locator=https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1` `source_timestamp=2024-09-11T04:23:00Z`
- Each batch happens every few minutes and is configured by sequencers. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1) `source_document_id=srcdoc_54c4859a9299d4382243db02be15bf49` `source_revision_id=srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598` `chunk_id=srcchunk_8099b327a427aa247debbf8711b01842` `native_locator=https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1` `source_timestamp=2024-09-11T04:23:00Z`
- Transaction finality time is the same as the base layer Ethereum. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1) `source_document_id=srcdoc_54c4859a9299d4382243db02be15bf49` `source_revision_id=srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598` `chunk_id=srcchunk_8099b327a427aa247debbf8711b01842` `native_locator=https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1` `source_timestamp=2024-09-11T04:23:00Z`
- State verification varies depending on the proof system used. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1) `source_document_id=srcdoc_54c4859a9299d4382243db02be15bf49` `source_revision_id=srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598` `chunk_id=srcchunk_8099b327a427aa247debbf8711b01842` `native_locator=https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1` `source_timestamp=2024-09-11T04:23:00Z`
- Transactions are compressed to reduce gas fees, and fees are much lower due to transaction batching. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1) `source_document_id=srcdoc_54c4859a9299d4382243db02be15bf49` `source_revision_id=srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598` `chunk_id=srcchunk_8099b327a427aa247debbf8711b01842` `native_locator=https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1` `source_timestamp=2024-09-11T04:23:00Z`
- Transactions happen on a sequencer which orders them; there is no public mempool, the sequencer has a private mempool. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1) `source_document_id=srcdoc_54c4859a9299d4382243db02be15bf49` `source_revision_id=srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598` `chunk_id=srcchunk_8099b327a427aa247debbf8711b01842` `native_locator=https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1` `source_timestamp=2024-09-11T04:23:00Z`
- Proof systems used include fraud proofs, validity proofs, and zk proofs. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1) `source_document_id=srcdoc_54c4859a9299d4382243db02be15bf49` `source_revision_id=srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598` `chunk_id=srcchunk_8099b327a427aa247debbf8711b01842` `native_locator=https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e#chunk-1` `source_timestamp=2024-09-11T04:23:00Z`

## Related Pages

- `modular-blockchain-layers`
- `optimism-deposit-flow`

## Sources

- `source_document_id`: `srcdoc_54c4859a9299d4382243db02be15bf49`
- `source_revision_id`: `srcrev_0ab04c2a0fcd1b2ea5d05ffe7bbb0598`
- `source_url`: [Notion source](https://www.notion.so/L2-L1-d329b413f4f7436d88bbfdaee1a7fb8e)
