---
title: "Halo Configuration"
type: "system"
slug: "systems/halo-configuration"
freshness: "2026-05-05T06:35:09Z"
tags:
  - "configuration"
  - "halo"
  - "toml"
owners: []
source_revision_ids:
  - "srcrev_728fda2d51fb2285c3720908aa9d02bc"
conflict_state: "none"
---

# Halo Configuration

## Summary

Configuration options for the Halo binary as defined in halo.toml.

## Claims

- The Halo configuration file is a TOML config file. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- The version of the Halo binary that created or last modified the config file is v0.1.5. This value should not be modified. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- The network parameter specifies the Omni network to participate in: mainnet, testnet, or devnet. The current value is devnet. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- The engine-endpoint is the Omni execution client Engine API http endpoint, set to http://localhost:8551. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- The engine-jwt-file is the Omni execution client JWT file used for authentication, located at /home/ec2-user/geth/data/geth/jwtsecret. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- The eigenlayer-key-password is the Eigenlayer generated operator private key password. The key itself should be stored in <home_dir>/config/*.ecdsa.key.json. The current value is empty. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- snapshot-interval specifies the height interval at which halo will take state sync snapshots. It defaults to 1000 (roughly once an hour). Setting this to 0 disables state snapshots. The current value is 1000. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all). The current value is 2. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- min-retain-blocks defines the minimum block height offset from the current block being committed, such that all blocks past this offset are pruned from CometBFT. A value of 0 indicates that no blocks should be pruned. This configuration only prunes CometBFT blocks, not application state. The current value is 0. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- The pruning parameter controls state storage: "default" keeps the last 362880 states, pruning at 10 block intervals; "nothing" saves all historic states (archiving node); "everything" keeps only the 2 latest states, pruning at 10 block intervals. The current value is "nothing". `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- app-db-backend defines the database backend type for the application and snapshots DBs. An empty string falls back to the db_backend value in CometBFT's config.toml. The current value is "goleveldb". `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- evm-build-delay defines the minimum delay between triggering an EVM payload build and fetching the result. It should be slightly higher than geth's --miner.recommit value. The current value is "600ms". `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`
- evm-build-optimistic enables optimistic EVM payload building. If true, the EVM payload is triggered on previous finalisation, allowing more time for block building while ensuring faster consensus blocks. The current value is true. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b) `source_document_id=srcdoc_5e2c0745a243a4f3e7d29db507e68644` `source_revision_id=srcrev_728fda2d51fb2285c3720908aa9d02bc` `chunk_id=srcchunk_ac0b7e1e727bfe579aaba2017ad92505` `native_locator=https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b` `source_timestamp=2026-05-05T06:35:09Z`

## Sources

- `source_document_id`: `srcdoc_5e2c0745a243a4f3e7d29db507e68644`
- `source_revision_id`: `srcrev_728fda2d51fb2285c3720908aa9d02bc`
- `source_url`: [Notion source](https://www.notion.so/halo-toml-0de5fdb3bb1746cd9a9b3f0bb79e972b)
