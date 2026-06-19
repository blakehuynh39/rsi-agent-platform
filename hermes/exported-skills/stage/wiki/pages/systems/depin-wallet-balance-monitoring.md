---
title: "DePIN IP Registration Wallet Balance Monitoring"
type: "system"
slug: "systems/depin-wallet-balance-monitoring"
freshness: "2026-04-28T01:43:26Z"
tags:
  - "alerting"
  - "depin"
  - "monitoring"
  - "wallets"
owners: []
source_revision_ids:
  - "srcrev_214e885682741e6130ae7684fbb4e333"
  - "srcrev_2817e1e7df94097f57e804bcc7775437"
  - "srcrev_58c3f82f2cf39179362a7bf70e717221"
  - "srcrev_7fe6ad4e44a786767c1f941e42adc4a4"
  - "srcrev_8268d9ae116f52784a0bb3e2d9ad0539"
  - "srcrev_8996d7d24c1e8ed094aae03f4133fe64"
  - "srcrev_ccfe45c739705230efd27e068071d1ce"
  - "srcrev_fc86f13f4d97f7aa2cdc1b4936ac57ae"
conflict_state: "none"
---

# DePIN IP Registration Wallet Balance Monitoring

## Summary

Configuration for monitoring and alerting on DePIN IP registration wallet balances, including thresholds and planned improvements.

## Claims

- Production warning alert triggers when IP registration wallet balance falls below 0.005 native token over a 5-minute window, sends notifications to Slack and PagerDuty via alert-api-critical-platform, and repeats every 1 hour. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_214e885682741e6130ae7684fbb4e333` `chunk_id=srcchunk_0a7f718b2062bb2a4e508078c9259a36` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777336021.714469` `source_timestamp=2026-04-28T00:27:01Z`
- Staging warning alert triggers at the same threshold (< 0.005 over 5 minutes) but uses a lower-severity notification (alert-api-warning, Slack only) and repeats every 2 hours. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_214e885682741e6130ae7684fbb4e333` `chunk_id=srcchunk_0a7f718b2062bb2a4e508078c9259a36` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777336021.714469` `source_timestamp=2026-04-28T00:27:01Z`
- Production critical alert triggers when balance drops below 0.001 native token over a 2‑minute window, sends Slack+PagerDuty via alert-api-critical-platform, and repeats every 1 hour. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_214e885682741e6130ae7684fbb4e333` `chunk_id=srcchunk_0a7f718b2062bb2a4e508078c9259a36` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777336021.714469` `source_timestamp=2026-04-28T00:27:01Z`
- Staging critical alert mirrors the same condition (< 0.001 over 2 minutes) but notifies only via Slack through alert-api-warning and repeats every 2 hours. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_214e885682741e6130ae7684fbb4e333` `chunk_id=srcchunk_0a7f718b2062bb2a4e508078c9259a36` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777336021.714469` `source_timestamp=2026-04-28T00:27:01Z`
- The underlying Prometheus metric is `ip_registration_wallet_balance`, per wallet, with labels `environment` (prod or stage) and `wallet_address`; the expression uses `min by (environment, wallet_address)` to evaluate the balance for each individual wallet. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_58c3f82f2cf39179362a7bf70e717221` `chunk_id=srcchunk_609a93e3f40b302005e19203e42d90cf` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777337250.310769` `source_timestamp=2026-04-28T00:47:30Z`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_ccfe45c739705230efd27e068071d1ce` `chunk_id=srcchunk_664249c1fc2ca712aabec340917b3af1` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777336802.373389` `source_timestamp=2026-04-28T00:40:02Z`
- There is currently no alert for the main funding wallet (a MetaMask wallet on a personal computer). `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_8996d7d24c1e8ed094aae03f4133fe64` `chunk_id=srcchunk_15ae19a30437502b4675ce1408fbbf4b` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777336826.883949` `source_timestamp=2026-04-28T00:40:26Z`
- It was suggested to add a higher-level warning threshold (higher than 0.005) to provide a buffer for wallet refills before reaching critical levels. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_8268d9ae116f52784a0bb3e2d9ad0539` `chunk_id=srcchunk_beed88d3885b25d555d6eca6f3b7a8cf` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777336404.109679` `source_timestamp=2026-04-28T00:33:24Z`
- Automation for the main funding wallet is planned as a priority P0.5 task. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_2817e1e7df94097f57e804bcc7775437` `chunk_id=srcchunk_299bf48a43b32f8885e95b86a3d68400` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777337589.086799` `source_timestamp=2026-04-28T00:53:09Z`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_fc86f13f4d97f7aa2cdc1b4936ac57ae` `chunk_id=srcchunk_eaa7dde8da117c2041ec10e4cc239ce7` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777337598.515159` `source_timestamp=2026-04-28T00:53:18Z`
- Improvements to the fake IP registration process (2 wallets, random delay, standalone binary) are also scheduled as P0.5. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3c6e0846357b7077adf26b207c9946ec` `source_revision_id=srcrev_7fe6ad4e44a786767c1f941e42adc4a4` `chunk_id=srcchunk_535068ad53d4a8217549d9e56b0feb42` `native_locator=slack:C0AL7EKNHDF:1777336021.714469:1777340606.143779` `source_timestamp=2026-04-28T01:43:26Z`

## Open Questions

- How will the main wallet funding automation be implemented?
- What are the exact specifications for the improved fake IP registration (standalone binary, random delay, etc.)?
- What should the higher warning threshold be?

## Sources

- `source_document_id`: `srcdoc_3c6e0846357b7077adf26b207c9946ec`
- `source_revision_id`: `srcrev_7fe6ad4e44a786767c1f941e42adc4a4`
