---
title: "Testnet IP Token Distribution Runbook"
type: "runbook"
slug: "runbooks/testnet-ip-token-distribution"
freshness: "2026-04-22T20:24:28Z"
tags:
  - "concurrency-testing"
  - "faucet"
  - "ip-token"
  - "testnet"
owners:
  - "U07A7AUGL5V"
source_revision_ids:
  - "srcrev_003e908e0eb631079bdacbbd431de9b8"
  - "srcrev_266749e9e757d62630d0bdb897aa9ac2"
  - "srcrev_2fd9cff66a3c1b9ab1a5932b08d6c9e1"
  - "srcrev_553e0b5a83e8aae949c7e27b7a237cf5"
  - "srcrev_682dd12bf427890578cf791a9d3e74d9"
  - "srcrev_968c6bf789021868ac784ff689afe501"
  - "srcrev_a79588af5be46c95759d3e816b7085b2"
  - "srcrev_dc9e85dc4fb87229751ff845177a4c43"
  - "srcrev_fec9a536c0bff19b89510ee301fe7a5d"
conflict_state: "none"
---

# Testnet IP Token Distribution Runbook

## Summary

Process for requesting and receiving testnet (and mainnet) IP tokens for testing purposes. Tokens are distributed manually by a team member upon request.

## Claims

- Users can request testnet IP tokens by contacting @U07A7AUGL5V and specifying the amount and wallet address. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_2fd9cff66a3c1b9ab1a5932b08d6c9e1` `chunk_id=srcchunk_9b10da2b07536815cdf9b59878ade623` `native_locator=slack:C04T5307FNU:1776111548.281659:1776111548.281659` `source_timestamp=2026-04-13T20:19:08Z`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_553e0b5a83e8aae949c7e27b7a237cf5` `chunk_id=srcchunk_635689ddb462dc32b7c51bf1c1eb487c` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114459.185299` `source_timestamp=2026-04-13T21:07:39Z`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_fec9a536c0bff19b89510ee301fe7a5d` `chunk_id=srcchunk_888b31efa1b5b3787fd5e01218c75d6f` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114482.681139` `source_timestamp=2026-04-13T21:08:02Z`
- It was decided to use 20k IP for functional concurrency testing with a central wallet for Meng. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_266749e9e757d62630d0bdb897aa9ac2` `chunk_id=srcchunk_7b2572e6fbfe2e02b3c7d2789e04f85d` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114604.442899` `source_timestamp=2026-04-13T21:10:04Z`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_682dd12bf427890578cf791a9d3e74d9` `chunk_id=srcchunk_ea4e9340c81b9599189c0ce9cd2d86f7` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114771.971929` `source_timestamp=2026-04-13T21:12:51Z`
- Testnet IP tokens were sent to address 0xB6f315F1072781deBE4Af09B24D2CC7f796790de. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_dc9e85dc4fb87229751ff845177a4c43` `chunk_id=srcchunk_ec10dee5aad9cc4a6a2d42470e1e4ea3` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114857.294739` `source_timestamp=2026-04-13T21:14:17Z`
- Mainnet IP tokens can be requested from the mainnet test wallet by direct messaging the wallet address to the provider. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_a79588af5be46c95759d3e816b7085b2` `chunk_id=srcchunk_28af400cf0b9778b4ed60f4988a298e6` `native_locator=slack:C04T5307FNU:1776111548.281659:1776741640.269099` `source_timestamp=2026-04-21T03:20:40Z`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_003e908e0eb631079bdacbbd431de9b8` `chunk_id=srcchunk_5b5346ffe6e68aabebe70134d649d951` `native_locator=slack:C04T5307FNU:1776111548.281659:1776742171.881949` `source_timestamp=2026-04-21T03:29:31Z`
- Additional testnet tokens can be requested on the same wallet, as seen when a user asked for another 20k testnet tokens. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_968c6bf789021868ac784ff689afe501` `chunk_id=srcchunk_0e5c3da84bf722dbbd79cd697b6d9e51` `native_locator=slack:C04T5307FNU:1776111548.281659:1776889468.305579` `source_timestamp=2026-04-22T20:24:28Z`

## Open Questions

- Is there a limit on how many tokens can be requested? How long does it take to receive tokens?

## Sources

- `source_document_id`: `srcdoc_de6d379280b7c02b21f301f10b9e46e7`
- `source_revision_id`: `srcrev_a79588af5be46c95759d3e816b7085b2`
