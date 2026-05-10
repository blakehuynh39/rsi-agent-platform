---
title: "Upgrade Tutorial: Geth (0.9.2 → 0.9.3)"
type: "runbook"
slug: "runbooks/upgrade-geth-0-9-2-to-0-9-3"
freshness: "2024-09-25T00:44:00Z"
tags:
  - "execution-client"
  - "geth"
  - "iliad"
  - "network-flag"
  - "upgrade"
owners: []
source_revision_ids:
  - "srcrev_c81da633f7c98beea96d9ae17abeabe9"
conflict_state: "none"
---

# Upgrade Tutorial: Geth (0.9.2 → 0.9.3)

## Summary

How to migrate from geth v0.9.2 with manual configuration to v0.9.3 using the baked-in --iliad network flag, and how to define custom network flags.

## Claims

- Node operators should spin up geth and story clients using the --iliad network flag, which bakes in sane config and genesis settings, eliminating the need for manual config files. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3#chunk-1) `source_document_id=srcdoc_9acd42d2ba4da0abfe6f886837ad6032` `source_revision_id=srcrev_c81da633f7c98beea96d9ae17abeabe9` `chunk_id=srcchunk_85ce4799bcbb64c024e55908228800c0` `native_locator=https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3#chunk-1` `source_timestamp=2024-09-25T00:44:00Z`
- With the network flags, you may override any configs, but you may not override the genesis settings. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3#chunk-1) `source_document_id=srcdoc_9acd42d2ba4da0abfe6f886837ad6032` `source_revision_id=srcrev_c81da633f7c98beea96d9ae17abeabe9` `chunk_id=srcchunk_85ce4799bcbb64c024e55908228800c0` `native_locator=https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3#chunk-1` `source_timestamp=2024-09-25T00:44:00Z`
- To add a custom --${NETWORK} flag in geth, you must: define Default${NETWORK}GenesisBlock in core/genesis.go, set const ${NETWORK}AllocData in core/genesis_alloc.go using the output of go run core/mkalloc.go ${NETWORK_GENESIS}.json, set ${NETWORK}Bootnodes in params/bootnodes.go, and set ${NETWORK}GenesisHash in params/config.go to the expected genesis hash (obtainable via eth.getBlock(0).hash). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3#chunk-1) `source_document_id=srcdoc_9acd42d2ba4da0abfe6f886837ad6032` `source_revision_id=srcrev_c81da633f7c98beea96d9ae17abeabe9` `chunk_id=srcchunk_85ce4799bcbb64c024e55908228800c0` `native_locator=https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3#chunk-1` `source_timestamp=2024-09-25T00:44:00Z`
- To upgrade from geth v0.9.2 to v0.9.3: stop the story and geth processes, switch to the new PR branch, run make geth (or copy the new binary), then start geth with --local --syncmode full --datadir pointing to the old data directory, and start story with --home unchanged. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3#chunk-2) `source_document_id=srcdoc_9acd42d2ba4da0abfe6f886837ad6032` `source_revision_id=srcrev_c81da633f7c98beea96d9ae17abeabe9` `chunk_id=srcchunk_9d7f025196b3be7a7bae20cba913f478` `native_locator=https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3#chunk-2` `source_timestamp=2024-09-25T00:44:00Z`

## Sources

- `source_document_id`: `srcdoc_9acd42d2ba4da0abfe6f886837ad6032`
- `source_revision_id`: `srcrev_c81da633f7c98beea96d9ae17abeabe9`
- `source_url`: [Notion source](https://www.notion.so/Upgrade-Tutorial-Geth-0-9-2-0-9-3-10b051299a5480f2a43fd2d207e750d3)
