---
title: "Listing IP Royalty Token Holders"
type: "concept"
slug: "concepts/listing-ip-royalty-token-holders"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "api"
  - "blockscout"
  - "holders"
  - "ip"
  - "royalty-tokens"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
  - "srcrev_674079a09632bf4342b5d81b3f6b2b11"
conflict_state: "none"
---

# Listing IP Royalty Token Holders

## Summary

There is no on-chain function to enumerate all royalty token holders for a given IP Asset. Off-chain solutions include using the Blockscout explorer (storyscan.io) API, which provides token holder data for ERC20 tokens.

## Claims

- There is no on-chain function to list all addresses holding royalty tokens (RTs) for a given ipId / IpRoyaltyVault; off-chain support is possible via indexers or backend APIs. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Blockscout (storyscan.io) shows a breakdown of token holders for each ERC20 token, including royalty tokens, providing a manual way to view holders for a specific IP. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
- Blockscout exposes an API for token holder data, documented at storyscan.io/api-docs, enabling programmatic retrieval of royalty token holders. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_674079a09632bf4342b5d81b3f6b2b11` `chunk_id=srcchunk_e54304517c943c929703dd6bbdc5a501` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856258.222539` `source_timestamp=2026-01-19T20:57:38Z`

## Open Questions

- Does Story Protocol have an internal API or service for querying royalty token holders? Mentioned as a possible future development task.

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_fcbf4bff5e2cbd33a3e5a2d63eb407f2`
