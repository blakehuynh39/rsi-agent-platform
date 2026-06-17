---
title: "IPTokenStaking Ownership and Withdrawal Rules"
type: "system"
slug: "systems/iptoken-staking-ownership-rules"
freshness: "2026-06-17T02:58:00Z"
tags:
  - "aeneid"
  - "contract"
  - "staking"
  - "tokenomics"
owners: []
source_revision_ids:
  - "srcrev_f6a187f4c0bfa95cc579a687c2b56f27"
conflict_state: "none"
---

# IPTokenStaking Ownership and Withdrawal Rules

## Summary

Rules governing stake ownership and withdrawal as implemented by the IPTokenStaking precompile on Aeneid. Stake is burned on EVM, accounting on consensus layer. Delegator (not payer) owns stake and receives unstaked funds unless withdrawal address changed. Key behaviors: stake/delegator assignment, unstake disbursement, default withdrawal address setting, and no overwrite of user-set addresses. Applied to genesis analysis of addresses A, C, F.

## Claims

- IPTokenStaking precompile address is 0xCccccc0000000000000000000000000000000001. `claim:claim_2_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_f6a187f4c0bfa95cc579a687c2b56f27` `chunk_id=srcchunk_258b0954f1d498ac4184e4d184b07161` `native_locator=https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1` `source_timestamp=2026-06-17T02:58:00Z`
- Staked tokens are burned on the EVM; `_stake()` transfers to address(0). Precompile EVM balance is 1 wei. Real accounting on consensus layer (x/staking + x/evmstaking). `claim:claim_2_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_f6a187f4c0bfa95cc579a687c2b56f27` `chunk_id=srcchunk_258b0954f1d498ac4184e4d184b07161` `native_locator=https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1` `source_timestamp=2026-06-17T02:58:00Z`
- The delegator owns the stake, not the payer. `stake` sets delegator to msg.sender; `stakeOnBehalf` sets delegator to the explicit delegator argument. `claim:claim_2_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_f6a187f4c0bfa95cc579a687c2b56f27` `chunk_id=srcchunk_258b0954f1d498ac4184e4d184b07161` `native_locator=https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1` `source_timestamp=2026-06-17T02:58:00Z`
- Unstake pays the delegator's withdrawal address. On first deposit, consensus layer sets withdrawal and reward addresses to the delegator's own address and never overwrites a user-set one. Only setWithdrawalAddress/setRewardsAddress can change these. `claim:claim_2_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_f6a187f4c0bfa95cc579a687c2b56f27` `chunk_id=srcchunk_258b0954f1d498ac4184e4d184b07161` `native_locator=https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1` `source_timestamp=2026-06-17T02:58:00Z`
- Only the delegator or an approved operator (via setOperator) can unstake. All observed staking positions are Flexible (stakingPeriod=0, delegationId=0) and unstakable anytime. `claim:claim_2_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1) `source_document_id=srcdoc_ccb6ef3f6922ad4a55322dedcc057deb` `source_revision_id=srcrev_f6a187f4c0bfa95cc579a687c2b56f27` `chunk_id=srcchunk_258b0954f1d498ac4184e4d184b07161` `native_locator=https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0#chunk-1` `source_timestamp=2026-06-17T02:58:00Z`

## Related Pages

- `aeneid-genesis-token-snapshot-2026-06-09`

## Sources

- `source_document_id`: `srcdoc_ccb6ef3f6922ad4a55322dedcc057deb`
- `source_revision_id`: `srcrev_f6a187f4c0bfa95cc579a687c2b56f27`
- `source_url`: [source](https://app.notion.com/p/Aeneid-Genesis-Native-Token-Allocation-Flow-Staking-Snapshot-37b051299a5481159765ed594e4531d0)
