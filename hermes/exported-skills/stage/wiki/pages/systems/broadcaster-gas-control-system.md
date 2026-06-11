---
title: "Broadcaster Gas Control System"
type: "system"
slug: "systems/broadcaster-gas-control-system"
freshness: "2026-06-11T19:28:00Z"
tags:
  - "broadcaster"
  - "closed-loop"
  - "controller"
  - "fee"
  - "gas"
owners: []
source_revision_ids:
  - "srcrev_cd57f843f2cba07ed38e6c2c94afa288"
conflict_state: "none"
---

# Broadcaster Gas Control System

## Summary

An automated gas strategy that spends an admin-defined daily IP budget on an admin-defined daily registration target by adjusting transaction fees through block fill modulation, using in-day counters to steer towards a rest-of-day setpoint.

## Claims

- The broadcaster spends an admin-defined daily IP budget on an admin-defined daily registration target, burning as much of that spend as possible. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_26c2eda97eac07b1d5aef27583bf83ed` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:28:00Z`
- Both the daily IP budget and the daily registration target are set by an admin through an API, persisted in storage, and applied immediately. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_26c2eda97eac07b1d5aef27583bf83ed` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:28:00Z`
- The broadcaster runs a closed-loop controller on two in-day counters: regsSoFar (registrations confirmed today) and spentSoFar ($IP spent on gas today). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_26c2eda97eac07b1d5aef27583bf83ed` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:28:00Z`
- The controller computes a rest-of-day setpoint as (dailyIPBudget - spentSoFar) / (dailyRegistrationTarget - regsSoFar). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_26c2eda97eac07b1d5aef27583bf83ed` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:28:00Z`
- The controller raises gas price via block fill when measured spend/reg is below the setpoint, and lowers it when above. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_26c2eda97eac07b1d5aef27583bf83ed` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:28:00Z`
- Batch size drift, retries, and failed transactions are automatically priced in via spentSoFar, requiring no separate estimation. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_26c2eda97eac07b1d5aef27583bf83ed` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:28:00Z`
- At midnight UTC, counters regsSoFar and spentSoFar reset, and if a USD/day budget is set, the day's IP budget is re-derived from the current IP/USD price. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_ec9b5974d923a78f41a3809bce67f0b8` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T19:28:00Z`
- Operation proceeds in three phases: Ramp (raise base fee via 90-95% block fill), Hold (modulate fill around 50% equilibrium to maintain spend/reg near setpoint), and Stop (cease fresh registration once daily target is hit). `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_ec9b5974d923a78f41a3809bce67f0b8` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T19:28:00Z`
- During Hold, the fill target is chosen from five discrete levels based on the deviation of spend/reg from the setpoint: 50% inside ±2.5% deadband, 52% for slightly under budget, 48% for slightly over budget, 30% for far over (>+15%), and 90-95% raise-gas mode for far under (<-15%). `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_ec9b5974d923a78f41a3809bce67f0b8` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T19:28:00Z`
- A hysteresis margin is applied at each fill-level transition to prevent flapping when spend/reg is near a boundary. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_ec9b5974d923a78f41a3809bce67f0b8` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T19:28:00Z`
- Absolute gwei backstops independent of the controller are applied: BROADCASTER_BASE_FEE_PAUSE_GWEI pauses fresh work if the live base fee exceeds it, and BROADCASTER_MAX_FEE_LIMIT_GWEI caps the maximum fee per gas on all transactions. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_0f945d3a4024a17b38d6bee09390b5f5` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T19:28:00Z`
- Fresh transactions are constructed with maxPriorityFeePerGas = 0.1 gwei and maxFeePerGas = min(baseFeePerGas * 1.25 + priority, BROADCASTER_MAX_FEE_LIMIT_GWEI). `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_0f945d3a4024a17b38d6bee09390b5f5` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T19:28:00Z`
- Replacements that clear a stuck nonce always bypass the submission gate. `claim:claim_1_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_ec9b5974d923a78f41a3809bce67f0b8` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T19:28:00Z`
- The admin API (`GET /api/admin/v1/broadcaster/gas-settings` and `PUT /api/admin/v1/broadcaster/gas-settings`) currently supports daily IP/USD budgets but lacks a `registration_target_per_day` field; exposure of live counters and current setpoint is also planned. `claim:claim_1_14` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_26c2eda97eac07b1d5aef27583bf83ed` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:28:00Z`
- An optional daily USD budget, when set, is converted to IP at midnight UTC and overrides the direct IP budget for that day. `claim:claim_1_15` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_cd57f843f2cba07ed38e6c2c94afa288` `chunk_id=srcchunk_0f945d3a4024a17b38d6bee09390b5f5` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T19:28:00Z`

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_cd57f843f2cba07ed38e6c2c94afa288`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
