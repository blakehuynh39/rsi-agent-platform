---
title: "IPTokenStaking Contract Ownership and Delegation Rules"
type: "concept"
slug: "concepts/iptokenstaking-contract-ownership-delegation-rules"
freshness: "2026-06-10T02:07:00Z"
tags:
  - "consensus"
  - "delegation"
  - "evm"
  - "precompile"
  - "staking"
owners:
  - "RSI Wiki Compiler"
source_revision_ids:
  - "srcrev_40adbdbc7e043cf9a286ddf8c7fa2e63"
conflict_state: "none"
---

# IPTokenStaking Contract Ownership and Delegation Rules

## Summary

Explanation of how the IPTokenStaking precompile (0xCccccc0000000000000000000000000000000001) assigns ownership of staked tokens, differentiating between the delegator and the payer, and the withdrawal address behavior.

## Claims

- Staking via IPTokenStaking burns tokens on the EVM; real accounting lives on the consensus layer (x/staking + x/evmstaking). `claim:claim_2_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_40adbdbc7e043cf9a286ddf8c7fa2e63` `chunk_id=srcchunk_ca07f0f841e6a3aa32f4fa98cd3d8b7a` `native_locator=https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0` `source_timestamp=2026-06-10T02:07:00Z`
- The delegator owns the stake, not the payer. stake() sets delegator = msg.sender; stakeOnBehalf() uses the explicit delegator arg. `claim:claim_2_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_40adbdbc7e043cf9a286ddf8c7fa2e63` `chunk_id=srcchunk_ca07f0f841e6a3aa32f4fa98cd3d8b7a` `native_locator=https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0` `source_timestamp=2026-06-10T02:07:00Z`
- Unstake pays the delegator's withdrawal address. On first deposit, the consensus layer sets withdraw and reward address to the delegator's own address and never overwrites a user-set one. Only setWithdrawalAddress/setRewardsAddress can change it. `claim:claim_2_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_40adbdbc7e043cf9a286ddf8c7fa2e63` `chunk_id=srcchunk_ca07f0f841e6a3aa32f4fa98cd3d8b7a` `native_locator=https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0` `source_timestamp=2026-06-10T02:07:00Z`
- None of the operational wallets A, C, F called setWithdrawalAddress or setRewardsAddress, so every position pays out to its delegator. `claim:claim_2_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_40adbdbc7e043cf9a286ddf8c7fa2e63` `chunk_id=srcchunk_ca07f0f841e6a3aa32f4fa98cd3d8b7a` `native_locator=https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0` `source_timestamp=2026-06-10T02:07:00Z`

## Related Pages

- `aeneid-genesis-ip-allocation-staking-snapshot`
- `aeneid-genesis-wallets`

## Sources

- `source_document_id`: `srcdoc_ccb6ef3f6922ad4a55322dedcc057deb`
- `source_revision_id`: `srcrev_40adbdbc7e043cf9a286ddf8c7fa2e63`
- `source_url`: [source](https://app.notion.com/p/Aeneid-Genesis-IP-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0)
