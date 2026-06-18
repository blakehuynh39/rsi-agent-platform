---
title: "Wallet IP Royalties Fetch Error Resolved"
type: "runbook"
slug: "runbooks/wallet-ip-royalties-fetch-error-resolved"
freshness: "2026-02-28T16:00:43Z"
tags:
  - "blake-huynh"
  - "royalties"
  - "story-api"
  - "timeout"
  - "wallet"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_4ebce2743f8007e5a6e439c81c7ab7cd"
  - "srcrev_cb2d9ecdbc3c00082e0d441168246da1"
conflict_state: "none"
---

# Wallet IP Royalties Fetch Error Resolved

## Summary

An error occurred in story-api when fetching wallet IP royalties, resulting in a context deadline exceeded error. The issue was resolved by Blake Huynh.

## Claims

- story-api error: failed to fetch wallet IP royalties due to a context deadline exceeded error from prod-story-orchestration-service for wallet 0x51Ea490184e0A7BB8133771D4C14E7f881d8433E. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8680d88c4832bad89401ee01009b5a32` `source_revision_id=srcrev_4ebce2743f8007e5a6e439c81c7ab7cd` `chunk_id=srcchunk_34c8317c9213fddb3689fe7c69fe6a58` `native_locator=slack:C07K3J4JTH6:1772246143.036249:1772246143.036249` `source_timestamp=2026-02-28T02:35:43Z`
- The issue (STORY-API-EA) was marked as resolved by blake.huynh@storyprotocol.xyz. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8680d88c4832bad89401ee01009b5a32` `source_revision_id=srcrev_cb2d9ecdbc3c00082e0d441168246da1` `chunk_id=srcchunk_d6e386dd24957225c4d0a605af2e754d` `native_locator=slack:C07K3J4JTH6:1772246143.036249:1772294443.030729` `source_timestamp=2026-02-28T16:00:43Z`

## Sources

- `source_document_id`: `srcdoc_8680d88c4832bad89401ee01009b5a32`
- `source_revision_id`: `srcrev_cb2d9ecdbc3c00082e0d441168246da1`
