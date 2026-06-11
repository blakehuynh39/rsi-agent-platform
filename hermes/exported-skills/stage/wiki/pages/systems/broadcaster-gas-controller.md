---
title: "Broadcaster Gas Controller"
type: "system"
slug: "systems/broadcaster-gas-controller"
freshness: "2026-06-11T19:55:00Z"
tags:
  - "broadcaster"
  - "controller"
  - "daily-budget"
  - "gas"
owners: []
source_revision_ids:
  - "srcrev_540c8b3bf250ce5a69a3c00d184e8c55"
conflict_state: "none"
---

# Broadcaster Gas Controller

## Summary

Design of the closed-loop gas controller that manages daily IP budget and registration target for the broadcaster.

## Claims

- The broadcaster aims to spend a configured daily IP budget while processing up to a daily registration target. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_56bc3a2b76145b7ed28b30eb2149fe35` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:55:00Z`
- Both budget and target are runtime settings managed through the admin frontend, persisted by the API, and applied immediately. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_56bc3a2b76145b7ed28b30eb2149fe35` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:55:00Z`
- The controller steers live IP spent per registration to a remaining-budget-per-remaining-registration setpoint: setpoint = (dailyIPBudget - spentSoFar) / (dailyRegistrationTarget - regsSoFar). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_56bc3a2b76145b7ed28b30eb2149fe35` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:55:00Z`
- Base fee is burned by the protocol; priority fee is paid as validator tip; maxFeePerGas is a ceiling, not the amount actually paid. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_56bc3a2b76145b7ed28b30eb2149fe35` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T19:55:00Z`
- At midnight UTC, counters and submission gate reset; in USD mode the day's IP budget is re-derived from the IP/USD price at that moment. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_d74aca1e42627d2e518ab1e96ef4740c` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T19:55:00Z`
- The controller operates in three phases: Ramp (fill at 90-95% to raise base fee), Hold (modulate fill around 50% using five levels with deadbands), and Stop (stop after reaching daily registration target). `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_d74aca1e42627d2e518ab1e96ef4740c` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T19:55:00Z`
- Fresh submissions pass through a gate with open, throttle, and pause states; replacements always bypass the gate to clear in-flight nonces. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_d74aca1e42627d2e518ab1e96ef4740c` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T19:55:00Z`
- Default configuration: BROADCASTER_RAMP_THROTTLE_FILL_PCT=90, BROADCASTER_RAMP_PAUSE_FILL_PCT=95, BROADCASTER_CONTROLLER_DEADBAND_BPS=250 (±2.5%), BROADCASTER_HOLD_FILL_MIN_PCT=48, BROADCASTER_HOLD_FILL_MAX_PCT=52, BROADCASTER_DECAY_FILL_PCT=30, BROADCASTER_CONTROLLER_OUTER_BPS=1500 (±15%), BROADCASTER_SPEND_WINDOW_REGS=100000, BROADCASTER_BASE_FEE_PAUSE_GWEI=operator-set, BROADCASTER_MAX_FEE_LIMIT_GWEI=operator-set, BROADCASTER_PRIORITY_FEE_WEI=100000000 (0.1 gwei). `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_60f1adf2bcf9a3d112faa5d313c9c540` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T19:55:00Z`
- Fee envelope: maxPriorityFeePerGas = 0.1 gwei; maxFeePerGas = min(baseFeePerGas * 1.25 + priority, BROADCASTER_MAX_FEE_LIMIT_GWEI). `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_60f1adf2bcf9a3d112faa5d313c9c540` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-3` `source_timestamp=2026-06-11T19:55:00Z`
- The outer controller loop sets the fill target from budget error; the inner loop modulates admission rate to reach that target, with gain K=0.25 as a starting point. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-4) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_540c8b3bf250ce5a69a3c00d184e8c55` `chunk_id=srcchunk_90c8158f460fc916343bdab964889115` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-4` `source_timestamp=2026-06-11T19:55:00Z`

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_540c8b3bf250ce5a69a3c00d184e8c55`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
