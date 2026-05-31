---
title: "4844 Blob Transaction Setup Guide"
type: "runbook"
slug: "runbooks/4844-blob-transaction-setup-guide"
freshness: "2024-10-09T16:17:00Z"
tags:
  - "4844"
  - "blob-transactions"
  - "setup"
  - "typescript"
  - "viem"
owners: []
source_revision_ids:
  - "srcrev_d2ceaf8d4b4de113b684a58308017972"
conflict_state: "none"
---

# 4844 Blob Transaction Setup Guide

## Summary

Step-by-step guide to set up a project for sending 4844 blob transactions using viem and TypeScript.

## Claims

- The guide shows how to replicate sending 4844 blob transactions, with a future tutorial planned for using `cast`. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417) `source_document_id=srcdoc_1bd982ac8f7e67b03d6544bd4928a7b3` `source_revision_id=srcrev_d2ceaf8d4b4de113b684a58308017972` `chunk_id=srcchunk_e4cf32ec5284541f2aa3b2a78eab8cb0` `native_locator=https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417` `source_timestamp=2024-10-09T16:17:00Z`
- Step 1: Create a new project directory and initialize it with a `package.json` file using `mkdir -p 4844-testing/src && cd 4844-testing && npm init -y`. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417) `source_document_id=srcdoc_1bd982ac8f7e67b03d6544bd4928a7b3` `source_revision_id=srcrev_d2ceaf8d4b4de113b684a58308017972` `chunk_id=srcchunk_e4cf32ec5284541f2aa3b2a78eab8cb0` `native_locator=https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417` `source_timestamp=2024-10-09T16:17:00Z`
- Step 2: Add specific dependencies to `package.json` including `@types/node`, `c-kzg`, `dotenv`, `ts-node`, `typescript`, and `viem`. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417) `source_document_id=srcdoc_1bd982ac8f7e67b03d6544bd4928a7b3` `source_revision_id=srcrev_d2ceaf8d4b4de113b684a58308017972` `chunk_id=srcchunk_e4cf32ec5284541f2aa3b2a78eab8cb0` `native_locator=https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417` `source_timestamp=2024-10-09T16:17:00Z`
- Step 3: Install the dependencies with `npm i`. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417) `source_document_id=srcdoc_1bd982ac8f7e67b03d6544bd4928a7b3` `source_revision_id=srcrev_d2ceaf8d4b4de113b684a58308017972` `chunk_id=srcchunk_e4cf32ec5284541f2aa3b2a78eab8cb0` `native_locator=https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417` `source_timestamp=2024-10-09T16:17:00Z`
- Step 4: Create a client in `src/client.ts` that loads environment variables, defines a custom chain (e.g., iliad with id 1511), and creates wallet and public clients using viem. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417) `source_document_id=srcdoc_1bd982ac8f7e67b03d6544bd4928a7b3` `source_revision_id=srcrev_d2ceaf8d4b4de113b684a58308017972` `chunk_id=srcchunk_e4cf32ec5284541f2aa3b2a78eab8cb0` `native_locator=https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417` `source_timestamp=2024-10-09T16:17:00Z`
- The chain `id` must be set to the chain ID being tested, and if running locally, `RPC_LOCAL_URL` must be set to the geth RPC URL. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417) `source_document_id=srcdoc_1bd982ac8f7e67b03d6544bd4928a7b3` `source_revision_id=srcrev_d2ceaf8d4b4de113b684a58308017972` `chunk_id=srcchunk_e4cf32ec5284541f2aa3b2a78eab8cb0` `native_locator=https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417` `source_timestamp=2024-10-09T16:17:00Z`
- Step 5: Add a KZG interface for computing data availability proofs of blobs, using `c-kzg` and `viem`'s `setupKzg`, and loading the trusted setup from `node_modules/viem/trusted-setups/mainnet.json`. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417) `source_document_id=srcdoc_1bd982ac8f7e67b03d6544bd4928a7b3` `source_revision_id=srcrev_d2ceaf8d4b4de113b684a58308017972` `chunk_id=srcchunk_e4cf32ec5284541f2aa3b2a78eab8cb0` `native_locator=https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417` `source_timestamp=2024-10-09T16:17:00Z`

## Sources

- `source_document_id`: `srcdoc_1bd982ac8f7e67b03d6544bd4928a7b3`
- `source_revision_id`: `srcrev_d2ceaf8d4b4de113b684a58308017972`
- `source_url`: [Notion source](https://www.notion.so/Creating-4844-Transactions-11a051299a5480a0bd8bd3caeb9b7417)
