---
title: "Broadcaster Gas Strategy"
type: "system"
slug: "systems/broadcaster-gas-strategy"
freshness: "2026-06-11T18:30:00Z"
tags:
  - "broadcaster"
  - "controller"
  - "EIP-1559"
  - "gas"
  - "registration"
owners: []
source_revision_ids:
  - "srcrev_47025b6e950c469858d27c2af778cc68"
conflict_state: "none"
---

# Broadcaster Gas Strategy

## Summary

Closed-loop controller that modulates block fill to steer IP spent per registration toward a setpoint derived from admin-set daily IP budget and daily registration target. Incorporates throttle, pause, and absolute gwei backstops to manage gas price.

## Claims

- The broadcaster gas strategy spends an admin-defined daily IP budget on an admin-defined daily registration target by burning as much of that spend as possible. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_28e530ffb6061ccad6e08761447a5b22` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T18:30:00Z`
- Admin sets daily IP budget, optional daily USD budget, and daily registration target through an authenticated API (admin frontend), persisted in storage, and applied immediately. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_28e530ffb6061ccad6e08761447a5b22` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T18:30:00Z`
- When USD budget is set, it is converted to an IP/day budget once every 24h at midnight UTC using the IP/USD price, overriding the IP budget for that day. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_28e530ffb6061ccad6e08761447a5b22` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T18:30:00Z`
- The controller uses two in-day counters: regsSoFar (registrations confirmed today) and spentSoFar ($IP spent on gas today). It steers spend/reg toward a rest-of-day setpoint: (dailyIPBudget - spentSoFar) / (dailyRegistrationTarget - regsSoFar). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_28e530ffb6061ccad6e08761447a5b22` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T18:30:00Z`
- If measured spend/reg < setpoint, it raises gas price via block fill; if > setpoint, it lowers gas price. Batch size drift, retries, and failed transactions are automatically priced in. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_28e530ffb6061ccad6e08761447a5b22` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T18:30:00Z`
- The daily cycle: ramps base fee at 90-95% block fill, then modulates fill around 50% to hold spend/reg at setpoint, and stops once daily registration target is hit. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_28e530ffb6061ccad6e08761447a5b22` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T18:30:00Z`
- Base fee is burned by the protocol, priority fee is paid as validator tip, max fee is only the transaction’s ceiling and not necessarily what we pay. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_28e530ffb6061ccad6e08761447a5b22` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T18:30:00Z`
- The controller cascades two loops: outer loop (slow) sets fill target from budget error; inner loop (fast) computes admission rate r (0…1) to hit that fill target, where r<1 throttles fresh work and r=0 pauses it. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_3f4735c3634a6665d5c227f550a6e4a1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T18:30:00Z`
- Outer loop error is (measured - setpoint) / setpoint, with measured as rolling-window $IP spent per registration. Fill targets: error < -15% → 90–95% raise mode; error > +15% → 30% decay mode; inside ±2.5% deadband → 50% equilibrium; slight under (-2.5% to -15%) → 52% drift up; slight over (+2.5% to +15%) → 48% drift down. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_3f4735c3634a6665d5c227f550a6e4a1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T18:30:00Z`
- All fill-target transitions use hysteresis to prevent flapping at boundaries (±2.5% and ±15%). `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_3f4735c3634a6665d5c227f550a6e4a1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T18:30:00Z`
- When regsSoFar hits the daily target, fresh registration stops. Replacements for in-flight nonces continue. The setpoint division is frozen as remaining registrations approach zero. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_3f4735c3634a6665d5c227f550a6e4a1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T18:30:00Z`
- Fresh submission gate uses most-restrictive-wins logic from block utilization controller and absolute gwei backstops. It has states Open, Throttle, Pause. Throttle admits new batches at continuous rate r; Pause holds all new batches. Replacements always bypass the gate. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_d18369a0d8bace0f85e7601857f803e6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T18:30:00Z`
- Two absolute safety backstops: BROADCASTER_BASE_FEE_PAUSE_GWEI — if the live base fee exceeds it, fresh work pauses; BROADCASTER_MAX_FEE_LIMIT_GWEI — hard maxFeePerGas ceiling for fresh work and replacements. `claim:claim_1_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_d18369a0d8bace0f85e7601857f803e6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T18:30:00Z`
- The fee envelope for fresh transactions: maxPriorityFeePerGas = 0.1 gwei (fixed); maxFeePerGas = min(baseFeePerGas * 1.25 + priority, BROADCASTER_MAX_FEE_LIMIT_GWEI). Expected paid fee is live base fee + 0.1 gwei priority. `claim:claim_1_14` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_d18369a0d8bace0f85e7601857f803e6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T18:30:00Z`
- Default config values: BROADCASTER_RAMP_THROTTLE_FILL_PCT=90, BROADCASTER_RAMP_PAUSE_FILL_PCT=95, BROADCASTER_CONTROLLER_DEADBAND_BPS=250 (2.5%), BROADCASTER_HOLD_FILL_MIN_PCT=48, BROADCASTER_HOLD_FILL_MAX_PCT=52, BROADCASTER_DECAY_FILL_PCT=30, BROADCASTER_CONTROLLER_OUTER_BPS=1500 (15%), BROADCASTER_SPEND_WINDOW_REGS=100000, BROADCASTER_BASE_FEE_PAUSE_GWEI operator-set, BROADCASTER_MAX_FEE_LIMIT_GWEI operator-set, BROADCASTER_PRIORITY_FEE_WEI=100000000 (0.1 gwei). `claim:claim_1_15` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_d18369a0d8bace0f85e7601857f803e6` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T18:30:00Z`
- GET /admin/broadcaster/targets returns current targets + live counters + current setpoint. PUT /admin/broadcaster/targets accepts dailyIpBudget, dailyUsdBudget, dailyRegistrationTarget. `claim:claim_1_16` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_47025b6e950c469858d27c2af778cc68` `chunk_id=srcchunk_28e530ffb6061ccad6e08761447a5b22` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T18:30:00Z`

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_47025b6e950c469858d27c2af778cc68`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
