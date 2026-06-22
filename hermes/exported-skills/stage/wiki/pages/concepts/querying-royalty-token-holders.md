---
title: "Querying Royalty Token Holders"
type: "concept"
slug: "concepts/querying-royalty-token-holders"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "api"
  - "blockscout"
  - "indexer"
  - "ip-royalty-vault"
  - "royalty-tokens"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
  - "srcrev_674079a09632bf4342b5d81b3f6b2b11"
  - "srcrev_eab9b46f35fbebc1c58059432bdef991"
conflict_state: "none"
---

# Querying Royalty Token Holders

## Summary

No on-chain function exists to list all addresses holding RTs for a given IP Royalty Vault. Off-chain retrieval is possible via Blockscout API or a custom indexer/backend API.

## Claims

- There is no on-chain function to list all addresses holding royalty tokens for a given IP Royalty Vault (ipId). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Off-chain solutions such as an indexer or a backend API can support querying token holder data. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Blockscout (Storyscan) shows a breakdown of token holders for ERC20 tokens, including royalty tokens, on the token's page. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
- Blockscout provides a REST API that exposes token holder information, documented at https://www.storyscan.io/api-docs?tab=rest_api. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_674079a09632bf4342b5d81b3f6b2b11` `chunk_id=srcchunk_e54304517c943c929703dd6bbdc5a501` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856258.222539` `source_timestamp=2026-01-19T20:57:38Z`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`
- A potential onboarding task for an intern is planned to implement token holder querying functionality if not urgent. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_eab9b46f35fbebc1c58059432bdef991` `chunk_id=srcchunk_671064d78edcdb466951c233096d87f2` `native_locator=slack:C04T5307FNU:1768763765.811689:1768777273.499599` `source_timestamp=2026-01-18T23:01:13Z`

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_5cc7f5c2c8c352d3527d880d9b5082d1`
