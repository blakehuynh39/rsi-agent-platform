---
title: "Story to Data Migration"
type: "project"
slug: "projects/story-to-data-migration"
freshness: "2026-06-05T18:30:00Z"
tags:
  - "migration"
  - "rebranding"
  - "token"
owners: []
source_revision_ids:
  - "srcrev_6ad59997abd2501f2acb36de01762132"
conflict_state: "none"
---

# Story to Data Migration

## Summary

The migration plan for rebranding Story to Data, including token changes, domain updates, and technical actions for exchanges and integrations.

## Claims

- Story is rebranding to Data. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- Story Foundation becomes Data Foundation. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- The L1 chain is renamed from Story to Data Network. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- The native token $IP is renamed to $DATA. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- A new ERC20 wrapper token, $WDATA, will be deployed. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- All domains containing 'story' will transition to 'data', with both versions supported for one month. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- The GitHub repositories piplabs/story and piplabs/story-geth will be archived, replaced by piplabs/data-network and piplabs/data-network-geth. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- No hardfork is required because the EVM execution client does not reference the native token ticker or network name; chain ID remains 1514. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- The Cosmos-based consensus client retains minor references to 'story' in the App ID ('story-1') and validator Bech32 prefix, but these are invisible to users and not changed due to complexity. `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- WDATA is an independently deployed ERC20 contract and will coexist with the existing WIP token. `claim:claim_1_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`
- Centralized exchanges need to update RPC endpoints from mainnet.storyrpc.io to mainnet.datanetworkrpc.io. `claim:claim_1_11` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6ad59997abd2501f2acb36de01762132` `chunk_id=srcchunk_0009d55dd23f35a93869e5e5cf69ce88` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:30:00Z`

## Open Questions

- Where will the GitHub repositories under the `storyprotocol` organization be moved?

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_6ad59997abd2501f2acb36de01762132`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
