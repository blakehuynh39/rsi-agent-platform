---
title: "Test Token Distribution"
type: "runbook"
slug: "runbooks/testnet-token-distribution"
freshness: "2026-04-22T20:27:37Z"
tags:
  - "concurrency-testing"
  - "faucet"
  - "IP-tokens"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_003e908e0eb631079bdacbbd431de9b8"
  - "srcrev_1d96aa25a1755f2026d49e19932d5776"
  - "srcrev_266749e9e757d62630d0bdb897aa9ac2"
  - "srcrev_553e0b5a83e8aae949c7e27b7a237cf5"
  - "srcrev_dc9e85dc4fb87229751ff845177a4c43"
conflict_state: "none"
---

# Test Token Distribution

## Summary

Process for requesting testnet IP tokens from the team, including typical quantities, wallet addresses, and handling of mainnet test tokens.

## Claims

- A team member can provide testnet IP tokens upon request, bypassing faucet restrictions. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_553e0b5a83e8aae949c7e27b7a237cf5` `chunk_id=srcchunk_635689ddb462dc32b7c51bf1c1eb487c` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114459.185299` `source_timestamp=2026-04-13T21:07:39Z`
- Previous test consumed 10k IP quickly, so 20k IP was suggested as sufficient for functional testing. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_266749e9e757d62630d0bdb897aa9ac2` `chunk_id=srcchunk_7b2572e6fbfe2e02b3c7d2789e04f85d` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114604.442899` `source_timestamp=2026-04-13T21:10:04Z`
- Tokens were sent to wallet 0xB6f315F1072781deBE4Af09B24D2CC7f796790de for functional testing. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_dc9e85dc4fb87229751ff845177a4c43` `chunk_id=srcchunk_ec10dee5aad9cc4a6a2d42470e1e4ea3` `native_locator=slack:C04T5307FNU:1776111548.281659:1776114857.294739` `source_timestamp=2026-04-13T21:14:17Z`
- Mainnet IP tokens (5 requested) can be obtained from a mainnet test wallet by contacting a team member directly. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_003e908e0eb631079bdacbbd431de9b8` `chunk_id=srcchunk_5b5346ffe6e68aabebe70134d649d951` `native_locator=slack:C04T5307FNU:1776111548.281659:1776742171.881949` `source_timestamp=2026-04-21T03:29:31Z`
- Subsequent testnet token request for 20k was fulfilled again. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de6d379280b7c02b21f301f10b9e46e7` `source_revision_id=srcrev_1d96aa25a1755f2026d49e19932d5776` `chunk_id=srcchunk_836d0b03a7d59db48f38255c3b54ad31` `native_locator=slack:C04T5307FNU:1776111548.281659:1776889657.536589` `source_timestamp=2026-04-22T20:27:37Z`

## Sources

- `source_document_id`: `srcdoc_de6d379280b7c02b21f301f10b9e46e7`
- `source_revision_id`: `srcrev_ae4e9fd6b076f0326c2e10f1a0180b54`
