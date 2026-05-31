---
title: "Postmortem: L1 Testnet Startup Failure (July 2024)"
type: "runbook"
slug: "runbooks/postmortem-l1-testnet-startup-failure-jul-2024"
freshness: "2024-07-24T08:34:00Z"
tags:
  - "app-state-mismatch"
  - "incident"
  - "p2p"
  - "postmortem"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_67a23e4dc691a5d0c0d416759d9e7283"
conflict_state: "none"
---

# Postmortem: L1 Testnet Startup Failure (July 2024)

## Summary

On July 17-18, 2024, the L1 testnet failed to restart after a network reset. The root cause was an external node connecting to the testnet bootnode, causing a network partition and app state mismatch. The incident also revealed issues with binary retention, merge commit policies, and node management.

## Claims

- A network reset was started on July 17, 2024 at 5pm, and the network did not start properly. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- The initial root cause effort involved checking code commits between the current and previous reset, with a plan to roll back to the previous binary. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- Only 5 binaries were kept in the S3 bucket, and the previous binary used by the last reset had been lost. The policy was updated to store 50 binaries going forward. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- Reverting recent code commits was difficult because merge commits contained many small commits. The code merge policy was set to only allow squash merge. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- Using the oldest available binary from July 16th still resulted in network failure. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- The error indicated an app state mismatch. A hypothesis was that another network was interfering with the testnet. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- The devnet's geth node was found connecting to the testnet. The devnet was turned down and a reset was performed, but errors persisted. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- An earlier commit had changed the configs of all node addresses to the same address. Reverting that change and resetting did not resolve the issue. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- The network did not halt at the beginning but always got stuck at around 20-40 blocks. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- A node set up to test fast sync was using the same config as validator nodes but different keys, connecting to both geth and Iliad bootnode. Turning this node down and resetting allowed the network to restart successfully. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- Some nodes could not be started before developers were in the node's tmux session. A workflow fix was implemented, and there are plans to replace tmux with systemd. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- The initial root cause hypothesis is that an outside node connecting to the bootnode caused some testnet nodes to enter state sync mode while others formed a separate network, eventually leading to an app state mismatch when the networks connected. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- Very short term prevention: ensure no external nodes are connecting to bootnodes before resetting the network. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`
- Before public test, one solution could be using a new bootnode every time the network is reset. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005) `source_document_id=srcdoc_462cb3b52ae2b21b3f389ecb321397d7` `source_revision_id=srcrev_67a23e4dc691a5d0c0d416759d9e7283` `chunk_id=srcchunk_3d90cfbc680a49e79ab5fa95d47eb17d` `native_locator=https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005` `source_timestamp=2024-07-24T08:34:00Z`

## Open Questions

- Exact behaviors of the network partition and state sync mode need further analysis and tests to confirm the root cause hypothesis.

## Sources

- `source_document_id`: `srcdoc_462cb3b52ae2b21b3f389ecb321397d7`
- `source_revision_id`: `srcrev_67a23e4dc691a5d0c0d416759d9e7283`
- `source_url`: [Notion source](https://www.notion.so/Postmortem-L1-Testnet-not-able-to-start-6954307212f64ee09e71444ebec59005)
