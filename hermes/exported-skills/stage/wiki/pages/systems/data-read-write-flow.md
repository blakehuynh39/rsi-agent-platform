---
title: "Data Read/Write Flow, Abstracted"
type: "system"
slug: "systems/data-read-write-flow"
freshness: "2025-08-11T19:06:00Z"
tags:
  - "data-upload"
  - "erasure-coding"
  - "s3-compatible"
  - "story-poseidon"
  - "sui"
  - "walrus"
owners: []
source_revision_ids:
  - "srcrev_0ab75d9b225b193f7925657e7213610f"
conflict_state: "none"
---

# Data Read/Write Flow, Abstracted

## Summary

Describes two data upload flows: the native Walrus flow using Sui storage units and blob registration, and an S3-compatible flow for Data Publishers that leverages Story Poseidon contracts and an SDK with automatic wallet creation and on-chain transactions.

## Claims

- Pay to acquire a storage unit of a specific size on Sui. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- Prepare data by erasure-coding it into slivers and calculating the Merkle root of slivers as the blob ID. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- Register the blob ID in the acquired storage unit on Sui; Sui contracts emit register events with the blob ID. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- Upload slivers to storage nodes; each storage node checks the slivers against registered blob IDs, accepting if the blob ID is registered on Sui and rejecting otherwise to prevent upload attempts without payments. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- Receive a certificate of availability on stored slivers after quorum. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- Data Publisher tops up $ into the billing account. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- DP uploads data to storage nodes via an S3-compatible API layer. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- DP specifies ACL policies for uploaded data. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- SDK creates an AA wallet if not exist. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- SDK erasure-codes data into slivers and calculates the Merkle root. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- SDK auto-signed onchain transaction purchases appropriate storage on Story Poseidon contracts for the EC slivers. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- SDK auto-signed onchain transaction registers the Merkle root to the purchased storage on Story Poseidon contracts. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`
- SDK uploads slivers to a storage node based on consistent hashing. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592) `source_document_id=srcdoc_48303b9ce218e911159c5a65275fe505` `source_revision_id=srcrev_0ab75d9b225b193f7925657e7213610f` `chunk_id=srcchunk_76dc763216cca4cd02595d8d8d69c6ca` `native_locator=https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592` `source_timestamp=2025-08-11T19:06:00Z`

## Open Questions

- The chunk ends abruptly with 'Storage n...'; the full details of the storage node interaction after upload are missing.

## Sources

- `source_document_id`: `srcdoc_48303b9ce218e911159c5a65275fe505`
- `source_revision_id`: `srcrev_0ab75d9b225b193f7925657e7213610f`
- `source_url`: [Notion source](https://www.notion.so/Data-Read-Write-Flow-Abstracted-1df051299a5480ada381d0774eb90592)
