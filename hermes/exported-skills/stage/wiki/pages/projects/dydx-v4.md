---
title: "Dydx v4"
type: "project"
slug: "projects/dydx-v4"
freshness: "2024-04-23T08:31:00Z"
tags:
  - "appchain"
  - "cosmos"
  - "defi"
  - "orderbook"
owners: []
source_revision_ids:
  - "srcrev_cb9908a3dcb31dc5d04d797e232b60b0"
conflict_state: "none"
---

# Dydx v4

## Summary

Dydx v4 is a decentralized derivatives exchange built as a standalone L1 blockchain using CometBFT and CosmosSDK. It uses an off-chain in-memory orderbook for matching, with trades committed on-chain, and off-chain indexer infrastructure (Postgres, Redis, Kafka).

## Claims

- Dydx v4 is an L1 blockchain built on CometBFT and CosmosSDK, using Tendermint Proof-of-stake consensus. `claim:claim_dydx_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-1) `source_document_id=srcdoc_514d4c4855471285a067efd05d2f1b8f` `source_revision_id=srcrev_cb9908a3dcb31dc5d04d797e232b60b0` `chunk_id=srcchunk_99ac679a0594f16c64dac505c40fbd14` `native_locator=https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-1` `source_timestamp=2024-04-23T08:31:00Z`
- Each validator runs an in-memory orderbook that is never committed to consensus (off-chain). Orders and cancellations are gossiped through the network, making the orderbook eventually consistent. `claim:claim_dydx_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-1) `source_document_id=srcdoc_514d4c4855471285a067efd05d2f1b8f` `source_revision_id=srcrev_cb9908a3dcb31dc5d04d797e232b60b0` `chunk_id=srcchunk_99ac679a0594f16c64dac505c40fbd14` `native_locator=https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-1` `source_timestamp=2024-04-23T08:31:00Z`
- Trades are matched by the network in real-time and the resulting trades are committed on-chain each block. `claim:claim_dydx_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-1) `source_document_id=srcdoc_514d4c4855471285a067efd05d2f1b8f` `source_revision_id=srcrev_cb9908a3dcb31dc5d04d797e232b60b0` `chunk_id=srcchunk_99ac679a0594f16c64dac505c40fbd14` `native_locator=https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-1` `source_timestamp=2024-04-23T08:31:00Z`
- Traders do not pay gas fees; instead they pay fees based on executed trades, which accrue to validators and their stakers. `claim:claim_dydx_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-1) `source_document_id=srcdoc_514d4c4855471285a067efd05d2f1b8f` `source_revision_id=srcrev_cb9908a3dcb31dc5d04d797e232b60b0` `chunk_id=srcchunk_99ac679a0594f16c64dac505c40fbd14` `native_locator=https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-1` `source_timestamp=2024-04-23T08:31:00Z`
- Indexers use Postgres for on-chain data, Redis for off-chain data, and Kafka for consuming and streaming on/off-chain data. `claim:claim_dydx_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-2) `source_document_id=srcdoc_514d4c4855471285a067efd05d2f1b8f` `source_revision_id=srcrev_cb9908a3dcb31dc5d04d797e232b60b0` `chunk_id=srcchunk_2348e956a4ee082f38b3967ea052ecdf` `native_locator=https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-2` `source_timestamp=2024-04-23T08:31:00Z`
- Dydx is building three open-source front ends: a React web app, an iOS app in Swift, and an Android app in Kotlin. The web app interacts with the Indexer via API for orderbook data and sends trades directly to the chain. `claim:claim_dydx_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-2) `source_document_id=srcdoc_514d4c4855471285a067efd05d2f1b8f` `source_revision_id=srcrev_cb9908a3dcb31dc5d04d797e232b60b0` `chunk_id=srcchunk_2348e956a4ee082f38b3967ea052ecdf` `native_locator=https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-2` `source_timestamp=2024-04-23T08:31:00Z`
- The dYdX Chain could offer up to 2,000 transactions per second. `claim:claim_dydx_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-2) `source_document_id=srcdoc_514d4c4855471285a067efd05d2f1b8f` `source_revision_id=srcrev_cb9908a3dcb31dc5d04d797e232b60b0` `chunk_id=srcchunk_2348e956a4ee082f38b3967ea052ecdf` `native_locator=https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b#chunk-2` `source_timestamp=2024-04-23T08:31:00Z`

## Related Pages

- `concepts/enshrined-protocol`

## Sources

- `source_document_id`: `srcdoc_514d4c4855471285a067efd05d2f1b8f`
- `source_revision_id`: `srcrev_cb9908a3dcb31dc5d04d797e232b60b0`
- `source_url`: [Notion source](https://www.notion.so/Dydx-f771c36473b1474ab4bf7cfa378d714b)
