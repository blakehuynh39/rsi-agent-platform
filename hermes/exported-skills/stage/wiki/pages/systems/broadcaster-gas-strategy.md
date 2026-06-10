---
title: "Broadcaster Gas Strategy"
type: "system"
slug: "systems/broadcaster-gas-strategy"
freshness: "2026-06-10T19:59:00Z"
tags:
  - "broadcaster"
  - "ethereum"
  - "gas"
  - "story-mainnet"
owners: []
source_revision_ids:
  - "srcrev_05ced0f070069397def4cdb26e77c105"
conflict_state: "none"
---

# Broadcaster Gas Strategy

## Summary

The broadcaster gas strategy aims to spend approximately 12,000 IP per day on 15 million registrations, targeting an effective gas price of ~28 gwei with the base fee at ~27.9 gwei and priority fee at 0.1 gwei. It utilizes EIP-1559 base fee mechanics and regulation mechanisms to control spending and maximize burn.

## Claims

- Daily registration target is 15,000,000 registrations. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`
- Daily IP budget is 12,000 IP. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`
- Target effective gas price is ~28 gwei, with base fee at ~27.9 gwei and priority fee at 0.1 gwei. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`
- Gas consumed per registration is ~28,000 gas. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`
- The budget per registration is 0.0008 IP, and at 28 gwei effective price, cost per registration is 0.000784 IP. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`
- Story mainnet uses EIP-1559 base fee adjustment with ElasticityMultiplier=2 and BaseFeeChangeDenominator=24. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`
- Per-block base fee change ratio is approximately (2 * gasUsed/gasLimit - 1) / 24. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`
- At 50% block fullness base fee is stable, at 100% fullness it increases ~4.167% per block, at 0% it decreases ~4.167%. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`
- The broadcaster regulation uses a submission gate with block utilization and fee thresholds: BROADCASTER_MAX_FEE_LIMIT_GWEI=37, BROADCASTER_BASE_FEE_THROTTLE_GWEI=29, BROADCASTER_BASE_FEE_PAUSE_GWEI=32, BROADCASTER_BASE_FEE_BAND_GWEI=27. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_c1d41b0ebca7c6fa627fa23177401f04` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T19:59:00Z`
- When base fee reaches 27 gwei, a hold band engages targeting 48%/52% block utilization; at 29 gwei fresh work is throttled; at ≥32 gwei fresh work pauses. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_c1d41b0ebca7c6fa627fa23177401f04` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T19:59:00Z`
- At the target base fee of 27.9 gwei, the expected paid fee is ~28.0 gwei (27.9 base + 0.1 priority). `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_c1d41b0ebca7c6fa627fa23177401f04` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T19:59:00Z`
- The broadcaster burns most of the spend because base fee is burned by the protocol, while only the priority fee goes to validators. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_05ced0f070069397def4cdb26e77c105` `chunk_id=srcchunk_449785670123e5edb9e8b9d3055066a6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T19:59:00Z`

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_05ced0f070069397def4cdb26e77c105`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
