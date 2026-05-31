---
title: "Running a Node"
type: "runbook"
slug: "runbooks/running-a-node"
freshness: "2024-04-15T09:00:00Z"
tags:
  - "docker"
  - "infrastructure"
  - "node"
  - "p2p"
  - "rpc"
owners: []
source_revision_ids:
  - "srcrev_e275e8d41129fd433107d6292f235ca7"
conflict_state: "none"
---

# Running a Node

## Summary

Instructions for running a story network node locally, including Docker images, RPC endpoints, and P2P sync IPs.

## Claims

- A GitHub repository (story-test-replica) is available for spinning up a node locally. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7) `source_document_id=srcdoc_4c067f9fb2d21db7cea778ca3298d46e` `source_revision_id=srcrev_e275e8d41129fd433107d6292f235ca7` `chunk_id=srcchunk_386c54da1ac9cf4d070418ee74c846b2` `native_locator=https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7` `source_timestamp=2024-04-15T09:00:00Z`
- Syncing to the latest block usually takes a few hours. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7) `source_document_id=srcdoc_4c067f9fb2d21db7cea778ca3298d46e` `source_revision_id=srcrev_e275e8d41129fd433107d6292f235ca7` `chunk_id=srcchunk_386c54da1ac9cf4d070418ee74c846b2` `native_locator=https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7` `source_timestamp=2024-04-15T09:00:00Z`
- Docker image for op-node: public.ecr.aws/i6b2w2n6/op-node:celestia-5.0.0 `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7) `source_document_id=srcdoc_4c067f9fb2d21db7cea778ca3298d46e` `source_revision_id=srcrev_e275e8d41129fd433107d6292f235ca7` `chunk_id=srcchunk_386c54da1ac9cf4d070418ee74c846b2` `native_locator=https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7` `source_timestamp=2024-04-15T09:00:00Z`
- Docker image for op-geth: public.ecr.aws/i6b2w2n6/op-geth:celestia-5.0.0 `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7) `source_document_id=srcdoc_4c067f9fb2d21db7cea778ca3298d46e` `source_revision_id=srcrev_e275e8d41129fd433107d6292f235ca7` `chunk_id=srcchunk_386c54da1ac9cf4d070418ee74c846b2` `native_locator=https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7` `source_timestamp=2024-04-15T09:00:00Z`
- Existing HTTP RPC endpoint: https://story-network.rpc.caldera.xyz/infra-partner-http `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7) `source_document_id=srcdoc_4c067f9fb2d21db7cea778ca3298d46e` `source_revision_id=srcrev_e275e8d41129fd433107d6292f235ca7` `chunk_id=srcchunk_386c54da1ac9cf4d070418ee74c846b2` `native_locator=https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7` `source_timestamp=2024-04-15T09:00:00Z`
- Existing WebSocket endpoint: wss://story-network.rpc.caldera.xyz/infra-partner-ws `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7) `source_document_id=srcdoc_4c067f9fb2d21db7cea778ca3298d46e` `source_revision_id=srcrev_e275e8d41129fd433107d6292f235ca7` `chunk_id=srcchunk_386c54da1ac9cf4d070418ee74c846b2` `native_locator=https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7` `source_timestamp=2024-04-15T09:00:00Z`
- P2P Sync IPs: 52.11.228.10, 54.149.241.249, 54.184.163.29. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7) `source_document_id=srcdoc_4c067f9fb2d21db7cea778ca3298d46e` `source_revision_id=srcrev_e275e8d41129fd433107d6292f235ca7` `chunk_id=srcchunk_386c54da1ac9cf4d070418ee74c846b2` `native_locator=https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7` `source_timestamp=2024-04-15T09:00:00Z`

## Sources

- `source_document_id`: `srcdoc_4c067f9fb2d21db7cea778ca3298d46e`
- `source_revision_id`: `srcrev_e275e8d41129fd433107d6292f235ca7`
- `source_url`: [Notion source](https://www.notion.so/Running-Node-0ca966f7008f4149a3a16d4c828c0ef7)
