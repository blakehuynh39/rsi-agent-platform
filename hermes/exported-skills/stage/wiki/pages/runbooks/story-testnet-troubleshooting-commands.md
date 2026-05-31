---
title: "Story Testnet Troubleshooting Commands"
type: "runbook"
slug: "runbooks/story-testnet-troubleshooting-commands"
freshness: "2024-10-04T18:37:00Z"
tags:
  - "rpc"
  - "story-protocol"
  - "systemd"
  - "testnet"
  - "troubleshooting"
owners: []
source_revision_ids:
  - "srcrev_2c78dfd508c1981220a3d08e48797cf7"
conflict_state: "none"
---

# Story Testnet Troubleshooting Commands

## Summary

A collection of commands for troubleshooting Story Protocol testnet nodes, including checking genesis hash, process status, network resources, latest block number, account balance, and systemd service status.

## Claims

- The genesis hash for the Story testnet can be retrieved by calling eth_getBlockByNumber with block number 0x0 on the testnet RPC endpoint https://testnet.storyrpc.io, which returns 0xf688549151cee34b707abd49c32c019aefb766d701488dd0c668601a91a67978. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-1) `source_document_id=srcdoc_b96b092961e1dcf7152abe43e62ddec8` `source_revision_id=srcrev_2c78dfd508c1981220a3d08e48797cf7` `chunk_id=srcchunk_8a2e17db3a6e2bcec408dd2ec48d90c4` `native_locator=https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-1` `source_timestamp=2024-10-04T18:37:00Z`
- The status of the Story testnet node processes can be checked using systemctl status node-geth and systemctl status node-story. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-1) `source_document_id=srcdoc_b96b092961e1dcf7152abe43e62ddec8` `source_revision_id=srcrev_2c78dfd508c1981220a3d08e48797cf7` `chunk_id=srcchunk_8a2e17db3a6e2bcec408dd2ec48d90c4` `native_locator=https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-1` `source_timestamp=2024-10-04T18:37:00Z`
- Network resources for the Story testnet can be found by checking the latest successful workflow run in the GitHub Actions tab of the node-launcher repository, specifically the 'Prepare the ansible inventory file' step in the deploy job, which lists all servers per role. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-1) `source_document_id=srcdoc_b96b092961e1dcf7152abe43e62ddec8` `source_revision_id=srcrev_2c78dfd508c1981220a3d08e48797cf7` `chunk_id=srcchunk_8a2e17db3a6e2bcec408dd2ec48d90c4` `native_locator=https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-1` `source_timestamp=2024-10-04T18:37:00Z`
- The latest block number on the Story testnet can be retrieved by calling eth_blockNumber on the RPC endpoint https://testnet.storyrpc.io. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-1) `source_document_id=srcdoc_b96b092961e1dcf7152abe43e62ddec8` `source_revision_id=srcrev_2c78dfd508c1981220a3d08e48797cf7` `chunk_id=srcchunk_8a2e17db3a6e2bcec408dd2ec48d90c4` `native_locator=https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-1` `source_timestamp=2024-10-04T18:37:00Z`
- The account balance for a given address on the Story testnet can be checked by calling eth_getBalance with the address and 'latest' block parameter on the RPC endpoint https://testnet.storyrpc.io. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-2) `source_document_id=srcdoc_b96b092961e1dcf7152abe43e62ddec8` `source_revision_id=srcrev_2c78dfd508c1981220a3d08e48797cf7` `chunk_id=srcchunk_227778e93e92f6f488a97f64dd5f8220` `native_locator=https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-2` `source_timestamp=2024-10-04T18:37:00Z`
- The existence and status of a systemd service can be checked using systemctl list-units --type=service --all | grep <systemd-svc-name>, for example grep node-story. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-2) `source_document_id=srcdoc_b96b092961e1dcf7152abe43e62ddec8` `source_revision_id=srcrev_2c78dfd508c1981220a3d08e48797cf7` `chunk_id=srcchunk_227778e93e92f6f488a97f64dd5f8220` `native_locator=https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb#chunk-2` `source_timestamp=2024-10-04T18:37:00Z`

## Sources

- `source_document_id`: `srcdoc_b96b092961e1dcf7152abe43e62ddec8`
- `source_revision_id`: `srcrev_2c78dfd508c1981220a3d08e48797cf7`
- `source_url`: [Notion source](https://www.notion.so/Commands-for-Troubleshooting-113051299a548038bdb6cd97020965fb)
