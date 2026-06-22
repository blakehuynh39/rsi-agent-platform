---
title: "Internal Token Distribution for Testing"
type: "policy"
slug: "policies/internal-token-distribution-for-testing"
freshness: "2026-04-21T03:29:31Z"
tags:
  - "distribution"
  - "faucet"
  - "mainnet"
  - "testnet"
  - "token"
owners:
  - "U08951K4SRY"
source_revision_ids:
  - "srcrev_003e908e0eb631079bdacbbd431de9b8"
  - "srcrev_266749e9e757d62630d0bdb897aa9ac2"
  - "srcrev_2fd9cff66a3c1b9ab1a5932b08d6c9e1"
  - "srcrev_553e0b5a83e8aae949c7e27b7a237cf5"
  - "srcrev_682dd12bf427890578cf791a9d3e74d9"
  - "srcrev_a79588af5be46c95759d3e816b7085b2"
  - "srcrev_dc9e85dc4fb87229751ff845177a4c43"
conflict_state: "none"
---

# Internal Token Distribution for Testing

## Summary

When public faucets impose restrictions, RSI team members can request testnet and mainnet IP tokens directly from internal token holders for testing purposes.

## Claims

- Public faucets for testnet IP have restrictions that limit token acquisition. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_2fd9cff66a3c1b9ab1a5932b08d6c9e1` `chunk_id=srcchunk_9b10da2b07536815cdf9b59878ade623` `native_locator=slack:C04T5307FNU:1776111548.281659:1776111548.281659` `source_timestamp=2026-04-13T20:19:08Z`
- U08951K4SRY can provide testnet IP tokens upon request, bypassing faucet restrictions. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_553e0b5a83e8aae949c7e27b7a237cf5` `chunk_id=srcchunk_635689ddb462dc32b7c51bf1c1eb487c` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114459.185299` `source_timestamp=2026-04-13T21:07:39Z`
- A standard amount for functional testing is 20,000 IP, based on previous usage of 10,000 IP being depleted quickly. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_266749e9e757d62630d0bdb897aa9ac2` `chunk_id=srcchunk_7b2572e6fbfe2e02b3c7d2789e04f85d` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114604.442899` `source_timestamp=2026-04-13T21:10:04Z`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_682dd12bf427890578cf791a9d3e74d9` `chunk_id=srcchunk_ea4e9340c81b9599189c0ce9cd2d86f7` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114771.971929` `source_timestamp=2026-04-13T21:12:51Z`
- The central wallet address for receiving testnet IP tokens is 0xB6f315F1072781deBE4Af09B24D2CC7f796790de. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_dc9e85dc4fb87229751ff845177a4c43` `chunk_id=srcchunk_ec10dee5aad9cc4a6a2d42470e1e4ea3` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114857.294739` `source_timestamp=2026-04-13T21:14:17Z`
- Mainnet IP tokens can be requested from internal mainnet test wallets. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_a79588af5be46c95759d3e816b7085b2` `chunk_id=srcchunk_28af400cf0b9778b4ed60f4988a298e6` `native_locator=slack:C04T5307FNU:1776111548.281659:1776741640.269099` `source_timestamp=2026-04-21T03:20:40Z`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_003e908e0eb631079bdacbbd431de9b8` `chunk_id=srcchunk_5b5346ffe6e68aabebe70134d649d951` `native_locator=slack:C04T5307FNU:1776111548.281659:1776742171.881949` `source_timestamp=2026-04-21T03:29:31Z`
- Internal token distribution is used for concurrency testing with a small number of wallets, and may involve providing a central wallet for a team member named Meng. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_682dd12bf427890578cf791a9d3e74d9` `chunk_id=srcchunk_ea4e9340c81b9599189c0ce9cd2d86f7` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114771.971929` `source_timestamp=2026-04-13T21:12:51Z`

## Open Questions

- Who is Meng, and what is the central wallet for Meng?

## Sources

- `source_document_id`: `srcdoc_de6d379280b7c02b21f301f10b9e46e7`
- `source_revision_id`: `srcrev_003e908e0eb631079bdacbbd431de9b8`
