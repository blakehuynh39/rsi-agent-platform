---
title: "SP Tokenomics \u0026 BLAST"
type: "project"
slug: "projects/sp-tokenomics-blast"
freshness: "2026-05-05T06:34:16Z"
tags:
  - "blast"
  - "draft"
  - "research"
  - "tokenomics"
owners: []
source_revision_ids:
  - "srcrev_00baa8c51d727a20c022dc1bb3a2075a"
  - "srcrev_3fe7cb644a60c863a7e3373eb3924aa0"
conflict_state: "none"
---

# SP Tokenomics & BLAST

## Summary

A draft research plan detailing topics related to Blast's tokenomics and their implications for SP.

## Claims

- The SP Tokenomics & BLAST document outlines planned topics including Blast intro, how it works, rebasing tokens vs normal yielding tokens, capital flow, gas fee-sharing mechanism, possible configurations for $IP/$sIP/$rIP tokens, and learnings for SP. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SP-Tokenomics-BLAST-455511f49da54234b84dc99b9bb4c5bb) `source_document_id=srcdoc_4dc9df780b0a0326a5deb9e2a3bcf26a` `source_revision_id=srcrev_3fe7cb644a60c863a7e3373eb3924aa0` `chunk_id=srcchunk_58859cfc93ba346fd0100d0bd7d9d59b` `native_locator=https://www.notion.so/SP-Tokenomics-BLAST-455511f49da54234b84dc99b9bb4c5bb` `source_timestamp=2026-05-05T06:34:13Z`
- Blast's 'How it works' document covers an overview, rebasing tokens vs normal yielding tokens, capital flow end-to-end from user to ETH Mainnet, gas fee-sharing mechanism, and key code snippets. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-1) `source_document_id=srcdoc_f1d50f2878692c1b57cf9ad13f2904f6` `source_revision_id=srcrev_00baa8c51d727a20c022dc1bb3a2075a` `chunk_id=srcchunk_93246c838ad3ad9fd3415c12b99434f8` `native_locator=https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-1` `source_timestamp=2026-05-05T06:34:16Z`
- On Blast, ETH and USDB are natively rebasing tokens. ETH balance for EOAs automatically rebases, and smart contracts can opt-in. USDB automatically rebases for EOAs and smart contracts, but smart contracts can opt-out. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-1) `source_document_id=srcdoc_f1d50f2878692c1b57cf9ad13f2904f6` `source_revision_id=srcrev_00baa8c51d727a20c022dc1bb3a2075a` `chunk_id=srcchunk_93246c838ad3ad9fd3415c12b99434f8` `native_locator=https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-1` `source_timestamp=2026-05-05T06:34:16Z`
- ETH yield on Blast comes from L1 staking (initially Lido) and is transferred to users via rebasing. USDB yield comes from MakerDAO's on-chain T-Bill protocol. The community may replace these yield sources in the future. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-1) `source_document_id=srcdoc_f1d50f2878692c1b57cf9ad13f2904f6` `source_revision_id=srcrev_00baa8c51d727a20c022dc1bb3a2075a` `chunk_id=srcchunk_93246c838ad3ad9fd3415c12b99434f8` `native_locator=https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-1` `source_timestamp=2026-05-05T06:34:16Z`
- Rebasing tokens on Blast have two options: automatic yield for EOAs and claimable yield for contracts. The governor address (defaulting to the contract's own address) is allowed to claim the contract's yield and gas. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-1) `source_document_id=srcdoc_f1d50f2878692c1b57cf9ad13f2904f6` `source_revision_id=srcrev_00baa8c51d727a20c022dc1bb3a2075a` `chunk_id=srcchunk_93246c838ad3ad9fd3415c12b99434f8` `native_locator=https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-1` `source_timestamp=2026-05-05T06:34:16Z`
- Blast gives net gas revenue back to Dapps programmatically, unlike other L2s. Dapp developers can keep this revenue or use it to subsidize gas fees for users. `claim:claim_2_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-2) `source_document_id=srcdoc_f1d50f2878692c1b57cf9ad13f2904f6` `source_revision_id=srcrev_00baa8c51d727a20c022dc1bb3a2075a` `chunk_id=srcchunk_192a3c72327d56b120edaada2c39215f` `native_locator=https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c#chunk-2` `source_timestamp=2026-05-05T06:34:16Z`

## Related Pages

- `projects/blast-intro`
- `projects/blast-learnings`
- `projects/sp-tokenomics-configurations`

## Sources

- `source_document_id`: `srcdoc_f1d50f2878692c1b57cf9ad13f2904f6`
- `source_revision_id`: `srcrev_00baa8c51d727a20c022dc1bb3a2075a`
- `source_url`: [Notion source](https://www.notion.so/How-it-works-d75bdc7e7ed543628fb176022b602c2c)
