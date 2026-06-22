---
title: "Querying Royalty Token Holders for IP Royalty Vault"
type: "concept"
slug: "concepts/royalty-token-holders-query"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "blockscout"
  - "ip-royalty-vault"
  - "off-chain"
  - "query"
  - "royalty-tokens"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
  - "srcrev_eab9b46f35fbebc1c58059432bdef991"
conflict_state: "none"
---

# Querying Royalty Token Holders for IP Royalty Vault

## Summary

Discussion on how to retrieve all addresses and their royalty token (RT) balances for a given IP Royalty Vault. No on-chain function exists; off-chain indexing or external APIs (like Blockscout) can provide this data.

## Claims

- There is no on-chain function to list all addresses holding royalty tokens for a given IP Royalty Vault (ipId). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- The feature can be supported off-chain through an indexer or backend API. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Building an off-chain indexing solution for royalty token holders could be assigned as an onboarding task for an intern (non-urgent). `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_eab9b46f35fbebc1c58059432bdef991` `chunk_id=srcchunk_671064d78edcdb466951c233096d87f2` `native_locator=slack:C04T5307FNU:1768777273.499599` `source_timestamp=2026-01-18T23:01:13Z`
- Blockscout (storyscan.io) displays a token holder breakdown for ERC20 tokens, including royalty tokens, for a specific IP. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
- Blockscout provides a REST API for token-related data; documentation available at https://www.storyscan.io/api-docs?tab=rest_api. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`

## Open Questions

- Is there a dedicated off-chain service (beyond Blockscout) that provides programmatic access to royalty token holders for a given ipId?

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_ed9341b0c4af5baf6a11ee1ee4442076`
