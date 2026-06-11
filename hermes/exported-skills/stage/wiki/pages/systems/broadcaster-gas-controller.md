---
title: "Broadcaster Gas Controller"
type: "system"
slug: "systems/broadcaster-gas-controller"
freshness: "2026-06-11T16:56:00Z"
tags:
  - "broadcaster"
  - "controller"
  - "eip-1559"
  - "gas"
  - "story"
owners:
  - "Operations"
source_revision_ids:
  - "srcrev_c0ccfb325c5579acd50dcc943cf32c34"
conflict_state: "none"
---

# Broadcaster Gas Controller

## Summary

The broadcaster gas strategy uses a closed-loop controller to spend a daily IP budget on registrations by manipulating block fill and thus the base fee. Admin sets daily IP budget and registration targets via a frontend API; the controller calculates a rest-of-day setpoint and adjusts transaction submission rates accordingly. Safety backstops provide absolute gwei limits.

## Claims

- The broadcaster spends an admin-defined daily IP budget on an admin-defined daily registration target, burning as much of that spend as possible. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_ac0b388967754b58f8490a368e447368` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T16:56:00Z`
- Admin targets are set via an authenticated admin frontend API, persisted in storage, and applied immediately without redeployment. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_ac0b388967754b58f8490a368e447368` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T16:56:00Z`
- The controller steers live $IP spent per registration toward a rest-of-day setpoint calculated as: setpoint = (dailyIPBudget - spentSoFar) / (dailyRegistrationTarget - regsSoFar). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_ac0b388967754b58f8490a368e447368` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T16:56:00Z`
- The controller tracks two in-day counters: regsSoFar (registrations confirmed today) and spentSoFar ($IP spent on gas today, including burned base fee and priority tips). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_ac0b388967754b58f8490a368e447368` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T16:56:00Z`
- Batch size drift, retries, and failed transactions are automatically priced in through spentSoFar, eliminating the need for gas-per-registration modeling. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_ac0b388967754b58f8490a368e447368` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T16:56:00Z`
- Each day the broadcaster ramps the base fee at 90–95% block fill, holds spend/reg at the setpoint by modulating fill around 50%, and stops registering once the daily registration target is hit, resuming the next day. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_ac0b388967754b58f8490a368e447368` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T16:56:00Z`
- Base fee is burned by the protocol, priority fee is paid as validator tip, and max fee is only the transaction's ceiling, not necessarily what is paid. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_ac0b388967754b58f8490a368e447368` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T16:56:00Z`
- Admin API endpoints: GET /admin/broadcaster/targets returns current targets and setpoint; PUT /admin/broadcaster/targets accepts dailyIpBudget, dailyUsdBudget, and dailyRegistrationTarget. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_ac0b388967754b58f8490a368e447368` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T16:56:00Z`
- Fresh submission is regulated by a single gate with two independent levers: block utilization (controller output) and absolute gwei backstops; the most restrictive lever wins. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_da69b4cbddf0e65590ed29dc027a07e1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T16:56:00Z`
- Throttle reduces new batch claims to 50% rate; pause halts all new batches; both act only on fresh work, while same-nonce replacements bypass the gate to avoid stranding in-flight transactions. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_da69b4cbddf0e65590ed29dc027a07e1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T16:56:00Z`
- The controller loop raises or lowers fill target in rate-limited steps based on error = (measured - setpoint)/setpoint, using a deadband; when error < -deadband it ramps to 90%/95% fill, when > +deadband it lowers fill, and when within deadband it holds. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_da69b4cbddf0e65590ed29dc027a07e1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T16:56:00Z`
- Absolute gwei backstops include BROADCASTER_BASE_FEE_PAUSE_GWEI (pauses fresh work when live base fee exceeds it) and BROADCASTER_MAX_FEE_LIMIT_GWEI (caps maxFeePerGas for all transactions). `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_da69b4cbddf0e65590ed29dc027a07e1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T16:56:00Z`
- Transaction fee envelope: maxPriorityFeePerGas = 0.1 gwei; maxFeePerGas = min(baseFeePerGas * 1.25 + 0.1, BROADCASTER_MAX_FEE_LIMIT_GWEI). `claim:claim_1_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_c0ccfb325c5579acd50dcc943cf32c34` `chunk_id=srcchunk_da69b4cbddf0e65590ed29dc027a07e1` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T16:56:00Z`

## Open Questions

- What are the precise default values for BROADCASTER_BASE_FEE_PAUSE_GWEI and BROADCASTER_MAX_FEE_LIMIT_GWEI?
- What values are used for the deadband and rate-limit step sizes?

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_c0ccfb325c5579acd50dcc943cf32c34`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
