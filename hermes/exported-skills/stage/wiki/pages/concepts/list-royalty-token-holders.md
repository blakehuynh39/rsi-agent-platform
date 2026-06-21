---
title: "Listing Royalty Token Holders for an IP Vault"
type: "concept"
slug: "concepts/list-royalty-token-holders"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "blockscout"
  - "indexer"
  - "ip-vault"
  - "royalty-tokens"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
  - "srcrev_674079a09632bf4342b5d81b3f6b2b11"
  - "srcrev_eab9b46f35fbebc1c58059432bdef991"
  - "srcrev_ed9341b0c4af5baf6a11ee1ee4442076"
conflict_state: "none"
---

# Listing Royalty Token Holders for an IP Vault

## Summary

There is no on-chain function to list all addresses and their royalty token balances for a given IP ID. Off-chain solutions include using an indexer/backend API or blockscout explorers. Blockscout provides a UI breakdown and a REST API for token holders.

## Claims

- There is no on-chain function to list all addresses holding RTs for a given ipId / IpRoyaltyVault. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Off-chain indexer or backend API can provide the list of royalty token holders for a given IP vault. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Blockscout explorer shows a breakdown of the token holders for each ERC20 token, including royalty tokens, for a specific IP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
- Blockscout has a REST API that can be used programmatically to get token holders. The API documentation is at https://www.storyscan.io/api-docs?tab=rest_api. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_ed9341b0c4af5baf6a11ee1ee4442076` `chunk_id=srcchunk_db850c6a1111461b927d29f78b3a30de` `native_locator=slack:C04T5307FNU:1768763765.811689:1768843219.291979` `source_timestamp=2026-01-19T17:20:19Z`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_674079a09632bf4342b5d81b3f6b2b11` `chunk_id=srcchunk_e54304517c943c929703dd6bbdc5a501` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856258.222539` `source_timestamp=2026-01-19T20:57:38Z`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`
- A backend API or indexer for listing royalty token holders may be developed as an intern onboarding task. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_eab9b46f35fbebc1c58059432bdef991` `chunk_id=srcchunk_671064d78edcdb466951c233096d87f2` `native_locator=slack:C04T5307FNU:1768763765.811689:1768777273.499599` `source_timestamp=2026-01-18T23:01:13Z`

## Open Questions

- Will an internal indexer/API be built for programmatic access to royalty token holders?

## Related Pages

- `blockscout-explorer`
- `ip-royalty-vault`
- `royalty-tokens`

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_52ddc4727a87382df07d729974783c96`
