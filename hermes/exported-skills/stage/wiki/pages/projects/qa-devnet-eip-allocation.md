---
title: "QA Devnet Elastic IP Allocation"
type: "project"
slug: "projects/qa-devnet-eip-allocation"
freshness: "2026-05-18T06:59:49Z"
tags:
  - "aws"
  - "cost"
  - "eip"
  - "qa-devnet"
  - "quota"
  - "story-devnet"
owners:
  - "U07KLPN0JN6"
  - "U07TNT9N4JC"
source_revision_ids:
  - "srcrev_4f421b35123b6417233642c12c3a3a1a"
  - "srcrev_a65cb01a7c1d3d63617e321e52592ce4"
  - "srcrev_c961bafa4fb39f042451bd5c80437842"
conflict_state: "none"
---

# QA Devnet Elastic IP Allocation

## Summary

Allocated 82 Elastic IPs in us-east-1 (account 408766208637) for QA devnet, covering validators, bootnode, and RPC. Quota increased from 12 to 100. Cost ~$298/month. Needed until end of month for mainnet validator reduction.

## Claims

- 82 Elastic IPs were allocated in us-east-1 under the story-devnet account (408766208637) for QA devnet. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_63d8b88a5fe95d6c501eb43f4a17e2f6` `source_revision_id=srcrev_a65cb01a7c1d3d63617e321e52592ce4` `chunk_id=srcchunk_23c3ee64c7291b2634ff50ac3231c36a` `native_locator=slack:C0547N89JUB:1778062059.002519:1778062059.002519` `source_timestamp=2026-05-06T10:07:39Z`
- The EIPs include: 80 for validators (tagged use1-devnet-v170-validator{1-80}-eip), 1 for bootnode (use1-devnet-v170-bootnode1-eip, IP 52.71.126.13), and 1 for RPC (use1-devnet-v170-rpc1-eip, IP 54.225.116.9). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_63d8b88a5fe95d6c501eb43f4a17e2f6` `source_revision_id=srcrev_a65cb01a7c1d3d63617e321e52592ce4` `chunk_id=srcchunk_23c3ee64c7291b2634ff50ac3231c36a` `native_locator=slack:C0547N89JUB:1778062059.002519:1778062059.002519` `source_timestamp=2026-05-06T10:07:39Z`
- The EIP quota was increased from 12 to 100 via Service Quotas (auto-approved). `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_63d8b88a5fe95d6c501eb43f4a17e2f6` `source_revision_id=srcrev_a65cb01a7c1d3d63617e321e52592ce4` `chunk_id=srcchunk_23c3ee64c7291b2634ff50ac3231c36a` `native_locator=slack:C0547N89JUB:1778062059.002519:1778062059.002519` `source_timestamp=2026-05-06T10:07:39Z`
- Expected duration of usage is ~1-2 weeks, and all EIPs will be released after QA completes. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_63d8b88a5fe95d6c501eb43f4a17e2f6` `source_revision_id=srcrev_a65cb01a7c1d3d63617e321e52592ce4` `chunk_id=srcchunk_23c3ee64c7291b2634ff50ac3231c36a` `native_locator=slack:C0547N89JUB:1778062059.002519:1778062059.002519` `source_timestamp=2026-05-06T10:07:39Z`
- The estimated monthly cost for these EIPs is ~$298. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_63d8b88a5fe95d6c501eb43f4a17e2f6` `source_revision_id=srcrev_4f421b35123b6417233642c12c3a3a1a` `chunk_id=srcchunk_b5dae804986c5f4048851037eab71590` `native_locator=slack:C0547N89JUB:1778062059.002519:1779087407.598869` `source_timestamp=2026-05-18T06:56:47Z`
- The EIPs are needed at least until end of month, due to mainnet validator reduction activation. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_63d8b88a5fe95d6c501eb43f4a17e2f6` `source_revision_id=srcrev_c961bafa4fb39f042451bd5c80437842` `chunk_id=srcchunk_48f3220fa4d1d94baa39ef0b7ad75371` `native_locator=slack:C0547N89JUB:1778062059.002519:1779087589.364369` `source_timestamp=2026-05-18T06:59:49Z`

## Sources

- `source_document_id`: `srcdoc_63d8b88a5fe95d6c501eb43f4a17e2f6`
- `source_revision_id`: `srcrev_c961bafa4fb39f042451bd5c80437842`
