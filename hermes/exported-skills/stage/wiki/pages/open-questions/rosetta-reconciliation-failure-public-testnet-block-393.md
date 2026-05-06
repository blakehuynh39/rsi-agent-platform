---
title: "Rosetta Reconciliation Failure on Public Testnet at Block 393"
type: "open_question"
slug: "open-questions/rosetta-reconciliation-failure-public-testnet-block-393"
freshness: "2026-05-05T06:41:03Z"
tags:
  - "bug"
  - "public-testnet"
  - "reconciliation"
  - "rosetta"
owners: []
source_revision_ids:
  - "srcrev_bb67047b3180a6c5db77518dfeeda272"
conflict_state: "none"
---

# Rosetta Reconciliation Failure on Public Testnet at Block 393

## Summary

A reconciliation failure occurred on the public-testnet at block height 393 for account 0x5687400189B13551137e330F7ae081142EdfD866. The computed balance differed from the live balance by a small amount, likely due to a fee calculation discrepancy.

## Claims

- A reconciliation failure occurred on public-testnet at block height 393 for account 0x5687400189B13551137e330F7ae081142EdfD866. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e) `source_document_id=srcdoc_01703ba7b5cad2d621808a857d52c02c` `source_revision_id=srcrev_bb67047b3180a6c5db77518dfeeda272` `chunk_id=srcchunk_8dd2b9b6da345a74d525dc1ec5fe5a90` `native_locator=https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e` `source_timestamp=2026-05-05T06:41:03Z`
- The computed balance was 199999000000000000000000000IP, while the live balance was 199998999999653499999853000IP. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e) `source_document_id=srcdoc_01703ba7b5cad2d621808a857d52c02c` `source_revision_id=srcrev_bb67047b3180a6c5db77518dfeeda272` `chunk_id=srcchunk_8dd2b9b6da345a74d525dc1ec5fe5a90` `native_locator=https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e` `source_timestamp=2026-05-05T06:41:03Z`
- The related transaction is 0x161248268bfb80fec2dd2dd0523bbd2fa22d4dcaf1d44f162c1714b33ce8ff14, which transferred 1000000000000000000000 tokens with a fee of 346500000147000. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e) `source_document_id=srcdoc_01703ba7b5cad2d621808a857d52c02c` `source_revision_id=srcrev_bb67047b3180a6c5db77518dfeeda272` `chunk_id=srcchunk_8dd2b9b6da345a74d525dc1ec5fe5a90` `native_locator=https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e` `source_timestamp=2026-05-05T06:41:03Z`
- The balance difference between block 392 and 393 includes the transfer amount plus a fee discrepancy of -7248569912. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e) `source_document_id=srcdoc_01703ba7b5cad2d621808a857d52c02c` `source_revision_id=srcrev_bb67047b3180a6c5db77518dfeeda272` `chunk_id=srcchunk_8dd2b9b6da345a74d525dc1ec5fe5a90` `native_locator=https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e` `source_timestamp=2026-05-05T06:41:03Z`
- The reconciliation error message indicates the failure was triggered by the 'reconciliation fail' action. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e) `source_document_id=srcdoc_01703ba7b5cad2d621808a857d52c02c` `source_revision_id=srcrev_bb67047b3180a6c5db77518dfeeda272` `chunk_id=srcchunk_8dd2b9b6da345a74d525dc1ec5fe5a90` `native_locator=https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e` `source_timestamp=2026-05-05T06:41:03Z`

## Open Questions

- Is this a known issue with the Rosetta implementation or specific to the public-testnet environment?
- What is the root cause of the fee discrepancy (-7248569912) that led to the reconciliation failure?

## Sources

- `source_document_id`: `srcdoc_01703ba7b5cad2d621808a857d52c02c`
- `source_revision_id`: `srcrev_bb67047b3180a6c5db77518dfeeda272`
- `source_url`: [Notion source](https://www.notion.so/Rosetta-testing-issues-111051299a5480fe809bdfefb094e96e)
