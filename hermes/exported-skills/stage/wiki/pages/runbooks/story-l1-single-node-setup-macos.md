---
title: "Story L1 Single Node Setup on macOS"
type: "runbook"
slug: "runbooks/story-l1-single-node-setup-macos"
freshness: "2024-10-02T23:48:00Z"
tags:
  - "geth"
  - "iliad"
  - "macos"
  - "node-setup"
  - "story-l1"
owners: []
source_revision_ids:
  - "srcrev_ef89487304eb28a087d0e96ebc25c45a"
conflict_state: "none"
---

# Story L1 Single Node Setup on macOS

## Summary

Step-by-step guide to set up a single-node private Story L1 network on macOS, including execution client (geth) and consensus client (iliad) configuration, genesis initialization, and debugging with Delve.

## Claims

- Create a geth and config folder for storing execution client configuration and data files using mkdir -p ~/geth/config. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_843bf9654d66160e300a8e5ccbaabd97` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1` `source_timestamp=2024-10-02T23:48:00Z`
- Place genesis.json and geth.toml files into ~/geth/config folder. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_843bf9654d66160e300a8e5ccbaabd97` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1` `source_timestamp=2024-10-02T23:48:00Z`
- In geth.toml, set DataDir to the path of the geth data folder (e.g., /Users/leeren/geth/data), set JWTSecret to the path of the jwtsecret file (e.g., /Users/leeren/geth/data/geth/jwtsecret), and set BootstrapNodes to an empty list. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_843bf9654d66160e300a8e5ccbaabd97` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1` `source_timestamp=2024-10-02T23:48:00Z`
- Clone the Story Protocol go-ethereum repository and build geth using make geth. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_843bf9654d66160e300a8e5ccbaabd97` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1` `source_timestamp=2024-10-02T23:48:00Z`
- Initialize the genesis block with ./build/bin/geth init --datadir="~/geth/data" ~/geth/config/genesis.json. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_843bf9654d66160e300a8e5ccbaabd97` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1` `source_timestamp=2024-10-02T23:48:00Z`
- Run geth with ./build/bin/geth --config ~/geth/config/geth.toml. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_843bf9654d66160e300a8e5ccbaabd97` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1` `source_timestamp=2024-10-02T23:48:00Z`
- To restart geth, remove the data directory and reinitialize: rm -rf ~/geth/data/* && ./build/bin/geth init --datadir="~/geth/data" ~/geth/config/genesis.json && ./build/bin/geth --config ~/geth/config/geth.toml. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_843bf9654d66160e300a8e5ccbaabd97` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-1` `source_timestamp=2024-10-02T23:48:00Z`
- Run go run scripts/pubkey_to_bech32.go to generate accAddr, valAddr, and evmAddr from the validator's base64 public key. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_3a001bf68434e9ccc22233917ff58530` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2` `source_timestamp=2024-10-02T23:48:00Z`
- In iliad/config/genesis.json, under app_state.genutil.gen_txs, keep only the first transaction and delete all others. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_3a001bf68434e9ccc22233917ff58530` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2` `source_timestamp=2024-10-02T23:48:00Z`
- In the remaining validator transaction, replace delegator_address with accAddr, validator_address with valAddr, and pubkey.key with the original public key value. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_3a001bf68434e9ccc22233917ff58530` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2` `source_timestamp=2024-10-02T23:48:00Z`
- Under app_state.auth.accounts, replace the first account address with accAddr and delete all other account entries. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_3a001bf68434e9ccc22233917ff58530` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2` `source_timestamp=2024-10-02T23:48:00Z`
- Under app_state.bank.balances, replace the first account address with accAddr and delete all other balance entries. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_3a001bf68434e9ccc22233917ff58530` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-2` `source_timestamp=2024-10-02T23:48:00Z`
- To re-run iliad, reset priv_validator_state.json to {"height":"0","round":0,"step":0}. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-3) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_f9c1437a70907a1bcbf1b22115a7ed48` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-3` `source_timestamp=2024-10-02T23:48:00Z`
- Restart geth, clear iliad/data, reset priv_validator_state, and run iliad with: rm -rf ~/iliad/data/* && echo '{"height":"0","round":0,"step":0}' > ~/iliad/data/priv_validator_state.json && ./iliad run --home ~/iliad. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-3) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_f9c1437a70907a1bcbf1b22115a7ed48` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-3` `source_timestamp=2024-10-02T23:48:00Z`
- For debugging, install Delve with brew install delve, build iliad with debug flags (go build -gcflags "all=-N -l"), and run with dlv exec ./iliad -- run --home ~/iliad. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-3) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_f9c1437a70907a1bcbf1b22115a7ed48` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-3` `source_timestamp=2024-10-02T23:48:00Z`
- A one-liner for re-running iliad in debug mode: rm -rf ~/iliad/data/* && echo '{"height":"0","round":0,"step":0}' > ~/iliad/data/priv_validator_state.json && dlv exec ./iliad -- run --home ~/iliad. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-3) `source_document_id=srcdoc_79616899ac642c02dce7ee42239b435e` `source_revision_id=srcrev_ef89487304eb28a087d0e96ebc25c45a` `chunk_id=srcchunk_f9c1437a70907a1bcbf1b22115a7ed48` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa#chunk-3` `source_timestamp=2024-10-02T23:48:00Z`

## Sources

- `source_document_id`: `srcdoc_79616899ac642c02dce7ee42239b435e`
- `source_revision_id`: `srcrev_ef89487304eb28a087d0e96ebc25c45a`
- `source_url`: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-One-Node-Setup-7415db64c729419cb8ec62ae323b12aa)
