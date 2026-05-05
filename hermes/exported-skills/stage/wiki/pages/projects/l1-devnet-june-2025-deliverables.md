---
title: "L1 Devnet June 2025 Deliverables"
type: "project"
slug: "projects/l1-devnet-june-2025-deliverables"
freshness: "2026-05-05T06:35:49Z"
tags:
  - "devnet"
  - "evm"
  - "l1"
  - "roadmap"
  - "staking"
owners:
  - "Andy"
  - "Haodi"
  - "Lutty"
  - "Zerui"
source_revision_ids:
  - "srcrev_000d7f6f2d02821741bd961801575299"
conflict_state: "none"
---

# L1 Devnet June 2025 Deliverables

## Summary

Plan for the L1 devnet deliverables targeting June 17, 2025, covering EVM end-to-end flow, monitoring, testing, staking integration, and documentation.

## Claims

- Week of May 27 L1 tasks include enabling EVM end-to-end flow: sending transactions through wallets (Zerui), viewing on explorer, setting up BlockScout (Andy), deploying smart contract tests, JSON RPC tests (Haodi & Lutty), CI/CD workflow (Andy), local testing using e2e docker local deployment, PR workflow with unit tests and docker container tests, disabling unnecessary omni PR workflows, deployment flow, binary version control/build flow, and setting up a 4+1 devnet for devs (Zerui) leveraging Cosmos SDK and Halo to auto-generate keys and config files. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c) `source_document_id=srcdoc_c61b89ce5bd3ecf734efa26e19b17d76` `source_revision_id=srcrev_000d7f6f2d02821741bd961801575299` `chunk_id=srcchunk_3cc4da01b2103d0a555657fe242e8981` `native_locator=https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c` `source_timestamp=2026-05-05T06:35:49Z`
- Week of May 27 also includes consensus/performance testing (Haodi & Lutty) with a test plan covering health tests via Comet, Cosmos, and Geth APIs, and performance tests for benchmark, TPS, and block time. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c) `source_document_id=srcdoc_c61b89ce5bd3ecf734efa26e19b17d76` `source_revision_id=srcrev_000d7f6f2d02821741bd961801575299` `chunk_id=srcchunk_3cc4da01b2103d0a555657fe242e8981` `native_locator=https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c` `source_timestamp=2026-05-05T06:35:49Z`
- Week of June 3 tasks include monitoring with Prometheus for both Geth and Comet, listing public APIs in Postman, HTTPS/domain name setup, performance testing, end-to-end user testing, integration of staking and predeploy/precompile, and documentation for public API, public RPC URL, and faucet. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c) `source_document_id=srcdoc_c61b89ce5bd3ecf734efa26e19b17d76` `source_revision_id=srcrev_000d7f6f2d02821741bd961801575299` `chunk_id=srcchunk_3cc4da01b2103d0a555657fe242e8981` `native_locator=https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c` `source_timestamp=2026-05-05T06:35:49Z`
- The internal devnet spec requires a working devnet with 15 validator nodes joined from genesis, 10 validators run by the eng team joining later, a faucet for all team members to claim tokens and perform delegation/undelegation, and a staking dashboard to earn rewards. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c) `source_document_id=srcdoc_c61b89ce5bd3ecf734efa26e19b17d76` `source_revision_id=srcrev_000d7f6f2d02821741bd961801575299` `chunk_id=srcchunk_3cc4da01b2103d0a555657fe242e8981` `native_locator=https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c` `source_timestamp=2026-05-05T06:35:49Z`
- Story protocol smart contracts are deployed to vanity addresses; everyone can stake to IPA and collect rewards; the explorer and all protocol functions work; applications can create a royalty distribution for an IP graph with up to 1024 nodes; users can interact with the devnet using MetaMask. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c) `source_document_id=srcdoc_c61b89ce5bd3ecf734efa26e19b17d76` `source_revision_id=srcrev_000d7f6f2d02821741bd961801575299` `chunk_id=srcchunk_3cc4da01b2103d0a555657fe242e8981` `native_locator=https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c` `source_timestamp=2026-05-05T06:35:49Z`
- Decomposition of requirements includes infra (15 genesis validators, 10 post-genesis nodes, block explorer, testing), staking/delegation (smart contracts for stake/unstake, reward distribution/collection), app (staking dashboard with RPC URL, validator query endpoint, debugging discussion, faucet), precompile/predeploy (IPA staking and rewards, IP graph of 1024 nodes for royalty distribution). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c) `source_document_id=srcdoc_c61b89ce5bd3ecf734efa26e19b17d76` `source_revision_id=srcrev_000d7f6f2d02821741bd961801575299` `chunk_id=srcchunk_3cc4da01b2103d0a555657fe242e8981` `native_locator=https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c` `source_timestamp=2026-05-05T06:35:49Z`

## Related Pages

- `june-17th-devnet-deliverables-platform`
- `june-devnet-retro`

## Sources

- `source_document_id`: `srcdoc_c61b89ce5bd3ecf734efa26e19b17d76`
- `source_revision_id`: `srcrev_000d7f6f2d02821741bd961801575299`
- `source_url`: [Notion source](https://www.notion.so/June-17th-L1-tasks-4f4c21eef35a409795ac79921d5b269c)
