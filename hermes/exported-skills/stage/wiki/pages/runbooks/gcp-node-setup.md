---
title: "GCP Node Setup"
type: "runbook"
slug: "runbooks/gcp-node-setup"
freshness: "2026-05-05T06:41:19Z"
tags:
  - "gcp"
  - "halo"
  - "node-setup"
  - "omni"
  - "runbook"
owners: []
source_revision_ids:
  - "srcrev_441b8711845981e6d80e2464e24eb741"
conflict_state: "none"
---

# GCP Node Setup

## Summary

Guide for setting up a GCP node for the Omni/Halo network, including requirements, steps, common issues, and useful commands.

## Claims

- A Jdub script (node-init-cmds.sh) is available as a resource. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_995c4e49ffa374519cffa4f93e92af3c` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1` `source_timestamp=2026-05-05T06:41:19Z`
- Requirements: Enable Kubernetes Engine API and ensure the public key matches ~/halo/config/priv_validator_key.json created by halo init. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_995c4e49ffa374519cffa4f93e92af3c` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1` `source_timestamp=2026-05-05T06:41:19Z`
- Setup steps: Clone Iliad repo, cd halo, go build, sudo cp halo /usr/local/bin, halo init --clean --network testnet --home ~/halo. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_995c4e49ffa374519cffa4f93e92af3c` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1` `source_timestamp=2026-05-05T06:41:19Z`
- Issue 1: Consensus Client Not Found / Execution Error on beacon chain. Solution: Ensure app_state.genutil.gen_txs.body has messages for all validator nodes and pubkey.key corresponds to the actual pubkey from halo_folder/config/priv_validator_key.json. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_995c4e49ffa374519cffa4f93e92af3c` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1` `source_timestamp=2026-05-05T06:41:19Z`
- Issue 2: Beacon initialization error (e.g. via engine_forchoiceUpdatedV3). Solution not provided in the source. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_995c4e49ffa374519cffa4f93e92af3c` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-1` `source_timestamp=2026-05-05T06:41:19Z`
- Issue 3: Could not figure out the node id. Solution: Run cometbft show_node_id --home ~/halo2. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_125eb3cefd99b84ab89592088c1fea30` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2` `source_timestamp=2026-05-05T06:41:19Z`
- Useful Docker commands: docker images --filter=reference='omniops/halo' and docker run -v output:/halo omniops/halo init --network testnet. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_125eb3cefd99b84ab89592088c1fea30` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2` `source_timestamp=2026-05-05T06:41:19Z`
- For debugging JSON-RPC, use: curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}' http://localhost:8545 | jq '.' `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_125eb3cefd99b84ab89592088c1fea30` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2` `source_timestamp=2026-05-05T06:41:19Z`
- GETH environment variable JWTSECRET_FILE is set to /Users/leeren/go-ethereum/data. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_125eb3cefd99b84ab89592088c1fea30` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2` `source_timestamp=2026-05-05T06:41:19Z`
- The genesis file for GETH uses chainId 1513 and includes various fork blocks (homesteadBlock, eip150Block, etc.) all set to 0, with terminalTotalDifficulty 0 and terminalTotalDifficultyPassed true, shanghaiTime 0, cancunTime 0. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_125eb3cefd99b84ab89592088c1fea30` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-2` `source_timestamp=2026-05-05T06:41:19Z`
- Geth defaults can be viewed via geth dumpconfig. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-4) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_fe7be1967440b7931a5afa7064fb7a53` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-4` `source_timestamp=2026-05-05T06:41:19Z`
- Explanation for JWT generation is available at https://notes.ethereum.org/@launchpad/kiln. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-4) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_fe7be1967440b7931a5afa7064fb7a53` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-4` `source_timestamp=2026-05-05T06:41:19Z`
- The document includes a 'Full Withdrawal Notes' section, but no content is provided. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-4) `source_document_id=srcdoc_0f5d84e0ab01a4a855bd41454a8f0498` `source_revision_id=srcrev_441b8711845981e6d80e2464e24eb741` `chunk_id=srcchunk_fe7be1967440b7931a5afa7064fb7a53` `native_locator=https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420#chunk-4` `source_timestamp=2026-05-05T06:41:19Z`

## Open Questions

- Full Withdrawal Notes section is empty.
- GETH JWTSECRET_FILE path may be user-specific.
- Solution for beacon initialization error (issue 2) is missing from the source.

## Sources

- `source_document_id`: `srcdoc_0f5d84e0ab01a4a855bd41454a8f0498`
- `source_revision_id`: `srcrev_441b8711845981e6d80e2464e24eb741`
- `source_url`: [Notion source](https://www.notion.so/GCP-Node-Setup-b914a3d77e4749beb2b915e63ef8a420)
