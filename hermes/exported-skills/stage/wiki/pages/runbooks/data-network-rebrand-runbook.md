---
title: "Data Network Rebrand Runbook"
type: "runbook"
slug: "runbooks/data-network-rebrand-runbook"
freshness: "2026-06-05T22:30:00Z"
tags:
  - "data-network"
  - "infrastructure"
  - "migration"
  - "rebrand"
  - "token"
owners: []
source_revision_ids:
  - "srcrev_0b2983847f4ac223e0e2218b198605ca"
conflict_state: "none"
---

# Data Network Rebrand Runbook

## Summary

Technical migration guide for the Story to Data rebrand. Covers token and domain changes, no-hardfork rationale, and required actions for exchanges and integrations.

## Claims

- Story rebrands to Data: Story Foundation → Data Foundation, Story L1 → Data Network, $IP token → $DATA, and a new ERC20 wrapper $WDATA is deployed. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- WDATA contract address is 0xD18a56346227f25D1410F98f78234305660bB877. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- Domains containing 'story' (e.g., storyrpc.io) will transition to 'data' (e.g., datarpc.io) with a 1-month overlap supporting both versions. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- GitHub repositories piplabs/story and piplabs/story-geth will be archived, with continued development in forked repositories piplabs/data-network and piplabs/data-network-geth. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- No hardfork is required because the EVM execution client (story-geth → data-network-geth) does not reference the native token ticker or network name; chain ID remains 1514. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- Consensus client retains 'story' references in app ID 'story-1' and validator bech32 prefix 'story', but a hardfork to change them was deprioritized due to low visibility and complexity. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- The genesis JSON contains a reference to $IP, but it must be preserved to maintain chain history syncability from block 0. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- WIP (Wrapped IP) is an immutable ERC20 contract, so both WIP and WDATA will coexist as native token wrappers. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- Centralized exchanges must update the token ticker to $DATA and adopt new RPC domains; no chain ID change is required. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_4bb2b47c8c16e0e35f82fbc38965a4b2` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-05T22:30:00Z`
- Integrations of Proof of Creativity protocol and IP Portal require no action; POC continues to support WIP, and SDK integrations remain unchanged. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_e609801373f1f79a627df51e2708b1ee` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-05T22:30:00Z`
- Externally, network name must be updated from 'Story' to 'Data Network' in backends and frontends, and graphic materials replaced. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_e609801373f1f79a627df51e2708b1ee` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-05T22:30:00Z`
- Chain list maintainers (e.g., viem, DefiLlama) should receive notice of the rebrand, though chain ID remains 1514. `claim:claim_1_12` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_0b2983847f4ac223e0e2218b198605ca` `chunk_id=srcchunk_e609801373f1f79a627df51e2708b1ee` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-05T22:30:00Z`

## Open Questions

- What are the exact finalized RPC domain names? The document shows mainnet.datanetworkrpc.io with '??', indicating uncertainty.
- Will GitHub repositories under the storyprotocol organization be moved? The document ends with 'Github Repositories under storyprotocol org will move to…?'

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_0b2983847f4ac223e0e2218b198605ca`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
