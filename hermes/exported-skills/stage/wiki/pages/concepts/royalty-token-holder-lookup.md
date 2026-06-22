---
title: "Royalty Token Holder Lookup"
type: "concept"
slug: "concepts/royalty-token-holder-lookup"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "api"
  - "blockscout"
  - "ip"
  - "royalty-tokens"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
conflict_state: "none"
---

# Royalty Token Holder Lookup

## Summary

How to programmatically obtain all addresses and their royalty token balances for a given IP asset's IpRoyaltyVault. There is no direct on-chain function, but Blockscout block explorer and its API can provide this data off-chain.

## Claims

- There is no on-chain function to list all addresses holding royalty tokens (RTs) for a given ipId or IpRoyaltyVault. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Off-chain support can be provided through an indexer or a backend API. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Blockscout block explorer displays a breakdown of token holders for each ERC20 token, including royalty tokens, for a specific IP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
- Blockscout provides an API that can be used to programmatically access token holder data; documentation is available at https://www.storyscan.io/api-docs?tab=rest_api. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_8930c3cb3d610b96f167c82c2b8e47f0`
