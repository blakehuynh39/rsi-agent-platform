---
title: "Royalty Token Holder Lookup"
type: "concept"
slug: "concepts/royalty-token-holder-lookup"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "blockscout"
  - "erc20"
  - "indexer"
  - "ip-token"
  - "royalty"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
  - "srcrev_674079a09632bf4342b5d81b3f6b2b11"
  - "srcrev_8930c3cb3d610b96f167c82c2b8e47f0"
conflict_state: "none"
---

# Royalty Token Holder Lookup

## Summary

How to retrieve the list of addresses and amounts of Royalty Tokens (RTs) for a given IP Royalty Vault. No on-chain function exists, but off-chain solutions like Blockscout API can provide this data.

## Claims

- There is no on-chain function to list all addresses holding RTs for a given ipId / IpRoyaltyVault. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- An off-chain solution can be implemented via an indexer or backend API. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Blockscout provides a token holder breakdown for ERC20 tokens, which includes royalty tokens. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_8930c3cb3d610b96f167c82c2b8e47f0` `chunk_id=srcchunk_ca299ecc44744c1a586b79a34124e497` `native_locator=slack:C04T5307FNU:1768763765.811689:1768842357.331249` `source_timestamp=2026-01-19T17:05:57Z`
- Blockscout has an API that exposes token holder data, documented at https://www.storyscan.io/api-docs?tab=rest_api. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_674079a09632bf4342b5d81b3f6b2b11` `chunk_id=srcchunk_e54304517c943c929703dd6bbdc5a501` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856258.222539` `source_timestamp=2026-01-19T20:57:38Z`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_c925d5e62745aef09c2a8f782facb531`
