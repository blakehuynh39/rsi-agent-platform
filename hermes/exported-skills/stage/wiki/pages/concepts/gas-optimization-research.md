---
title: "Gas Optimization Research"
type: "concept"
slug: "concepts/gas-optimization-research"
freshness: "2023-03-31T16:19:00Z"
tags:
  - "gas-optimization"
  - "research"
  - "solidity"
owners: []
source_revision_ids:
  - "srcrev_ac803ca7d8c644a5115369da0f98c7ad"
conflict_state: "none"
---

# Gas Optimization Research

## Summary

A collection of resources and notes on gas optimization techniques for Solidity smart contracts.

## Claims

- The document references a collection of gas optimization tricks from the OpenZeppelin forum. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3) `source_document_id=srcdoc_59859be4c4ca50609e48898a30bedc5e` `source_revision_id=srcrev_ac803ca7d8c644a5115369da0f98c7ad` `chunk_id=srcchunk_382ca49193da40c667330755d0c40542` `native_locator=https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3` `source_timestamp=2023-03-31T16:19:00Z`
- Custom errors are listed as a gas optimization technique. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3) `source_document_id=srcdoc_59859be4c4ca50609e48898a30bedc5e` `source_revision_id=srcrev_ac803ca7d8c644a5115369da0f98c7ad` `chunk_id=srcchunk_382ca49193da40c667330755d0c40542` `native_locator=https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3` `source_timestamp=2023-03-31T16:19:00Z`
- Compilation with IR is noted as a gas optimization technique, but it had security issues before Solidity 0.8.17 and needs further research. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3) `source_document_id=srcdoc_59859be4c4ca50609e48898a30bedc5e` `source_revision_id=srcrev_ac803ca7d8c644a5115369da0f98c7ad` `chunk_id=srcchunk_382ca49193da40c667330755d0c40542` `native_locator=https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3` `source_timestamp=2023-03-31T16:19:00Z`
- The Solmate library is mentioned as a gas optimization library but is considered risky. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3) `source_document_id=srcdoc_59859be4c4ca50609e48898a30bedc5e` `source_revision_id=srcrev_ac803ca7d8c644a5115369da0f98c7ad` `chunk_id=srcchunk_382ca49193da40c667330755d0c40542` `native_locator=https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3` `source_timestamp=2023-03-31T16:19:00Z`
- A comparison between ERC721a and current OpenZeppelin implementations is suggested for batch minting gas optimization. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3) `source_document_id=srcdoc_59859be4c4ca50609e48898a30bedc5e` `source_revision_id=srcrev_ac803ca7d8c644a5115369da0f98c7ad` `chunk_id=srcchunk_382ca49193da40c667330755d0c40542` `native_locator=https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3` `source_timestamp=2023-03-31T16:19:00Z`

## Open Questions

- Is compilation with IR safe for production use after Solidity 0.8.17?
- What are the specific risks of using Solmate for gas optimization?

## Sources

- `source_document_id`: `srcdoc_59859be4c4ca50609e48898a30bedc5e`
- `source_revision_id`: `srcrev_ac803ca7d8c644a5115369da0f98c7ad`
- `source_url`: [Notion source](https://www.notion.so/Gas-Optimization-research-cf39cdfc6680424a95de6b764ad450d3)
