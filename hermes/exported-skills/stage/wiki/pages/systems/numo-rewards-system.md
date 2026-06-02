---
title: "Numo Rewards System"
type: "system"
slug: "systems/numo-rewards-system"
freshness: "2026-06-02T00:25:00Z"
tags:
  - "multipliers"
  - "payouts"
  - "rewards"
  - "wallet"
owners:
  - "PIP Labs"
source_revision_ids:
  - "srcrev_f4f0737c3faeb96f40356202cf217726"
conflict_state: "none"
---

# Numo Rewards System

## Summary

Numo's reward mechanism includes points, USD payouts, a multiplier system, leaderboard, and wallet with Stripe Connect withdrawals.

## Claims

- Advance (15%) is paid at submission time and remainder (85%) when a reviewer accepts. Rejected submissions forfeit remainder but keep advance. `claim:claim_3_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-2) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_273097bce29d7fb26792285db79444ef` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-2` `source_timestamp=2026-06-02T00:25:00Z`
- Bonus payment is calculated as base_rate × (effective_multiplier − 1.0) × α, with α = 0.25 dampening coefficient. `claim:claim_3_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-2) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_273097bce29d7fb26792285db79444ef` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-2` `source_timestamp=2026-06-02T00:25:00Z`
- Total multiplier is capped at 2.0×, composed of permanent boosts (e.g., 100 verified submissions → 1.10×) and temporary boosts (highest single wins: 20+ submissions today → 1.20×, 7-day streak → 1.15×, Poseidon Season 1 handoff → 1.05× to 2.00× with 4-week full value + 8-week linear decay). `claim:claim_3_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-2) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_273097bce29d7fb26792285db79444ef` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-2` `source_timestamp=2026-06-02T00:25:00Z`
- Wallet features include available balance, rolling 7-day earnings, advance/remainder/bonus breakdown, lifetime earnings. Withdrawal gated by withdrawal_threshold_usd. Routes: /wallet, /wallet/earnings-history, /wallet/multiplier-tasks. `claim:claim_3_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-2) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_273097bce29d7fb26792285db79444ef` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-2` `source_timestamp=2026-06-02T00:25:00Z`
- Payouts use Stripe Connect; withdrawals require country selection from allowed list (US, CA, GB, VN, AU, IE, DE, FR, JP, SG), admin review, and then Stripe payout. `claim:claim_3_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_7d4b83c27c1d9a8c093d00dfe6826e8c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-4` `source_timestamp=2026-06-02T00:25:00Z`

## Related Pages

- `numo-architecture`
- `numo-platform-overview`

## Sources

- `source_document_id`: `srcdoc_5b2f0266ee274de6a7ca742845dfbcd4`
- `source_revision_id`: `srcrev_f4f0737c3faeb96f40356202cf217726`
- `source_url`: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3)
