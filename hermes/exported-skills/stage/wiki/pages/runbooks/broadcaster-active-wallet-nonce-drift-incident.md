---
title: "Broadcaster Active Wallet Nonce Drift Incident"
type: "runbook"
slug: "runbooks/broadcaster-active-wallet-nonce-drift-incident"
freshness: "2026-06-16T17:18:22Z"
tags:
  - "alert"
  - "nonce-drift"
  - "resolved"
  - "story-api"
owners:
  - "blake.huynh@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_1221c373c3673879948394731b662111"
  - "srcrev_79295e9fbcbade43235150e1442767b0"
  - "srcrev_ae3ae7fde298f47595c7eb522cff1d85"
conflict_state: "none"
---

# Broadcaster Active Wallet Nonce Drift Incident

## Summary

An alert for broadcaster active wallet nonce drift was triggered in story-api on multiple occasions. The issue was later resolved by Blake Huynh via Sentry.

## Claims

- The story-api broadcaster active wallet nonce drift was detected. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e9734c0d99d44bc77933c830aeb83c78` `source_revision_id=srcrev_79295e9fbcbade43235150e1442767b0` `chunk_id=srcchunk_4c6de8c2bf7a7d802c5fa66e963d153b` `native_locator=slack:C07K3J4JTH6:1780810368.929899:1780810368.929899` `source_timestamp=2026-06-07T05:32:48Z`
  - citation: `source_document_id=srcdoc_e9734c0d99d44bc77933c830aeb83c78` `source_revision_id=srcrev_ae3ae7fde298f47595c7eb522cff1d85` `chunk_id=srcchunk_63d1ad8c4d2946a2ee3fcb36216fbfe0` `native_locator=slack:C07K3J4JTH6:1780810368.929899:1780898909.974869` `source_timestamp=2026-06-08T06:08:29Z`
- The Sentry issue STORY-API-EN was marked as resolved by Blake Huynh. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e9734c0d99d44bc77933c830aeb83c78` `source_revision_id=srcrev_1221c373c3673879948394731b662111` `chunk_id=srcchunk_7473a386e6f3bf71f1979e8450814f71` `native_locator=slack:C07K3J4JTH6:1780810368.929899:1781630302.932839` `source_timestamp=2026-06-16T17:18:22Z`

## Sources

- `source_document_id`: `srcdoc_e9734c0d99d44bc77933c830aeb83c78`
- `source_revision_id`: `srcrev_ae3ae7fde298f47595c7eb522cff1d85`
