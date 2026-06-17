---
title: "Listing Royalty Token Holders for an IP Royalty Vault"
type: "concept"
slug: "concepts/listing-rt-holders-for-ip-vault"
freshness: "2026-01-19T21:04:08Z"
tags:
  - "api"
  - "blockscout"
  - "off-chain"
  - "on-chain"
  - "royalty"
  - "vault"
owners: []
source_revision_ids:
  - "srcrev_52ddc4727a87382df07d729974783c96"
  - "srcrev_5cc7f5c2c8c352d3527d880d9b5082d1"
  - "srcrev_6024e16f837d45c0c29af7eb7695c126"
  - "srcrev_674079a09632bf4342b5d81b3f6b2b11"
  - "srcrev_c925d5e62745aef09c2a8f782facb531"
  - "srcrev_eab9b46f35fbebc1c58059432bdef991"
conflict_state: "none"
---

# Listing Royalty Token Holders for an IP Royalty Vault

## Summary

Covers methods to list all addresses and their royalty token (RT) balances for a given ipId's royalty vault, including on-chain limitations and off-chain solutions like Blockscout explorer and API.

## Claims

- There is no on-chain function to list all addresses holding RTs for a given ipId / IpRoyaltyVault. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- This information can be supported off-chain through an indexer or a backend API. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_52ddc4727a87382df07d729974783c96` `chunk_id=srcchunk_33776992dfb031ba8546adb9cde9cac1` `native_locator=slack:C04T5307FNU:1768763765.811689:1768764504.314079` `source_timestamp=2026-01-18T19:28:24Z`
- Blockscout explorer visually displays token holder breakdowns for ERC20 tokens, including royalty tokens for a specific IP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_5cc7f5c2c8c352d3527d880d9b5082d1` `chunk_id=srcchunk_fb7ce8a61455fb36b1401d6acabb8cb8` `native_locator=slack:C04T5307FNU:1768763765.811689:1768818771.474869` `source_timestamp=2026-01-19T10:36:57Z`
- Blockscout provides REST APIs to programmatically access token holder data; documentation is available at https://www.storyscan.io/api-docs. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_6024e16f837d45c0c29af7eb7695c126` `chunk_id=srcchunk_068ccbdbd84fb8ee35280bc68864c156` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856547.013329` `source_timestamp=2026-01-19T21:04:08Z`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_674079a09632bf4342b5d81b3f6b2b11` `chunk_id=srcchunk_e54304517c943c929703dd6bbdc5a501` `native_locator=slack:C04T5307FNU:1768763765.811689:1768856258.222539` `source_timestamp=2026-01-19T20:57:38Z`
- An off-chain solution to directly list RT holders may be developed internally, possibly as an onboarding task (not urgent). `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_c925d5e62745aef09c2a8f782facb531` `chunk_id=srcchunk_4d66db7efe046c44885f0d363eeb6b43` `native_locator=slack:C04T5307FNU:1768763765.811689:1768776705.673839` `source_timestamp=2026-01-18T22:51:45Z`
  - citation: `source_document_id=srcdoc_6c1fb8bf95903965f3395767bea5ad57` `source_revision_id=srcrev_eab9b46f35fbebc1c58059432bdef991` `chunk_id=srcchunk_671064d78edcdb466951c233096d87f2` `native_locator=slack:C04T5307FNU:1768763765.811689:1768777273.499599` `source_timestamp=2026-01-18T23:01:13Z`

## Open Questions

- Is there an existing off-chain indexer or service that currently provides RT holder lists for ipIds? What is the timeline for the intern onboarding task?

## Sources

- `source_document_id`: `srcdoc_6c1fb8bf95903965f3395767bea5ad57`
- `source_revision_id`: `srcrev_6024e16f837d45c0c29af7eb7695c126`
