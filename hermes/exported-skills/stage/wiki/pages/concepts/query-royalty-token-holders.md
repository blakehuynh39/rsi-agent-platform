---
title: "Querying Royalty Token Holders for an IP"
type: "concept"
slug: "concepts/query-royalty-token-holders"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "api"
  - "blockscout"
  - "ip"
  - "query"
  - "royalty-tokens"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_5f87700970e9e3846a2f067c3e706b39"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
conflict_state: "none"
---

# Querying Royalty Token Holders for an IP

## Summary

There is no on-chain function to list all addresses holding royalty tokens for a given IP royalty vault. However, off-chain solutions exist using an indexer, backend API, or the Blockscout explorer's API.

## Claims

- There is no on-chain function to list all addresses holding RTs for a given ipId / IpRoyaltyVault. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- This can be supported off-chain through an indexer or a backend API. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Blockscout (Storyscan) shows a breakdown of token holders for each ERC20 token, including royalty tokens. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
- Blockscout provides an API to query token holders; documentation is at https://www.storyscan.io/api-docs?tab=rest_api. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`
- A developer is looking for a programmatic way to retrieve this data, and Blockscout API may suffice. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5f87700970e9e3846a2f067c3e706b39` `chunk_id=srcchunk_93860717565fc1e6c5cbd438ba6275fd` `native_locator=slack:C04T5307FNU:1768763765.811689:1768842449.720939` `source_timestamp=2026-01-19T17:07:29Z`

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_58ce9d9ea6abe0ea1a7d85b3a1d3bde5`
