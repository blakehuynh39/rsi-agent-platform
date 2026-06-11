---
title: "Broadcaster Gas Strategy"
type: "system"
slug: "systems/broadcaster-gas-strategy"
freshness: "2026-06-11T17:27:00Z"
tags:
  - "broadcaster"
  - "controller"
  - "gas"
  - "ip-budget"
  - "registration"
owners: []
source_revision_ids:
  - "srcrev_9a88dc873695c2901ee578a1c79c9cce"
conflict_state: "none"
---

# Broadcaster Gas Strategy

## Summary

The Broadcaster Gas Strategy controls IP spending to achieve a daily registration target using a closed-loop controller that modulates gas price via block fill. Admin sets daily IP budget and registration target via API; the system adjusts fees in real-time, with absolute gwei safety bounds.

## Claims

- The broadcaster gas strategy spends an admin-defined daily IP budget on an admin-defined daily registration target, burning as much of that spend as possible. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- Both targets are set by the admin through the admin frontend, persisted in storage, and applied immediately — they are API-managed values, not deploy-time configs. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- The broadcaster runs a closed-loop controller on two in-day counters: registrations confirmed today (regsSoFar) and $IP spent on gas today (spentSoFar). `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- The controller steers the live $IP spent per registration toward a rest-of-day setpoint: setpoint = (dailyIPBudget - spentSoFar) / (dailyRegistrationTarget - regsSoFar). `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- While measured spend/reg < setpoint → raise the gas price (via block fill); while measured spend/reg > setpoint → lower the gas price (via block fill). `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- Batch size drift, retries, and failed transactions are priced in automatically because they show up in spentSoFar. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- If spend/reg converges to the setpoint, hitting the registration target means spending the budget exactly. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- Each day the broadcaster ramps the base fee at 90–95% block fill, holds spend/reg at the setpoint by modulating fill around 50%, and stops registering once the daily registration target is hit — resuming the next day. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- Base fee is burned by the protocol, priority fee is paid as validator tip, and max fee is only the transaction's ceiling, not necessarily what is paid. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- The admin sets daily IP budget, daily registration target, and optional daily USD budget via GET/PUT /admin/broadcaster/targets. The frontend writes them via authenticated API to persistent storage; the broadcaster reads stored values and applies changes immediately. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- The daily USD budget, if set, is converted to an IP/day budget via the IP/USD price and overrides the IP budget. This accommodates USD-denominated revenue targets when IP price fluctuates. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_40d969fc48d76d53f6fac79ad511ed2b` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-1` `source_timestamp=2026-06-11T17:27:00Z`
- Story's base fee moves by (2f - 1) / 24 per block, so steering block fullness raises or lowers spend/reg toward the setpoint. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_be4420c994a25ac9c3b0aa751c6ca474` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:27:00Z`
- Absolute gwei backstops are operator-set safety bounds on the live base fee: BROADCASTER_BASE_FEE_PAUSE_GWEI pauses fresh work if the live base fee exceeds it; BROADCASTER_MAX_FEE_LIMIT_GWEI is the hard maxFeePerGas ceiling for fresh work and replacements. `claim:claim_1_13` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_be4420c994a25ac9c3b0aa751c6ca474` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:27:00Z`
- Throttle and pause states act on fresh work only: throttle admits new batches at a reduced rate (one claim in two), pause holds all new batches until a lever reopens. Same-nonce replacements always bypass the gate to clear stuck nonces. `claim:claim_1_14` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_be4420c994a25ac9c3b0aa751c6ca474` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:27:00Z`
- The controller loop on each poll tick calculates error = (measured spend/reg - setpoint) / setpoint; if |error| <= deadband, hold fill target; if error < -deadband, raise fill target (rate-limited) up to ramp mode; if error > +deadband, lower fill target down to decrease target. `claim:claim_1_15` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_be4420c994a25ac9c3b0aa751c6ca474` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:27:00Z`
- Block-utilization thresholds by mode: in raise-gas mode, throttle at 90% block fill and pause at 95%; in hold mode, throttle at controller output (modulated around 50%) and pause at controller output + 5%. `claim:claim_1_16` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_be4420c994a25ac9c3b0aa751c6ca474` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:27:00Z`
- The fee envelope on every fresh transaction is: maxPriorityFeePerGas = priority (0.1 gwei), maxFeePerGas = min(baseFeePerGas * 1.25 + priority, BROADCASTER_MAX_FEE_LIMIT_GWEI). The 1.25x headroom keeps transactions valid as base fee moves; the cap bounds worst-case envelope. Expected paid fee is live base fee + 0.1 gwei priority. `claim:claim_1_17` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_be4420c994a25ac9c3b0aa751c6ca474` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:27:00Z`
- Default policy: Daily IP budget is set via API+storage; Daily USD budget is unset by default; both are administered through the admin frontend and take immediate effect. `claim:claim_1_18` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2) `source_document_id=srcdoc_883b49cc46e4eb7ecc91abe5101b8f47` `source_revision_id=srcrev_9a88dc873695c2901ee578a1c79c9cce` `chunk_id=srcchunk_be4420c994a25ac9c3b0aa751c6ca474` `native_locator=https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa#chunk-2` `source_timestamp=2026-06-11T17:27:00Z`

## Sources

- `source_document_id`: `srcdoc_883b49cc46e4eb7ecc91abe5101b8f47`
- `source_revision_id`: `srcrev_9a88dc873695c2901ee578a1c79c9cce`
- `source_url`: [source](https://app.notion.com/p/Broadcaster-Gas-Modeling-37b051299a548060a733cf1309802daa)
