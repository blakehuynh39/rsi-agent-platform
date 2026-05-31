---
title: "Partner Testnet v1.0.0-alpha Release"
type: "project"
slug: "projects/partner-testnet-v1-0-0-alpha-release"
freshness: "2024-07-30T16:20:00Z"
tags:
  - "blockchain"
  - "partner"
  - "release"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_6542b890b33267426c704921cf69920d"
conflict_state: "none"
---

# Partner Testnet v1.0.0-alpha Release

## Summary

Release of the L1 blockchain partner testnet version 1.0.0-alpha, enabling partners to test and develop with EVM-equivalence, 2-second block time, staking, and slashing.

## Claims

- The partner testnet v1.0.0-alpha was released on Mon Jul 29 2024. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- The testnet is EVM-equivalent, supporting all EVM transaction types via a communication layer between CometBFT consensus and Ethereum execution. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Block time is 2 seconds with a 30 million gas limit per block. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Staking contract in the execution layer allows creating validators, depositing, redelegating, and withdrawing stakes. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Stake rewards are automatically distributed to stakers/delegators in the execution layer from the consensus layer. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Slashing is enabled for misbehaving validators (double-sign, offline). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Known issue: Validator and stake data may have inconsistencies between execution and consensus layers due to slashing. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Known issue: When geth has multiple hard forks and a new node joins, the new node may choose the wrong fork to sync. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Known issue: When the network resets, the reset nodes may sync with existing nodes that haven’t been reset. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Known issue: Jailed validators may not rejoin validator sets after calling unjail function. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`
- Known issue: RPC `eth_getFilterLogs` method may return "filter not found". `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946) `source_document_id=srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84` `source_revision_id=srcrev_6542b890b33267426c704921cf69920d` `chunk_id=srcchunk_c1fe4a47cf0df65d6ca7ce6c38fc0de9` `native_locator=https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946` `source_timestamp=2024-07-30T16:20:00Z`

## Sources

- `source_document_id`: `srcdoc_bbc60af68e9dd8e4f8cdac8e01020f84`
- `source_revision_id`: `srcrev_6542b890b33267426c704921cf69920d`
- `source_url`: [Notion source](https://www.notion.so/Partner-Testnet-Release-Notes-Version-1-0-66c414119dc24b2a992eb0add80cd946)
