---
title: "$DATA Migration Technical Handbook (External)"
type: "decision"
slug: "decisions/data-migration-handbook"
freshness: "2026-06-24T20:51:00Z"
tags:
  - "data"
  - "domain"
  - "ip"
  - "migration"
  - "rebrand"
  - "wdata"
  - "wip"
owners: []
source_revision_ids:
  - "srcrev_6f0bab4577550323b13235037be33536"
conflict_state: "none"
---

# $DATA Migration Technical Handbook (External)

## Summary

Documentation of the Story to Data rebrand: token renaming, domain transitions, WDATA deployment, chain ID stability, guidelines for centralized exchanges, wallets, SDKs, and migration from WIP to WDATA.

## Claims

- Story is rebranding to Data to focus on training data infrastructure. `claim:rebrand_story_to_data` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Story Foundation becomes Data Foundation. `claim:story_foundation_to_data_foundation` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Story L1 chain becomes Data Network. `claim:story_l1_to_data_network` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Native network token $IP renames to $DATA. `claim:native_token_ip_renames_to_data` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- A new ERC-20 wrapper of the native token, $WDATA, is deployed. `claim:wdata_deployed` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- WDATA address is 0xD18a56346227f25D1410F98f78234305660bB877. `claim:wdata_address` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- All story domains transition to data equivalents with 1 month dual support: storyrpc.io → datarpc.io, storyscan.io → datanetscan.io, storyapis.com → dataapis.io, story.foundation → datafdn.org. `claim:domain_transitions` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- New websites replace existing Story sites; staking sites are updated to reflect the new brand. `claim:new_websites_and_staking` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- The GitHub repositories piplabs/story and piplabs/story-geth are archived; development continues in piplabs/data-network and piplabs/data-network-geth. `claim:github_repos_archived` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Chain ID remains 1514; no hard fork is required for token or network name changes. `claim:chain_id_unchanged` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Consensus client App ID 'story-1' and validator Bech32 prefix 'story' remain unchanged; $IP reference in genesis JSON kept to allow syncing from block 0. `claim:app_id_and_bech32_unchanged` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- WDATA is an independently deployed ERC-20 contract; WIP is immutable, so both tokens coexist as functional native token wrappers. `claim:wdata_and_wip_coexist` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Centralized exchanges should update ticker to $DATA, network name to Data Network, RPC endpoints to mainnet.datarpc.io and aeneid.datarpc.io, block explorer to datanetscan.io, and refresh all graphic materials. `claim:cex_actions` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Wallets and frontends should update ticker to $DATA, network name to Data Network, refresh graphics, and point to latest chain list commits. `claim:wallet_and_frontend_updates` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Proof of Creativity protocol and IP Portal require no changes; POC continues to support WIP. `claim:poc_and_ip_portal_no_change` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- SDK integrations require no modifications. `claim:sdk_no_change` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Twitter handles change: @StoryProtocol → @datafdn, @StoryEcosystem → @data_ecosystem, @StoryEngs → @databuilders. `claim:twitter_handle_changes` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- WDATA can be bridged to BSC via a new Layer Zero bridge. `claim:wdata_bridge_to_bsc` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- Balances of native tokens are not affected by the rebranding. `claim:native_balances_unchanged` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_706e1db1d38b6bbc147f3e05618f194d` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-24T20:51:00Z`
- To migrate from WIP to WDATA programmatically: call WIP (0x1514000000000000000000000000000000000000) withdraw, then call WDATA (0xD18a56346227f25D1410F98f78234305660bB877) deposit, or send native tokens directly to the WDATA address. `claim:programmatic_migration` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_f3f285fdcc352304ccd2f21c1f34705a` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-24T20:51:00Z`
- A dedicated WDATA migration page is planned but the URL is not yet confirmed. `claim:wdata_migration_page_tbd` `confidence:0.80`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_6f0bab4577550323b13235037be33536` `chunk_id=srcchunk_f3f285fdcc352304ccd2f21c1f34705a` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-24T20:51:00Z`

## Open Questions

- Are there any pending decisions on the GitHub org for storyprotocol accounts?
- What is the URL for the WDATA migration page?

## Related Pages

- `projects/lion-team`

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_6f0bab4577550323b13235037be33536`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
