---
title: "Broadcaster Gas Modeling"
type: "system"
slug: "systems/broadcaster-gas-modeling"
freshness: "2026-06-11T20:13:00Z"
tags:
  - "broadcaster"
  - "budget"
  - "controller"
  - "ethereum"
  - "fees"
  - "gas"
  - "IP"
  - "registration"
owners: []
source_revision_ids:
  - "srcrev_92117d56efd4e0425da52e90739ea9ed"
conflict_state: "none"
---

# Broadcaster Gas Modeling

## Summary

A closed-loop controller manages the broadcaster's daily IP budget and registration target by modulating block fill to influence the base fee, with phases for ramp, hold, and stop.

## Claims

- The broadcaster operates on a configured daily IP budget and daily registration target, both managed through the admin frontend, persisted by the API, and applied immediately at runtime. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_a78b1ea7c5637ce32422eb0f377e9f2e` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T20:13:00Z`
- The controller steers the live IP spent per registration toward a remaining-budget-per-remaining-registration setpoint: setpoint = (dailyIPBudget - spentSoFar) / (dailyRegistrationTarget - regsSoFar). If measured spend/reg converges to this setpoint, hitting the registration target exactly exhausts the budget. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_a78b1ea7c5637ce32422eb0f377e9f2e` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T20:13:00Z`
- At midnight UTC the counters (regsSoFar, spentSoFar) and the submission gate reset. In USD mode, the day's IP budget is re-derived from the IP/USD price at that moment, held fixed until the next reset unless the admin changes the USD budget mid-day. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_72fb0a6c67d0ba62e046b76ef3a056cf` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T20:13:00Z`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_a78b1ea7c5637ce32422eb0f377e9f2e` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T20:13:00Z`
- The base fee is burned by the protocol, the priority fee is paid as a validator tip, and maxFeePerGas is only a ceiling on what a transaction can pay, not the amount actually paid. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_a78b1ea7c5637ce32422eb0f377e9f2e` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T20:13:00Z`
- The admin API already supports GET/PUT /api/admin/v1/broadcaster/gas-settings with fields target_mode, usd_per_day, ip_per_day, updated_by, updated_at, and a historical log. For the new design, registration_target_per_day will be added to this settings object, and live counters (regsSoFar, spentSoFar) plus the current setpoint will be exposed. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_a78b1ea7c5637ce32422eb0f377e9f2e` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T20:13:00Z`
- The daily operation proceeds in three phases: Ramp (raise-gas mode), Hold (modulate fill to keep spend/reg at setpoint), and Stop (no more fresh registrations once target is hit). `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_a78b1ea7c5637ce32422eb0f377e9f2e` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T20:13:00Z`
- During the Ramp phase, blocks fill at 90% (throttle) to 95% (pause), raising the base fee by approximately 3.33% per block. This lasts from the ~23 wei floor for a few hundred blocks (~tens of minutes). `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_72fb0a6c67d0ba62e046b76ef3a056cf` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T20:13:00Z`
- During the Hold phase, the fill target is selected from five fixed levels based on the percentage error from the setpoint: - Inside deadband (±2.5%) → 50% (base fee holds). - Slightly under budget (−2.5% to −15%) → 52% (fee drifts up ~+0.167%/block). - Slightly over budget (+2.5% to +15%) → 48% (fee drifts down ~−0.167%/block). - Far over (+>15%) → 30% (fee decays ~−1.67%/block). - Far under (<−15%) → 90–95% raise-gas mode. Hysteresis margins prevent flapping at boundaries. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_72fb0a6c67d0ba62e046b76ef3a056cf` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T20:13:00Z`
- When regsSoFar reaches the daily registration target, fresh registration stops for the rest of the day, although replacements may still clear in-flight nonces. The setpoint freezes when the remaining-registration denominator becomes too small. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_a78b1ea7c5637ce32422eb0f377e9f2e` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T20:13:00Z`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_72fb0a6c67d0ba62e046b76ef3a056cf` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T20:13:00Z`
- Fresh submissions pass through a single gate that can be Open, Throttle, or Pause. Replacements (fee bumps) always bypass the gate, so throttle/pause never strand an in-flight transaction, only regulating new work entry rate. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_72fb0a6c67d0ba62e046b76ef3a056cf` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T20:13:00Z`
- The controller consists of two cascaded loops: an outer loop (slow, ~minutes) sets the fill target from budget error using the setpoint; an inner loop (fast) adjusts the admission rate r to meet the fill target. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_72fb0a6c67d0ba62e046b76ef3a056cf` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T20:13:00Z`
- Default configuration values include:
- BROADCASTER_RAMP_THROTTLE_FILL_PCT = 90
- BROADCASTER_RAMP_PAUSE_FILL_PCT = 95
- BROADCASTER_CONTROLLER_DEADBAND_BPS = 250 (±2.5%)
- BROADCASTER_HOLD_FILL_MIN_PCT = 48
- BROADCASTER_HOLD_FILL_MAX_PCT = 52
- BROADCASTER_DECAY_FILL_PCT = 30
- BROADCASTER_CONTROLLER_OUTER_BPS = 1500 (±15%)
- BROADCASTER_SPEND_WINDOW_REGS = 100000
- BROADCASTER_PRIORITY_FEE_WEI = 100000000 (0.1 gwei)
- BROADCASTER_BASE_FEE_PAUSE_GWEI and BROADCASTER_MAX_FEE_LIMIT_GWEI are operator-set, not defaulted. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_7c0433313aba5f275ba560a48194bf20` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T20:13:00Z`
- The fee envelope for every fresh transaction uses: maxPriorityFeePerGas = priority (0.1 gwei), maxFeePerGas = min(baseFeePerGas * 1.25 + priority, BROADCASTER_MAX_FEE_LIMIT_GWEI). The expected paid fee is the live base fee + 0.1 gwei priority; the 1.25x headroom allows valid inclusion within a few blocks. `claim:claim_1_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_7c0433313aba5f275ba560a48194bf20` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T20:13:00Z`
- An absolute backstop, BROADCASTER_BASE_FEE_PAUSE_GWEI, pauses fresh work if the live base fee exceeds it, regardless of the controller state, while still allowing replacements to clear stuck nonces. `claim:claim_1_14` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_7c0433313aba5f275ba560a48194bf20` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T20:13:00Z`
- A hard ceiling, BROADCASTER_MAX_FEE_LIMIT_GWEI, caps the maxFeePerGas for both fresh work and replacements. `claim:claim_1_15` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_92117d56efd4e0425da52e90739ea9ed` `chunk_id=srcchunk_7c0433313aba5f275ba560a48194bf20` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T20:13:00Z`

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_92117d56efd4e0425da52e90739ea9ed`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
