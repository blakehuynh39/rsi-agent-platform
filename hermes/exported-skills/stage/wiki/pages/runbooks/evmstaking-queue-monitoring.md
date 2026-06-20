---
title: "EVMStaking Queue Depth Monitoring"
type: "runbook"
slug: "runbooks/evmstaking-queue-monitoring"
freshness: "2026-06-02T08:27:03Z"
tags:
  - "evmstaking"
  - "grafana"
  - "monitoring"
  - "seneca"
owners: []
source_revision_ids:
  - "srcrev_2d55be99fb3409ee11d659b85fae55e4"
  - "srcrev_cf426d2e82c2432d80e502f12061a5ae"
  - "srcrev_d4d152e07a27d939c084f230b68514ef"
  - "srcrev_d69dbb5796ecdd77e9b24208c9894d9d"
conflict_state: "none"
---

# EVMStaking Queue Depth Monitoring

## Summary

Grafana dashboard and alert configuration for monitoring EVMStaking withdrawal and reward queue depths, created for the Seneca validator reduction and mainnet-only ongoing use.

## Claims

- A request was made to set up Grafana monitoring for the evmstaking withdrawal queue for the upcoming Seneca event. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ca767a4c8ff6c3467561c0c5270d60e7` `source_revision_id=srcrev_2d55be99fb3409ee11d659b85fae55e4` `chunk_id=srcchunk_a25e58fa96b936e035e1649033d38f56` `native_locator=slack:C0547N89JUB:1780382680.121369:1780382680.121369` `source_timestamp=2026-06-02T06:44:40Z`
- The withdrawal queue is designed to drain quickly at a fixed max_withdrawal_per_block = 32 entries per block, with priority over the reward withdrawal queue. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ca767a4c8ff6c3467561c0c5270d60e7` `source_revision_id=srcrev_d69dbb5796ecdd77e9b24208c9894d9d` `chunk_id=srcchunk_26aa66c022898fd315b583d74c381897` `native_locator=slack:C0547N89JUB:1780382680.121369:1780382685.219729` `source_timestamp=2026-06-02T06:44:45Z`
- Existing metrics EVMStakingWithdrawalQueueDepth and EVMStakingRewardQueueDepth are already exposed and scraped via Thanos, with mainnet showing 19 withdrawal series. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ca767a4c8ff6c3467561c0c5270d60e7` `source_revision_id=srcrev_d4d152e07a27d939c084f230b68514ef` `chunk_id=srcchunk_b5d88b0e94dcc605bdfcf232db8041e4` `native_locator=slack:C0547N89JUB:1780382680.121369:1780383012.201469` `source_timestamp=2026-06-02T06:50:12Z`
- A Grafana dashboard was created and later restricted to mainnet-only nodes by filtering on network="mainnet", removing Aeneid testnet data. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ca767a4c8ff6c3467561c0c5270d60e7` `source_revision_id=srcrev_cf426d2e82c2432d80e502f12061a5ae` `chunk_id=srcchunk_bc7fc74e40c810d653f87a07d9c9c112` `native_locator=slack:C0547N89JUB:1780382680.121369:1780388823.994259` `source_timestamp=2026-06-02T08:27:03Z`
- Proposed alert thresholds: withdrawal >50 for 5m (warning), >200 for 5m (critical); reward >100 for 5m (warning), >500 for 5m (critical). `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ca767a4c8ff6c3467561c0c5270d60e7` `source_revision_id=srcrev_d4d152e07a27d939c084f230b68514ef` `chunk_id=srcchunk_b5d88b0e94dcc605bdfcf232db8041e4` `native_locator=slack:C0547N89JUB:1780382680.121369:1780383012.201469` `source_timestamp=2026-06-02T06:50:12Z`
- Alert provisioning via Grafana API is blocked due to token permissions; a team member with access or the monitoring config repo path is needed to land the alert rules. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ca767a4c8ff6c3467561c0c5270d60e7` `source_revision_id=srcrev_d4d152e07a27d939c084f230b68514ef` `chunk_id=srcchunk_b5d88b0e94dcc605bdfcf232db8041e4` `native_locator=slack:C0547N89JUB:1780382680.121369:1780383012.201469` `source_timestamp=2026-06-02T06:50:12Z`

## Open Questions

- Who can provision the Grafana alert rules (either via API or config repo)?

## Sources

- `source_document_id`: `srcdoc_ca767a4c8ff6c3467561c0c5270d60e7`
- `source_revision_id`: `srcrev_cf426d2e82c2432d80e502f12061a5ae`
