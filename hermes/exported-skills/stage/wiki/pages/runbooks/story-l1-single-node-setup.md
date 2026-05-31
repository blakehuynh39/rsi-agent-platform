---
title: "Story L1 Single Node Setup"
type: "runbook"
slug: "runbooks/story-l1-single-node-setup"
freshness: "2024-09-19T22:00:00Z"
tags:
  - "debugging"
  - "geth"
  - "iliad"
  - "node-setup"
  - "story-l1"
owners: []
source_revision_ids:
  - "srcrev_2644a85650bcffe9bebeeeb3500ef3bf"
conflict_state: "none"
---

# Story L1 Single Node Setup

## Summary

Instructions for setting up a single-node private Story L1 network locally on macOS, including execution client (geth), consensus client (iliad), genesis configuration, and debugging with Delve.

## Claims

- Create geth and config folders: mkdir -p ~/Library/Story/story and mkdir -p ~/Library/Story/geth. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_b78ab1aeb94f105e7732ff790e6dcd17` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1` `source_timestamp=2024-09-19T22:00:00Z`
- Place genesis.json and geth.toml into ~/geth/config folder. In geth.toml, set DataDir, JWTSecret, and remove BootstrapNodes (set to empty list). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_b78ab1aeb94f105e7732ff790e6dcd17` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1` `source_timestamp=2024-09-19T22:00:00Z`
- Clone story-geth repo, build geth with make geth (requires Go >=1.21), then initialize genesis block: ./build/bin/geth init --datadir="~/geth/data" ~/geth/config/genesis.json. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_b78ab1aeb94f105e7732ff790e6dcd17` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1` `source_timestamp=2024-09-19T22:00:00Z`
- Run geth: ./build/bin/geth --config ~/geth/config/geth.toml. To restart cleanly: rm -rf ~/geth/data/* && ./build/bin/geth init --datadir="~/geth/data" ~/geth/config/genesis.json && ./build/bin/geth --config ~/geth/config/geth.toml. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_b78ab1aeb94f105e7732ff790e6dcd17` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1` `source_timestamp=2024-09-19T22:00:00Z`
- Create iliad folder: mkdir -p ~/iliad/config. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_b78ab1aeb94f105e7732ff790e6dcd17` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1` `source_timestamp=2024-09-19T22:00:00Z`
- Clone Iliad-Node repo, checkout relevant branch, pull latest, then build iliad binary: cd ~/Iliad_Node/client && go build && mv ./client ./iliad. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_b78ab1aeb94f105e7732ff790e6dcd17` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1` `source_timestamp=2024-09-19T22:00:00Z`
- Initialize iliad testnet configuration: ./iliad init --clean --network testnet --home ~/iliad. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_b78ab1aeb94f105e7732ff790e6dcd17` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-1` `source_timestamp=2024-09-19T22:00:00Z`
- Modify genesis.json: keep only one gen_txs entry, replace pubkey.key, delegator_address, validator_address with values from step 4 script; update first account address in auth.accounts and bank.balances to accAddr; ensure supply matches sum of balances; set evmstaking.params.min_partial_withdrawal_amount to 100. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-2) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_1c421ff4d101fac5ab489a5b8628f6d0` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-2` `source_timestamp=2024-09-19T22:00:00Z`
- Copy contents of iliad.toml from node-launcher repo to ~/iliad/config/iliad.toml, ensuring engine-endpoint points to localhost engine API. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-2) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_1c421ff4d101fac5ab489a5b8628f6d0` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-2` `source_timestamp=2024-09-19T22:00:00Z`
- Install Delve (brew install delve), build iliad with debug flags (go build -gcflags "all=-N -l"), then run with dlv exec ./iliad -- run --home ~/iliad. For frequent re-runs, use: rm -rf ~/iliad/data/* && echo '{"height": "0", "round": 0, "step": 0}' > ~/iliad/data/priv_validator_state.json && dlv exec ./iliad -- run --home ~/iliad. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-3) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_7218e1e7d4254510c1357bb356e333c6` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-3` `source_timestamp=2024-09-19T22:00:00Z`
- In Delve debugger, set breakpoints (e.g., break x/evmengine/keeper/abci.go:32), use c to continue, stack for stack trace, goroutines to list goroutines, goroutine to switch, print for expressions, args for function arguments, locals for local variables. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-3) `source_document_id=srcdoc_a90e629db8afd1b4022a7c78f112f60e` `source_revision_id=srcrev_2644a85650bcffe9bebeeeb3500ef3bf` `chunk_id=srcchunk_7218e1e7d4254510c1357bb356e333c6` `native_locator=https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083#chunk-3` `source_timestamp=2024-09-19T22:00:00Z`

## Sources

- `source_document_id`: `srcdoc_a90e629db8afd1b4022a7c78f112f60e`
- `source_revision_id`: `srcrev_2644a85650bcffe9bebeeeb3500ef3bf`
- `source_url`: [Notion source](https://www.notion.so/Story-L1-One-Node-Setup-106051299a548039a78dfd5c996a5083)
