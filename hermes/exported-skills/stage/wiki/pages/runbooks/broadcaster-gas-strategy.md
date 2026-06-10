---
title: "Broadcaster Gas Strategy"
type: "runbook"
slug: "runbooks/broadcaster-gas-strategy"
freshness: "2026-06-10T20:45:00Z"
tags:
  - "broadcaster"
  - "eip-1559"
  - "gas"
  - "story"
owners:
  - "broadcaster-team"
source_revision_ids:
  - "srcrev_00d576c88b8c2103ccc5aecfced90183"
conflict_state: "none"
---

# Broadcaster Gas Strategy

## Summary

Controls transaction fees for the broadcaster to burn ~12K IP per day on 15M registrations while keeping priority fee low. Uses EIP-1559 base fee mechanics and configurable thresholds to regulate spending.

## Claims

- The daily registration target is 15,000,000. `claim:daily-registration-target` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- Gas consumed per registration is approximately 28,000 gas, assuming a batch of 50 registrations. `claim:gas-per-registration` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- The daily IP budget is 12,000 IP. `claim:daily-ip-budget` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- The target effective gas price is ~28 gwei to stay within the daily budget. `claim:target-effective-gas-price` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- The target base fee is ~27.9 gwei. `claim:target-base-fee` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- The target priority fee is 0.1 gwei. `claim:target-priority-fee` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- Story mainnet uses an ElasticityMultiplier of 2, meaning target block usage is 50% of the gas limit. `claim:story-mainnet-elasticity-multiplier` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- Story mainnet uses a BaseFeeChangeDenominator of 24, meaning the base fee adjusts gradually by block pressure. `claim:story-mainnet-base-fee-change-denominator` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- The approximate per-block base fee change ratio is (2 * gasUsed/gasLimit - 1) / 24. `claim:base-fee-change-formula` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_7f31256c58a755f18e509ff58850785f` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-10T20:45:00Z`
- The configured BROADCASTER_PRIORITY_FEE_WEI is 100,000,000 (0.1 gwei). `claim:broadcaster-priority-fee-wei` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`
- BROADCASTER_BASE_FEE_HEADROOM_BPS is 12,500, equating to 1.25x base-fee headroom. `claim:broadcaster-base-fee-headroom-bps` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`
- BROADCASTER_MAX_FEE_LIMIT_GWEI is 37 gwei, acting as an absolute hard ceiling for maxFeePerGas. `claim:broadcaster-max-fee-limit-gwei` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`
- BROADCASTER_BASE_FEE_THROTTLE_GWEI is 29 gwei; fresh work is throttled when base fee reaches this threshold. `claim:broadcaster-base-fee-throttle-gwei` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`
- BROADCASTER_BASE_FEE_PAUSE_GWEI is 32 gwei; fresh work is paused when base fee reaches this threshold. `claim:broadcaster-base-fee-pause-gwei` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`
- BROADCASTER_BASE_FEE_BAND_GWEI is 27 gwei; the block-fill hold band engages at this base fee. `claim:broadcaster-base-fee-band-gwei` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`
- BROADCASTER_BASE_FEE_POLL_SECONDS is 5 seconds; base-fee monitor cadence. `claim:broadcaster-base-fee-poll-seconds` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`
- BROADCASTER_SUBMISSION_GATE_THROTTLE_FACTOR is 0.5; when throttled, half of fresh-claim ticks are admitted. `claim:broadcaster-submission-gate-throttle-factor` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`
- At base fee 0 gwei, sent max fee 0.1 gwei, paid fee ~0.1 gwei, burned ~0 gwei. At 10 gwei: sent 12.6 gwei, paid ~10.1 gwei, burned ~10 gwei. At 27 gwei: sent 33.85 gwei, paid ~27.1 gwei, burned ~27 gwei, hold band engages. At 27.9 gwei: sent 34.975 gwei, paid ~28.0 gwei, burned ~27.9 gwei (target). At 29 gwei: sent 36.35 gwei, paid ~29.1 gwei, burned ~29 gwei, fresh work throttles. At >=32 gwei: sent 37 gwei, fresh work pauses. `claim:base-fee-behavior-table` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_00d576c88b8c2103ccc5aecfced90183` `chunk_id=srcchunk_9dd9dc25b7cf5791b95ef84c7cf7b7a9` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-10T20:45:00Z`

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_00d576c88b8c2103ccc5aecfced90183`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
