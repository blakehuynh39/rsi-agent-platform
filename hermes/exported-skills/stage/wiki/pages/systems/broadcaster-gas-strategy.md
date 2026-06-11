---
title: "Broadcaster Gas Strategy"
type: "system"
slug: "systems/broadcaster-gas-strategy"
freshness: "2026-06-11T17:51:00Z"
tags:
  - "broadcaster"
  - "gas"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_1c79d429562aec60c6aee788c983b1c4"
conflict_state: "none"
---

# Broadcaster Gas Strategy

## Summary

The broadcaster spends an admin-defined daily IP budget on a daily registration target, using a closed-loop controller that adjusts block fill to steer actual spend/registration toward a continuously recalculated setpoint. This ensures the budget is burned by the time the target is reached, automatically pricing in batch size drift, retries, and failures.

## Claims

- The broadcaster gas strategy spends an admin-defined daily IP budget on an admin-defined daily registration target, burning as much of that spend as possible. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_70c81832377cfc1024b98eebd2dc32d0` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:51:00Z`
- The admin targets (daily IP budget, daily registration target, optional daily USD budget) are set via an authenticated API (`GET /admin/broadcaster/targets` and `PUT /admin/broadcaster/targets`), persisted in storage, and applied immediately. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_70c81832377cfc1024b98eebd2dc32d0` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:51:00Z`
- The broadcaster runs a closed‑loop controller on two in‑day counters: `regsSoFar` (registrations confirmed today) and `spentSoFar` ($IP spent on gas today). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_70c81832377cfc1024b98eebd2dc32d0` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:51:00Z`
- The controller steers the live $IP spent per registration toward a continuously recalculated setpoint: `setpoint = (dailyIPBudget - spentSoFar) / (dailyRegistrationTarget - regsSoFar)`. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_70c81832377cfc1024b98eebd2dc32d0` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:51:00Z`
- Fees: base fee is burned by the protocol, priority fee is paid as a validator tip, and max fee is only the transaction’s ceiling (not necessarily what is paid). `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_70c81832377cfc1024b98eebd2dc32d0` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:51:00Z`
- Fresh submission is regulated through a single submission gate fed by two levers: block utilization (controller output) and absolute gwei backstops (operator‑set safety bounds). The gate states are Open, Throttle (admission rate r ∈ (0,1)), and Pause; replacements always bypass the gate. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_5abe9ecd811ebddf9b9047fad82e3e0a` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:51:00Z`
- The controller comprises two loops: an outer loop (~minutes) sets a fill target based on budget error, and an inner loop (~seconds) modulates a continuous admission rate r to keep observed block fill at the target, using integral gain K=0.25 per tick (BROADCASTER_FILL_GAIN_BPS=2500). `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_5abe9ecd811ebddf9b9047fad82e3e0a` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:51:00Z`
- When the rolling $IP spent per registration is within a ±5% deadband of the setpoint, the fill target is held around 50% (modulated within 48–52%). If spend is far above the setpoint (> +15%), the fill target drops to 30%; if far below (< -15%), it rises to 90% (or pauses at 95%). `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_3b5c1f9609aff1f1130fb99e11d5e366` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T17:51:00Z`
- Key configuration parameters and defaults: CONTROLLER_DEADBAND_BPS=500 (5%), MAX_STEP_PCT=2, HOLD_FILL_MIN_PCT=48, HOLD_FILL_MAX_PCT=52, DECAY_FILL_PCT=30, OUTER_BPS=1500 (±15%), SPEND_WINDOW_REGS=100000, PRIORITY_FEE_WEI=100000000 (0.1 gwei), BASE_FEE_HEADROOM_BPS=12500 (1.25x), BASE_FEE_POLL_SECONDS=5, FILL_GAIN_BPS=2500, FILL_WINDOW_BLOCKS=5. BASE_FEE_PAUSE_GWEI and MAX_FEE_LIMIT_GWEI are operator-set. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_3b5c1f9609aff1f1130fb99e11d5e366` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T17:51:00Z`
- Once `regsSoFar` reaches the daily registration target, fresh registration stops until the next day; replacements still clear in‑flight nonces, and the setpoint division is frozen as the gate closes. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_1c79d429562aec60c6aee788c983b1c4` `chunk_id=srcchunk_5abe9ecd811ebddf9b9047fad82e3e0a` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:51:00Z`

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_1c79d429562aec60c6aee788c983b1c4`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
