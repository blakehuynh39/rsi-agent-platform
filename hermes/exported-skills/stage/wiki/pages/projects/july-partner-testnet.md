---
title: "July Partner Testnet"
type: "project"
slug: "projects/july-partner-testnet"
freshness: "2026-05-05T06:36:21Z"
tags:
  - "devops"
  - "infrastructure"
  - "slashing"
  - "staking"
  - "testing"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_a3c8d446bed67a2a7d261d11b16f720f"
conflict_state: "none"
---

# July Partner Testnet

## Summary

Project plan for the July partner testnet, covering new features (slashing, epoch staking, staking dashboard APIs, cosmos API, staking hardening), fixes, separation from Omni codebase, infrastructure improvements, and testing.

## Claims

- The testnet will introduce slashing as a new feature. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`
- Epoch staking (P1) is planned. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`
- Staking dashboard APIs will include APY, historical data (uncertain), and address/pubkey mapping. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`
- Cosmos API will be added for testnet observability. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`
- Staking hardening is planned. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`
- Fixes include GitHub issues #61 and #59, and debugging the 'Prefetcher missed to load trie' error. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`
- The codebase will be separated from Omni by deleting irrelevant code (attestation, valsync), renaming addresses from 'omni' to 'story', removing logs, and moving to a separate repo. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`
- Infrastructure improvements include DevEx (4+1 devnet, local dev setup), testnet infra (config consolidation, RPC nodes with load balancer, bootnode/seednode setup, logging, health checks, disaster recovery, distributed nodes), and DevOps (daily deployment, regression testing, release packaging, fast sync, anti-spam). `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`
- Testing will cover regression (API tests, staking, account transfer), performance (transaction finality measurement), and consensus (malicious node cases, slashing tests, node crash). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f) `source_document_id=srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3` `source_revision_id=srcrev_a3c8d446bed67a2a7d261d11b16f720f` `chunk_id=srcchunk_b01defc0b41b067e11388189f8614a70` `native_locator=https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f` `source_timestamp=2026-05-05T06:36:21Z`

## Open Questions

- Is historical data included in staking dashboard APIs? (marked with '?')
- Should distributed nodes use containers or bare metal?
- Should orchestration use k8s or nomad?
- Will multiple cloud providers be used?
- Will nodes be deployed in multiple regions?

## Sources

- `source_document_id`: `srcdoc_ae20a4ff3c7642cef5da6d525ee20fe3`
- `source_revision_id`: `srcrev_a3c8d446bed67a2a7d261d11b16f720f`
- `source_url`: [Notion source](https://www.notion.so/July-partner-testnet-99671b15cad2431a82545d1996865a8f)
