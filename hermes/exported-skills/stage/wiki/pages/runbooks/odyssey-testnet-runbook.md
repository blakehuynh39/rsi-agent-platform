---
title: "Odyssey Testnet Runbook"
type: "runbook"
slug: "runbooks/odyssey-testnet-runbook"
freshness: "2024-11-12T21:57:00Z"
tags:
  - "infrastructure"
  - "odyssey"
  - "runbook"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94"
conflict_state: "none"
---

# Odyssey Testnet Runbook

## Summary

Runbook for the Odyssey testnet, including service endpoints, infrastructure details, and validator/RPC node configurations.

## Claims

- The Odyssey testnet ChainID is 1516. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The external blockchain explorer for Odyssey testnet is https://odyssey.storyscan.xyz/. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The internal blockchain explorer for Odyssey testnet is https://odyssey-testnet-explorer.storyscan.xyz/ and is marked for decommission soon. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The internal RPC endpoint for Odyssey testnet is https://odyssey.storyrpc.io/. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The public RPC endpoint for Odyssey testnet is https://rpc.odyssey.storyrpc.io/. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The Grafana monitoring dashboard for Odyssey testnet is available at https://monitor.storyprotocol.net/. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The PagerDuty service directory for Odyssey testnet is at https://storyprotocol.pagerduty.com/. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The testnet faucet for Odyssey is available at https://faucet.story.foundation/. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The network status page for Odyssey testnet is hosted at https://story-haodi.betteruptime.com/. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The Slack communication channel for Odyssey testnet on-call is #l1-oncall. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The Odyssey testnet infrastructure includes 2 bootnodes and 8 validators. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- Story validator 1 is reachable at c620962842166c42083c348baf7f68f44ae83e6e@3.20.133.100:26656. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- Story validator 2 is reachable at 04e5734295da362f09a61dd0a9999449448a0e5c@52.14.39.177:26656. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- Story validator 3 is reachable at 9e2fabda41e3c3317c25f5ef6c604c1d78370aba@50.112.252.101:26656. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- Story validator 4 is reachable at 046909534c2849ff8dccc15ee43ee63d2c60b21c@54.190.123.194:26656. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- Bharvest validator 1 is reachable at 793b2289ef17b17b229a6487b13ee0ce08056328@odyssey-testnet-p1.bharvest.io:26656. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- Bharvest validator 2 is reachable at 097657b156965c8f99a8c4fcb569b84767f7d076@odyssey-testnet-p2.bharvest.io:26656. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- Blockdaemon validator 1 is reachable at 2086affe2a3ea6ba3a9e6ca16a3ba406906f6eea@141.98.217.151:26656. `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- Blockdaemon validator 2 is reachable at 2fb7d62902b9aeb9615eebc980d750f9e11ac872@64.130.55.48:26656. `claim:claim_1_19` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- There are 2 internal RPC nodes on AWS and 4 external RPC nodes on Velia with Intel Xeon E-2286G, 32 GB RAM, 1 TB SATA, 100 TB traffic, and 1 Gbps internet. `claim:claim_1_20` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The external RPC nodes have IP addresses 148.72.138.5, 148.72.138.4, 148.72.138.6, and 148.72.138.7. `claim:claim_1_21` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`
- The external explorer is hosted by the BlockScout team at https://odyssey.storyscan.xyz/. `claim:claim_1_22` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590) `source_document_id=srcdoc_0de2a600690d5591e594f115e44d4b0b` `source_revision_id=srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94` `chunk_id=srcchunk_44413fb303bc122a25df937c41c54f61` `native_locator=https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590` `source_timestamp=2024-11-12T21:57:00Z`

## Sources

- `source_document_id`: `srcdoc_0de2a600690d5591e594f115e44d4b0b`
- `source_revision_id`: `srcrev_fc47152d33d6dd0d28d01a6ffd6a3c94`
- `source_url`: [Notion source](https://www.notion.so/Odyssey-Testnet-Runbook-112051299a548066b55eed90bbc58590)
