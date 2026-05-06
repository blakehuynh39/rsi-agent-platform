---
title: "L1 Devnet Launch Plan"
type: "project"
slug: "projects/l1-devnet-launch"
freshness: "2026-05-05T06:41:13Z"
tags:
  - "blockchain"
  - "devnet"
  - "infrastructure"
  - "ip-graph"
  - "staking"
  - "story-protocol"
owners:
  - "57blocks"
  - "5d2965c9-40b2-44ff-9f1f-ae930770080e"
  - "Allen"
  - "Andrew"
  - "Andy"
  - "Bobby"
  - "Don"
  - "Evan"
  - "fbfcfcc8-1369-415f-9d72-043b5771184c"
  - "Jacob"
  - "Sam"
source_revision_ids:
  - "srcrev_8e90d78d19236aa57fb5ad74a05d6518"
conflict_state: "none"
---

# L1 Devnet Launch Plan

## Summary

Plan to launch an internal L1 Devnet with 25 validators (15 at genesis + 10 later), enshrined Story Protocol v1.1, IP graph support up to 1024 nodes, and full staking functionality. Target initial devnet readiness by June 17, 2024, with subsequent Hub and Staking Dashboard milestones leading to mainnet-style launch in August 2024.

## Claims

- Launch an internal Devnet by June 17 with 15 validator nodes joined from the genesis block and 10 additional validators run by the eng team joined at a later stage. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_79ea9973493a13506523cc958f1569ed` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1` `source_timestamp=2026-05-05T06:41:13Z`
- Every team member, including non‑tech, can claim tokens via a faucet, perform delegation/undelegation, and earn staking rewards using the staking dashboard. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_79ea9973493a13506523cc958f1569ed` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1` `source_timestamp=2026-05-05T06:41:13Z`
- Story protocol smart contracts are deployed to vanity addresses of the blockchain, and everyone can stake to IPA and collect rewards. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_79ea9973493a13506523cc958f1569ed` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1` `source_timestamp=2026-05-05T06:41:13Z`
- The explorer and all protocol functions work correctly, and applications can create a royalty distribution for IP graph with up to 1024 nodes. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_79ea9973493a13506523cc958f1569ed` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1` `source_timestamp=2026-05-05T06:41:13Z`
- Users can interact with the devnet using MetaMask and use the block explorer to verify transactions. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_79ea9973493a13506523cc958f1569ed` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1` `source_timestamp=2026-05-05T06:41:13Z`
- Devnet adopts the Omni network architecture, engine API, and execution/consensus clients separation, and enshrines Story Protocol v1.1 with 1024‑node IP graph precompile support. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_79ea9973493a13506523cc958f1569ed` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1` `source_timestamp=2026-05-05T06:41:13Z`
- Any developer can create validators and join the network, and any user can delegate tokens to validators. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_79ea9973493a13506523cc958f1569ed` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1` `source_timestamp=2026-05-05T06:41:13Z`
- Infrastructure includes faucet, graph node, block scout, and a staking dashboard v0.1 with create validator, staking/unstaking, delegation/undelegation, and basic stats. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_79ea9973493a13506523cc958f1569ed` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-1` `source_timestamp=2026-05-05T06:41:13Z`
- The L1 devnet launch date is August 25, 2024, with dev complete by August 11, 2024. Milestone 1 dev complete date is June 16, 2024, and Milestone 2 is July 14, 2024. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_2e94cfca0ae3ad399e042971d66e6f2f` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2` `source_timestamp=2026-05-05T06:41:13Z`
- Staking dashboard launch date is August 24, 2024, with dev complete by August 11, 2024. Milestone 0 dev complete date is June 16, 2024, Milestone 1 is July 14, 2024, and Milestone 2 is August 24, 2024. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_2e94cfca0ae3ad399e042971d66e6f2f` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2` `source_timestamp=2026-05-05T06:41:13Z`
- Protocol work includes IPA staking contract (p0), graph DB support via precompile (p1), and vector DB support via precompile (p2). `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_2e94cfca0ae3ad399e042971d66e6f2f` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2` `source_timestamp=2026-05-05T06:41:13Z`
- App team (Allen, Don, Sam, Bobby, Jacob) works on Hub and designs including Figma, Storykit, IPAsset provider, home, navbar, hero, featured categories, collections, publishing, and admin dashboard. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_2e94cfca0ae3ad399e042971d66e6f2f` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2` `source_timestamp=2026-05-05T06:41:13Z`
- Hub APIs include GET/POST for featured categories, featured IP collections, IP categories, IP collections, and upload publishing. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2) `source_document_id=srcdoc_68fa048692ccd3c23b2c9c0cdda92727` `source_revision_id=srcrev_8e90d78d19236aa57fb5ad74a05d6518` `chunk_id=srcchunk_2e94cfca0ae3ad399e042971d66e6f2f` `native_locator=https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a#chunk-2` `source_timestamp=2026-05-05T06:41:13Z`

## Open Questions

- Basic Solidity staking contract to demonstrate staking capability in Cosmos SDK
- BLS aggregation on CometBFT – feasibility and estimated work
- Concrete identification of potential Sybil attacks and malicious cases
- Discuss emission curves per IP
- Governance vs. Utility token study (Augur)
- Identify emission fragmentation and impacts on incentivization
- Narrow down to 2 decisions on Proof of Creativity + Staking
- p2p stack of CometBFT – feasibility and estimated work

## Sources

- `source_document_id`: `srcdoc_68fa048692ccd3c23b2c9c0cdda92727`
- `source_revision_id`: `srcrev_8e90d78d19236aa57fb5ad74a05d6518`
- `source_url`: [Notion source](https://www.notion.so/L1-devnet-launch-plan-7113c4ab760643288298f867efcf991a)
