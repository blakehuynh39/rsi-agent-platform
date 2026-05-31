---
title: "Poseidon Contracts Deployment / Upgrade Guide"
type: "runbook"
slug: "runbooks/poseidon-contracts-deployment-upgrade-guide"
freshness: "2025-10-30T18:59:00Z"
tags:
  - "contracts"
  - "deployment"
  - "l1"
  - "l2"
  - "poseidon"
  - "psdn"
  - "upgrade"
owners: []
source_revision_ids:
  - "srcrev_e92013ced42c3d1f5b8e80dda69af31a"
conflict_state: "none"
---

# Poseidon Contracts Deployment / Upgrade Guide

## Summary

Runbook for deploying and upgrading Poseidon subnet contracts on L2 and the PSDN token on L1, including environment setup, deployment commands, verification, and upgrade scripting.

## Claims

- The Poseidon subnet contracts live on the Poseidon L2 network, while the PSDN token is deployed on L1. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-1) `source_document_id=srcdoc_770c27f0c8dccfff5421bf48022e2a49` `source_revision_id=srcrev_e92013ced42c3d1f5b8e80dda69af31a` `chunk_id=srcchunk_27dbce78b0b054dd82ec9254793f3757` `native_locator=https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-1` `source_timestamp=2025-10-30T18:59:00Z`
- Deployment environment variables include RPC_URL, ADMIN address, and VERIFIER_URL, with the admin private key stored in Foundry Keystore under the 'subnet-admin' account name. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-1) `source_document_id=srcdoc_770c27f0c8dccfff5421bf48022e2a49` `source_revision_id=srcrev_e92013ced42c3d1f5b8e80dda69af31a` `chunk_id=srcchunk_27dbce78b0b054dd82ec9254793f3757` `native_locator=https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-1` `source_timestamp=2025-10-30T18:59:00Z`
- L2 contracts are deployed using a Foundry script command that broadcasts and verifies via Blockscout. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-1) `source_document_id=srcdoc_770c27f0c8dccfff5421bf48022e2a49` `source_revision_id=srcrev_e92013ced42c3d1f5b8e80dda69af31a` `chunk_id=srcchunk_27dbce78b0b054dd82ec9254793f3757` `native_locator=https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-1` `source_timestamp=2025-10-30T18:59:00Z`
- Contract verification on L2 can be performed manually via the explorer using the Solidity (Foundry) verification method. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-1) `source_document_id=srcdoc_770c27f0c8dccfff5421bf48022e2a49` `source_revision_id=srcrev_e92013ced42c3d1f5b8e80dda69af31a` `chunk_id=srcchunk_27dbce78b0b054dd82ec9254793f3757` `native_locator=https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-1` `source_timestamp=2025-10-30T18:59:00Z`
- Contract upgrades are performed by adapting the UpgradeTaskQueue script, replacing the contract address and name, deploying a new implementation, and calling upgradeToAndCall. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-2) `source_document_id=srcdoc_770c27f0c8dccfff5421bf48022e2a49` `source_revision_id=srcrev_e92013ced42c3d1f5b8e80dda69af31a` `chunk_id=srcchunk_faf02739365f27dcaa96afa6d51f72d4` `native_locator=https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-2` `source_timestamp=2025-10-30T18:59:00Z`
- The upgrade script is executed with a forge command using the appropriate RPC URL (L1_RPC or L2_RPC). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-2) `source_document_id=srcdoc_770c27f0c8dccfff5421bf48022e2a49` `source_revision_id=srcrev_e92013ced42c3d1f5b8e80dda69af31a` `chunk_id=srcchunk_faf02739365f27dcaa96afa6d51f72d4` `native_locator=https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded#chunk-2` `source_timestamp=2025-10-30T18:59:00Z`

## Sources

- `source_document_id`: `srcdoc_770c27f0c8dccfff5421bf48022e2a49`
- `source_revision_id`: `srcrev_e92013ced42c3d1f5b8e80dda69af31a`
- `source_url`: [Notion source](https://www.notion.so/Contracts-Deployment-Upgrade-Guide-25c051299a548086a97def7861404ded)
