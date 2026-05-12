---
title: "Halo ETH Withdrawal Flow"
type: "system"
slug: "systems/halo-eth-withdraw-flow"
freshness: "2024-09-19T09:16:00Z"
tags:
  - "cosmos-sdk"
  - "evmengine"
  - "evmstaking"
  - "halo"
  - "withdrawal"
owners: []
source_revision_ids:
  - "srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60"
conflict_state: "none"
---

# Halo ETH Withdrawal Flow

## Summary

Describes the end-to-end flow for ETH withdrawals in Halo, covering full withdrawal requests through the unbonding queue, processing of mature unbondings, and final transfer to the execution client.

## Claims

- A user calls `withdraw` on a custom staking contract via the execution client to trigger a smart contract withdraw event. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-1) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_fca8d6a6f463e55135fae313329c01a1` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-1` `source_timestamp=2024-09-19T09:16:00Z`
- Once a proposer is selected, the consensus client invokes PrepareProposal to collect the geth execution payload. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-1) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_fca8d6a6f463e55135fae313329c01a1` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-1` `source_timestamp=2024-09-19T09:16:00Z`
- PrepareProposal is setup in `app/app.go` within the `newApp` function via `bapp.SetPrepareProposal(app.EVMEngKeeper.PrepareProposal)`. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-1) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_fca8d6a6f463e55135fae313329c01a1` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-1` `source_timestamp=2024-09-19T09:16:00Z`
- CometBFT triggers the ABCI call via the proxy app, which calls `PrepareProposal` in the ABCI wrapper, ultimately calling the custom EVM Engine Keeper's `PrepareProposal`. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-1) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_fca8d6a6f463e55135fae313329c01a1` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-1` `source_timestamp=2024-09-19T09:16:00Z`
- ProcessProposal is invoked by the TM proxy app, handled in `app/app.go` via the baseapp config, and routed through `app/prouter.go`. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-2) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_6886a8e38de53edd42d69493a09aaeef` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-2` `source_timestamp=2024-09-19T09:16:00Z`
- The EVM engine keeper calls `RegisterProposalService` whose implementation is in `evmengine/keeper/proposal_server.go`. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-2) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_6886a8e38de53edd42d69493a09aaeef` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-2` `source_timestamp=2024-09-19T09:16:00Z`
- The `Finalize` function in the EVM engine keeper (called after the actual `finalizeBlock` internal call) dequeues eligible withdrawals. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-3) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_5aa6a3a59a6c58dc19654a245775ce61` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-3` `source_timestamp=2024-09-19T09:16:00Z`
- Based on the Withdraw event, the depositor and validator addresses are unpacked and the `x/staking` module's `Undelegate` function is called to initiate a full withdrawal. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-4) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_d802a1ddd8f9154f1efe1a8cc80005d4` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-4` `source_timestamp=2024-09-19T09:16:00Z`
- During the process, the account token is burned. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-4) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_5aa6a3a59a6c58dc19654a245775ce61` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-4` `source_timestamp=2024-09-19T09:16:00Z`
- Full withdrawals check mature UnbondingDelegation entries and insert requests into a withdraw queue; partial withdrawals calculate rewards, check eligibility, withdraw rewards, and also insert into the queue up to a sweep limit. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-5) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_5aa6a3a59a6c58dc19654a245775ce61` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-5` `source_timestamp=2024-09-19T09:16:00Z`
- In the prepareProposal stage, the evmengine pulls entries from the withdraw queue until a quota is reached and transfers them to the execution client. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-5) `source_document_id=srcdoc_0cf8113e4f62be9d726ee578783d7e0f` `source_revision_id=srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60` `chunk_id=srcchunk_5aa6a3a59a6c58dc19654a245775ce61` `native_locator=https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a#chunk-5` `source_timestamp=2024-09-19T09:16:00Z`

## Sources

- `source_document_id`: `srcdoc_0cf8113e4f62be9d726ee578783d7e0f`
- `source_revision_id`: `srcrev_b34c328fbf0ea488f1c89bf6bbf2ac60`
- `source_url`: [Notion source](https://www.notion.so/Halo-ETH-Withdraw-flows-a6bfbdce83484147a2459fad284e433a)
