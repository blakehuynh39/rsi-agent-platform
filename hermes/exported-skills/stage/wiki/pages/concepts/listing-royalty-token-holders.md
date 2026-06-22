---
title: "Listing Royalty Token Holders"
type: "concept"
slug: "concepts/listing-royalty-token-holders"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "api"
  - "developers"
  - "indexer"
  - "royalty-tokens"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
  - "srcrev_eab9b46f35fbebc1c58059432bdef991"
  - "srcrev_fcbf4bff5e2cbd33a3e5a2d63eb407f2"
conflict_state: "none"
---

# Listing Royalty Token Holders

## Summary

There is no on-chain function to enumerate all addresses holding Royalty Tokens (RTs) for a given IP Royalty Vault; off-chain solutions like indexers or Blockscout's REST API can provide this data.

## Claims

- There is no on-chain function to list all addresses holding RTs for a given ipId / IpRoyaltyVault. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- This can be supported off-chain through an indexer or a backend API. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Blockscout shows a breakdown of the token holders for each ERC20 token, and royalty tokens are ERC20 so they show up. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
- A screenshot of Blockscout UI demonstrating the token holder breakdown for a specific IP was shared. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_fcbf4bff5e2cbd33a3e5a2d63eb407f2` `chunk_id=srcchunk_da5e6dbfb4d663ec29f6bdf8ee00212c` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818986.576409` `source_timestamp=2026-01-19T10:36:26Z`
- Blockscout exposes REST APIs related to token holders, available at https://www.storyscan.io/api-docs?tab=rest_api. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`
- An intern may be tasked with building this functionality if it is not super urgent. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_eab9b46f35fbebc1c58059432bdef991` `chunk_id=srcchunk_671064d78edcdb466951c233096d87f2` `native_locator=slack:C04T5307FNU:1768763765.811689:1768777273.499599` `source_timestamp=2026-01-18T23:01:13Z`

## Open Questions

- Is the Blockscout API reliable and complete for this purpose?
- What is the best programmatic approach to fetch all holders given an ipId?

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_674079a09632bf4342b5d81b3f6b2b11`
