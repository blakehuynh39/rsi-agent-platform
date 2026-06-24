---
title: "Mainnet RPC Report: User-Side Metamask Issue (2026-01-13)"
type: "runbook"
slug: "runbooks/mainnet-rpc-user-issue-20260113"
freshness: "2026-01-27T06:58:08Z"
tags:
  - "incident"
  - "mainnet"
  - "metamask"
  - "rpc"
owners: []
source_revision_ids:
  - "srcrev_04ca2c026aec72e98dc7cbb47e5c6637"
  - "srcrev_6e4518e435a7359c1b68ba1cd91ab34a"
  - "srcrev_e108af2fe4f8cd79a31c55c17dca3e86"
  - "srcrev_f081c2aeb5cac156744f18fd50834410"
conflict_state: "none"
---

# Mainnet RPC Report: User-Side Metamask Issue (2026-01-13)

## Summary

A user reported mainnet RPC not working when sending tokens via Ledger; it was determined to be an outdated MetaMask client, not a server-side problem.

## Claims

- On 2026-01-13, a user reported that mainnet RPC was not working when trying to send tokens, specifically with a Ledger device. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_04ca2c026aec72e98dc7cbb47e5c6637` `chunk_id=srcchunk_56e08fde7d8a11d328443d68369243e6` `native_locator=slack:C0547N89JUB:1769495662.272849:1769495662.272849` `source_timestamp=2026-01-27T06:34:22Z`
- The user mentioned using a Ledger device for the transaction. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_6e4518e435a7359c1b68ba1cd91ab34a` `chunk_id=srcchunk_0af817a71ba9f369056b1ceaa0b96cd7` `native_locator=slack:C0547N89JUB:1769495731.681049:1769495731.681049` `source_timestamp=2026-01-27T06:35:31Z`
- The RPC endpoint mainnet.storyrpc.io was confirmed working by another team member. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_f081c2aeb5cac156744f18fd50834410` `chunk_id=srcchunk_638e2badd14d6db2c128437a19d0e00f` `native_locator=slack:C0547N89JUB:1769495824.424039:1769495824.424039` `source_timestamp=2026-01-27T06:52:18Z`
- The root cause was an outdated MetaMask client; the issue was resolved after updating MetaMask. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_e108af2fe4f8cd79a31c55c17dca3e86` `chunk_id=srcchunk_df6541f54549763b330c8f82868339ea` `native_locator=slack:C0547N89JUB:1769497088.871689:1769497088.871689` `source_timestamp=2026-01-27T06:58:08Z`

## Sources

- `source_document_id`: `srcdoc_b853172082170c2d6c8b40804fa03731`
- `source_revision_id`: `srcrev_27f479a31af8e0fc9b8b72aaa09b2dfd`
