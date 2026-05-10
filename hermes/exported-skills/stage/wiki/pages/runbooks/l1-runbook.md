---
title: "L1 Runbook"
type: "runbook"
slug: "runbooks/l1-runbook"
freshness: "2024-12-26T05:06:00Z"
tags:
  - "blockchain"
  - "node-management"
  - "runbook"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_cc9cd55812b31dd66673b43ba57859f3"
conflict_state: "none"
---

# L1 Runbook

## Summary

Collection of operational commands and procedures for managing Story Protocol L1 nodes, including peer checks, log tailing, validator queries, genesis hash retrieval, stake management, log filtering, systemd control, and cosmovisor testing.

## Claims

- Command to check iliad seed node’s peer: curl localhost:26657/net_info | jq '.result.peers.[].node_info.moniker' `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Tail logs in nodes: journalctl -u node-geth.service -f `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Get validator keys with filters: curl -X GET https://staking.odyssey.storyrpc.io/api/staking/validators | jq -c '[ .msg.validators.[] | select(.description.moniker | contains("Story")) ]' | jq '.[] | [.consensus_pubkey, .description]' `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Get delegation of a validator: curl -X GET https://staking.odyssey.storyrpc.io/api/staking/validators/storyvaloper1pdj0z84lau0l7vf2jl4qs7yggv48p3avy82spr/delegations | jq . `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Retrieve Partner Genesis Hash: curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x0", false],"id":1}' -H "Content-Type: application/json" https://rpc.partner.testnet.storyprotocol.net/ | jq '.result.hash' `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Retrieve Devnet Genesis Hash: curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x0", false],"id":1}' -H "Content-Type: application/json" https://rpc.devnet.storyprotocol.net | jq '.result.hash' `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Set Min stake: cast send --rpc-url https://testnet.storyrpc.io --private-key <TESTNET-KEY> 0xCCcCcC0000000000000000000000000000000001 "setMinStakeAmount(uint256)" <AMOUNT_IN_WEI> --legacy `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Verify min stake (value in ether units): cast call 0xCCcCcC0000000000000000000000000000000001 "minStakeAmount()" --rpc-url https://testnet.storyrpc.io | cast --to-dec | cast --from-wei `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Set Min redelegation: cast send --rpc-url https://testnet.storyrpc.io --private-key <TESTNET-KEY> 0xCCcCcC0000000000000000000000000000000001 "setMinRedelegateAmount(uint256)" <AMOUNT_IN_WEI> --legacy `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Verify min redelegation (value in ether units): cast call 0xCCcCcC0000000000000000000000000000000001 "minRedelegateAmount()" --rpc-url https://testnet.storyrpc.io | cast --to-dec | cast --from-wei `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Read logs and filter: journalctl -u node-story.service -p debug --since "10 seconds ago" --no-pager -f | grep -v -e "Delegation" -e "VerifyVoteExtension" -e "DEBU absent validator" -e "delegator rewards" `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Check status of node-story service: systemctl status node-story `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Restart node-story service: sudo systemctl restart node-story `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`
- Test cosmovisor: Initialize env variables: export DAEMON_NAME=client, export DAEMON_DATA_BACKUP_DIR=~/.story/backup, export DAEMON_HOME=~/Library/Story/story. Setup & run geth: rm -rf ~/Library/Story/geth && ./build/bin/geth --local --syncmode full. Setup story: rm -rf ~/Library/Story/story/* && client/client init --network local. Run cosmovisor: ~/cosmo... (truncated) `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce) `source_document_id=srcdoc_c786fbaa6e53b70ab372758a292ab819` `source_revision_id=srcrev_cc9cd55812b31dd66673b43ba57859f3` `chunk_id=srcchunk_cc8f50da18858ae9701b8d8b389b4e04` `native_locator=https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce` `source_timestamp=2024-12-26T05:06:00Z`

## Sources

- `source_document_id`: `srcdoc_c786fbaa6e53b70ab372758a292ab819`
- `source_revision_id`: `srcrev_cc9cd55812b31dd66673b43ba57859f3`
- `source_url`: [Notion source](https://www.notion.so/L1-runbook-48c35ab5c1444181a7738978d79293ce)
