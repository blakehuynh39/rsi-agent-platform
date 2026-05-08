---
title: "IPA Staking"
type: "system"
slug: "systems/ipa-staking"
freshness: "2024-06-18T16:59:00Z"
tags:
  - "devnet"
  - "ipa"
  - "staking"
owners: []
source_revision_ids:
  - "srcrev_cf9ce6b90d33b6b09532c5060d990aa6"
conflict_state: "none"
---

# IPA Staking

## Summary

IPA Staking allows any user to stake an ERC20 Stake Token to any IP Asset, claim staking rewards (sIP), and withdraw with a cooldown. The protocol admin defines a global reward emission rate and a percentage of rewards allocated to the IP owner. Deployed on devnet with WIP and IPAStaking contracts.

## Claims

- Any user can stake Stake Token (ERC20) to any IP Asset via the Staking Dashboard. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c) `source_document_id=srcdoc_db7f99193edf0561ccd7533f88f80ac8` `source_revision_id=srcrev_cf9ce6b90d33b6b09532c5060d990aa6` `chunk_id=srcchunk_b4b583a5179c483b7a93de1a374dc7e4` `native_locator=https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c` `source_timestamp=2024-06-18T16:59:00Z`
- An IPA pool allows multiple users to stake to one IPA. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c) `source_document_id=srcdoc_db7f99193edf0561ccd7533f88f80ac8` `source_revision_id=srcrev_cf9ce6b90d33b6b09532c5060d990aa6` `chunk_id=srcchunk_b4b583a5179c483b7a93de1a374dc7e4` `native_locator=https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c` `source_timestamp=2024-06-18T16:59:00Z`
- Any user can claim a staking reward (sIP) token from any IP Asset at any time via the Staking Dashboard. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c) `source_document_id=srcdoc_db7f99193edf0561ccd7533f88f80ac8` `source_revision_id=srcrev_cf9ce6b90d33b6b09532c5060d990aa6` `chunk_id=srcchunk_b4b583a5179c483b7a93de1a374dc7e4` `native_locator=https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c` `source_timestamp=2024-06-18T16:59:00Z`
- Users can withdraw their staking (WIP) token from any IP Asset at any time but with a cooldown delay. The dashboard shows an error if withdrawal is attempted within the cooldown period. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c) `source_document_id=srcdoc_db7f99193edf0561ccd7533f88f80ac8` `source_revision_id=srcrev_cf9ce6b90d33b6b09532c5060d990aa6` `chunk_id=srcchunk_b4b583a5179c483b7a93de1a374dc7e4` `native_locator=https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c` `source_timestamp=2024-06-18T16:59:00Z`
- Protocol Admin defines the reward emission rate (APY) at deployment time. The APY is the same for all IP Assets on devnet. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c) `source_document_id=srcdoc_db7f99193edf0561ccd7533f88f80ac8` `source_revision_id=srcrev_cf9ce6b90d33b6b09532c5060d990aa6` `chunk_id=srcchunk_b4b583a5179c483b7a93de1a374dc7e4` `native_locator=https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c` `source_timestamp=2024-06-18T16:59:00Z`
- IP Owner can receive/claim a reward (e.g., 1%) from staking on their IPA. The global reward percentage is defined by the protocol admin at deployment time. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c) `source_document_id=srcdoc_db7f99193edf0561ccd7533f88f80ac8` `source_revision_id=srcrev_cf9ce6b90d33b6b09532c5060d990aa6` `chunk_id=srcchunk_b4b583a5179c483b7a93de1a374dc7e4` `native_locator=https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c` `source_timestamp=2024-06-18T16:59:00Z`
- WIP contract (forked WETH) and WSTORY contract deployed to devnet. IPAssetStaking.sol deployed to devnet. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c) `source_document_id=srcdoc_db7f99193edf0561ccd7533f88f80ac8` `source_revision_id=srcrev_cf9ce6b90d33b6b09532c5060d990aa6` `chunk_id=srcchunk_b4b583a5179c483b7a93de1a374dc7e4` `native_locator=https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c` `source_timestamp=2024-06-18T16:59:00Z`
- WIP address: 0xCEb8483Ba7889f9af3B21f52B208cC1D9C188De5, IPAStaking address: 0x5d03d004D7Fa8D18F2EE9b53579f7BF8308d5C05. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c) `source_document_id=srcdoc_db7f99193edf0561ccd7533f88f80ac8` `source_revision_id=srcrev_cf9ce6b90d33b6b09532c5060d990aa6` `chunk_id=srcchunk_b4b583a5179c483b7a93de1a374dc7e4` `native_locator=https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c` `source_timestamp=2024-06-18T16:59:00Z`

## Open Questions

- Design Protocol level emission rate (APY) mechanism for all IPAssets (options: multiple stages, bonding curve, governance).
- Should IP Owner be able to configure the reward percentage to itself only once?
- Should the wrapped token be ERC20Permit to allow signatures when staking?

## Sources

- `source_document_id`: `srcdoc_db7f99193edf0561ccd7533f88f80ac8`
- `source_revision_id`: `srcrev_cf9ce6b90d33b6b09532c5060d990aa6`
- `source_url`: [Notion source](https://www.notion.so/Protocol-Devnet-Goals-IPA-Staking-and-Precompiles-f1f1e5ab7e1143488dfa2145b790eb4c)
