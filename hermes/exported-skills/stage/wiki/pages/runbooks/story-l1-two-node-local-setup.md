---
title: "Story L1 Two-Node Local Setup"
type: "runbook"
slug: "runbooks/story-l1-two-node-local-setup"
freshness: "2024-10-02T23:46:00Z"
tags:
  - "geth"
  - "iliad"
  - "local-network"
  - "setup"
  - "story-l1"
owners: []
source_revision_ids:
  - "srcrev_d70592364b1e7ce5f36332f30f6ccf1a"
conflict_state: "none"
---

# Story L1 Two-Node Local Setup

## Summary

Step-by-step guide to set up a two-node private Story L1 network on macOS, covering execution client (geth) and consensus client (iliad) configuration, genesis file customization, and debugging.

## Claims

- Create two folders geth-1 and geth-2 each with an embedded config folder for storing execution client configuration and data files. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_3a8343fef64640797b2abb720eb63cb6` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1` `source_timestamp=2024-10-02T23:46:00Z`
- Place genesis.json and geth.toml files in each geth folder. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_3a8343fef64640797b2abb720eb63cb6` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1` `source_timestamp=2024-10-02T23:46:00Z`
- In geth.toml, set DataDir and JWTSecret paths for each node; JWTSecret will be auto-generated later under geth-{1,2}/data/geth/jwtsecret. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_3a8343fef64640797b2abb720eb63cb6` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1` `source_timestamp=2024-10-02T23:46:00Z`
- The BootstrapNodes array under [Node.P2P] must include the enode URL of connected peer nodes; for two nodes, each geth's config should list the other's enode. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_3a8343fef64640797b2abb720eb63cb6` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1` `source_timestamp=2024-10-02T23:46:00Z`
- To fetch the enode prior to running geth, use bootnode -nodekeyhex $(cat ~/geth-{1,2}/data/geth/nodekey) --writeaddress. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_3a8343fef64640797b2abb720eb63cb6` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-1` `source_timestamp=2024-10-02T23:46:00Z`
- If nodes are not synchronizing, clear the geth data directory, reinitialize with genesis, and re-run. Sample commands provided for geth-1 and geth-2. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-2) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_5f6d3c32993bae872182294add937f30` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-2` `source_timestamp=2024-10-02T23:46:00Z`
- For iliad consensus clients, modify config.toml: set different prometheus_listen_addr for each node, replace seeds/persistent_peers with peer node IDs and addresses, set log_level to debug, empty external_address, and addr_book_strict to false. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-3) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_117ec78f5c4f1efc96572625265bd027` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-3` `source_timestamp=2024-10-02T23:46:00Z`
- Customize genesis.json: replace validator pubkeys, delegator/validator addresses, account addresses, balances, and ensure supply matches total balances; set min_partial_withdrawal_amount to 100. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-4) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_dbfc28b3305f72891d4eff077d3a2620` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-4` `source_timestamp=2024-10-02T23:46:00Z`
- To re-run iliad, restart geth, clear iliad data, reset priv_validator_state.json to height 0, round 0, step 0, then run iliad binary. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-5) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_95424b01c803cb7b942957e08374b8af` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-5` `source_timestamp=2024-10-02T23:46:00Z`
- Use Delve debugger on macOS: install via brew, build iliad with debug flags, run with dlv exec ./iliad -- run --home ~/iliad-{1,2}. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-5) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_95424b01c803cb7b942957e08374b8af` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-5` `source_timestamp=2024-10-02T23:46:00Z`

## Sources

- `source_document_id`: `srcdoc_63839ac5d351533add9aa1535af1a240`
- `source_revision_id`: `srcrev_d70592364b1e7ce5f36332f30f6ccf1a`
- `source_url`: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14)
