---
title: "Odyssey Node Setup"
type: "runbook"
slug: "runbooks/odyssey-node-setup"
freshness: "2024-12-03T23:41:00Z"
tags:
  - "geth"
  - "node"
  - "odyssey"
  - "setup"
  - "story"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_86e62049ec10873ddf921c485c420875"
conflict_state: "none"
---

# Odyssey Node Setup

## Summary

Guide for setting up a node for the Odyssey test network, including execution client (story-geth) and consensus client (story) setup, system requirements, configuration, automation, and debugging.

## Claims

- Story draws inspiration from ETH PoS in decoupling execution and consensus clients. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_d903ed930d83e3d0900abf4d65af4461` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1` `source_timestamp=2024-12-03T23:41:00Z`
- The execution client `story-geth` relays EVM blocks into the `story` consensus client via Engine ABI, using an ABCI++ adapter to make EVM state compatible with that of CometBFT. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_d903ed930d83e3d0900abf4d65af4461` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1` `source_timestamp=2024-12-03T23:41:00Z`
- With this architecture, consensus efficiency is no longer bottlenecked by execution transaction throughput. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_d903ed930d83e3d0900abf4d65af4461` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1` `source_timestamp=2024-12-03T23:41:00Z`
- The `story` and `geth` binaries are available from the latest release pages. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_d903ed930d83e3d0900abf4d65af4461` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1` `source_timestamp=2024-12-03T23:41:00Z`
- The `story-geth` execution client release link is https://github.com/piplabs/story-geth/releases/tag/v0.10.0. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_d903ed930d83e3d0900abf4d65af4461` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1` `source_timestamp=2024-12-03T23:41:00Z`
- The `story` consensus client release link is https://github.com/piplabs/story/releases/tag/v0.12.0. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_d903ed930d83e3d0900abf4d65af4461` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-1` `source_timestamp=2024-12-03T23:41:00Z`
- On Mac OS X, the binaries have yet to be signed by the build process, so you may need to unquarantine them manually using `sudo xattr -rd com.apple.quarantine ./story`. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- Initialize the `story` client with `./story init --network odyssey --external-address ${EXERNAL_ADDRESS}:26656`. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- By default, `story init` uses your username for the moniker; you may override this by passing in `-moniker ${NODE_MONIKER}`. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- If you would like to initialize the node using your own data directory, you can pass in `-home ${STORY_DATA_DIR}`. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- If you already have config and data files, and would like to re-initialize from scratch, you can add the `--clean` flag. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- We recommend adding the `--external-address ${EXTERNAL_IP}:26656` parameter during initialization via `--init` to prevent issues due to cometBFT syncing. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- Run `story` with `./story run`. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- Currently you might see a bunch of `Stopping peer for error` logs - this is a known issue around peer connection stability with our bootnodes that we are currently fixing - for now please ignore it and rest assured that it does not impact block progression. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- To re-join the network while preserving your key, run `rm -rf ${STORY_DATA_ROOT}/data/* && echo '{"height": "0", "round": 0, "step": 0}' > ${STORY_DATA_ROOT}/data/priv_validator_state.json && ./story run`. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- To join the network from a completely fresh state (WARNING: THIS WILL DELETE YOUR `priv_validator_key.json` FILE), run `rm -rf ${STORY_DATA_ROOT} && ./story init --network odyssey && ./story run`. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- To quickly check if the node is syncing, you can check the geth RPC endpoint to see if blocks are increasing using `curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545`. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_4dd9f0de37a50e54a43ccf8e6f289e6c` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-2` `source_timestamp=2024-12-03T23:41:00Z`
- Set Up Story v0.12.0 linux amd64 download link: https://github.com/piplabs/story/releases/download/v0.12.0/story-linux-amd64 `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_13eb3282740f849c8649591f9774622f` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3` `source_timestamp=2024-12-03T23:41:00Z`
- Set Up Story v0.12.0 linux arm64 download link: https://github.com/piplabs/story/releases/download/v0.12.0/story-linux-arm64 `claim:claim_1_19` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_13eb3282740f849c8649591f9774622f` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3` `source_timestamp=2024-12-03T23:41:00Z`
- Initialize the node with `story init --network odyssey --moniker "Your_moniker_name"`. `claim:claim_1_20` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_13eb3282740f849c8649591f9774622f` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3` `source_timestamp=2024-12-03T23:41:00Z`
- `${STORY_DATA_ROOT}/config/config.toml` can be modified to change network and consensus settings. `claim:claim_1_21` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_13eb3282740f849c8649591f9774622f` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3` `source_timestamp=2024-12-03T23:41:00Z`
- `${STORY_DATA_ROOT}/config/story.toml` can be modified to update various client configs. `claim:claim_1_22` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_13eb3282740f849c8649591f9774622f` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3` `source_timestamp=2024-12-03T23:41:00Z`
- `${STORY_DATA_ROOT}/priv_validator_key.json` is a sensitive file containing your validator key, but may be replaced with your own. `claim:claim_1_23` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_13eb3282740f849c8649591f9774622f` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3` `source_timestamp=2024-12-03T23:41:00Z`
- Sample Systemd configuration for geth and story services is provided for local and VPS automation on Linux. `claim:claim_1_24` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_13eb3282740f849c8649591f9774622f` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3` `source_timestamp=2024-12-03T23:41:00Z`
- To check the status of `story` while it is running, query its internal JSONRPC/HTTP endpoint, e.g., `curl localhost:26657/net_info | jq '.result.peers[].node_info.moniker'` to get a list of consensus peers. `claim:claim_1_25` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3) `source_document_id=srcdoc_067ac43005f03d689be8cfd5c51a5228` `source_revision_id=srcrev_86e62049ec10873ddf921c485c420875` `chunk_id=srcchunk_13eb3282740f849c8649591f9774622f` `native_locator=https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b#chunk-3` `source_timestamp=2024-12-03T23:41:00Z`

## Sources

- `source_document_id`: `srcdoc_067ac43005f03d689be8cfd5c51a5228`
- `source_revision_id`: `srcrev_86e62049ec10873ddf921c485c420875`
- `source_url`: [Notion source](https://www.notion.so/Odyssey-Node-Setup-128051299a5480418a35eff9fa2b3a4b)
