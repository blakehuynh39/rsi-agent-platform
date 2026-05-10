---
title: "Proof of Liquidity"
type: "concept"
slug: "concepts/proof-of-liquidity"
freshness: "2025-02-06T16:00:00Z"
tags:
  - "berachain"
  - "blockchain"
  - "consensus"
  - "defi"
owners: []
source_revision_ids:
  - "srcrev_1c4c9656c91ee2458930c8011ef82ca1"
conflict_state: "none"
---

# Proof of Liquidity

## Summary

Proof of Liquidity (PoL) is a consensus mechanism used by Berachain that rewards liquidity provision rather than staking. It presents centralization risks and exposes validators to price volatility, while its liquid staking benefit is moot compared to PoS protocols like Lido. The system involves non-transferable BGT rewards, a gas token BERA, and a stablecoin HONEY, and draws parallels to Curve's veCRV model.

## Claims

- Proof of Liquidity presents major centralization risks and exposes validators to greater price volatility, at the small benefit of making stakes liquid. PoS achieves liquid stakes via non-consensus-level staking protocols like Lido and Rocket Pool, making PoL's main selling point moot. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-1) `source_document_id=srcdoc_f0bd16f211685e52e39d9c699e4ac209` `source_revision_id=srcrev_1c4c9656c91ee2458930c8011ef82ca1` `chunk_id=srcchunk_dffbe657e9d251ea0d66b0e3db887cc0` `native_locator=https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-1` `source_timestamp=2025-02-06T16:00:00Z`
- BGT is a non-transferable reward emission token for pools, redeemable 1:1 for BERA, the gas token. HONEY is a stablecoin. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-1) `source_document_id=srcdoc_f0bd16f211685e52e39d9c699e4ac209` `source_revision_id=srcrev_1c4c9656c91ee2458930c8011ef82ca1` `chunk_id=srcchunk_dffbe657e9d251ea0d66b0e3db887cc0` `native_locator=https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-1` `source_timestamp=2025-02-06T16:00:00Z`
- PoL rewards liquidity by emitting BGT proportionally to the liquidity of pools, similar to the Curve protocol, rather than using weighted randomization like PoS. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-1) `source_document_id=srcdoc_f0bd16f211685e52e39d9c699e4ac209` `source_revision_id=srcrev_1c4c9656c91ee2458930c8011ef82ca1` `chunk_id=srcchunk_dffbe657e9d251ea0d66b0e3db887cc0` `native_locator=https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-1` `source_timestamp=2025-02-06T16:00:00Z`
- Validators in PoL are analogous to veCRV aggregation protocols (Convex, Yearn, etc.), and LPs in PoL are analogous to LPs in Curve pools. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-1) `source_document_id=srcdoc_f0bd16f211685e52e39d9c699e4ac209` `source_revision_id=srcrev_1c4c9656c91ee2458930c8011ef82ca1` `chunk_id=srcchunk_dffbe657e9d251ea0d66b0e3db887cc0` `native_locator=https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-1` `source_timestamp=2025-02-06T16:00:00Z`
- Boost in PoL is equivalent to Curve's Gauge Boost, with a maximum of 2.5x normal reward, as per the Curve boost formula. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-2) `source_document_id=srcdoc_f0bd16f211685e52e39d9c699e4ac209` `source_revision_id=srcrev_1c4c9656c91ee2458930c8011ef82ca1` `chunk_id=srcchunk_2505870f055e8aab45c18053082484a9` `native_locator=https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-2` `source_timestamp=2025-02-06T16:00:00Z`
- In Berachain, securing the chain for reward can happen via any tokens, not just BERA, leading to less value accrual for BERA compared to ETH in Ethereum. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-3) `source_document_id=srcdoc_f0bd16f211685e52e39d9c699e4ac209` `source_revision_id=srcrev_1c4c9656c91ee2458930c8011ef82ca1` `chunk_id=srcchunk_8778b6c46de4d9209cd7c35d066deb2a` `native_locator=https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-3` `source_timestamp=2025-02-06T16:00:00Z`
- LPs in Berachain are exposed to impermanent/divergence loss and the price volatility of all tokens they provide liquidity in, unlike Ethereum where only ETH stakes secure the network. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-3) `source_document_id=srcdoc_f0bd16f211685e52e39d9c699e4ac209` `source_revision_id=srcrev_1c4c9656c91ee2458930c8011ef82ca1` `chunk_id=srcchunk_8778b6c46de4d9209cd7c35d066deb2a` `native_locator=https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-3` `source_timestamp=2025-02-06T16:00:00Z`
- The requirement to provide liquidity to become a validator (since BGT is non-transferable) naturally incentivizes large liquidity and disincentivizes small players, increasing centralization, with no reward randomization like Ethereum. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-3) `source_document_id=srcdoc_f0bd16f211685e52e39d9c699e4ac209` `source_revision_id=srcrev_1c4c9656c91ee2458930c8011ef82ca1` `chunk_id=srcchunk_8778b6c46de4d9209cd7c35d066deb2a` `native_locator=https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36#chunk-3` `source_timestamp=2025-02-06T16:00:00Z`

## Open Questions

- Can BGT become transferable through smart contract wallets, and what would be the impact?
- How does the BGT boost mechanism affect long-term tokenomics?
- What are the specific centralization risks inherited by Berachain from PoL?

## Sources

- `source_document_id`: `srcdoc_f0bd16f211685e52e39d9c699e4ac209`
- `source_revision_id`: `srcrev_1c4c9656c91ee2458930c8011ef82ca1`
- `source_url`: [Notion source](https://www.notion.so/Proof-of-Liquidity-e48df8cd80b448a2a74b2689bde40a36)
