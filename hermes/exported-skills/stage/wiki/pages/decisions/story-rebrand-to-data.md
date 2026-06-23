---
title: "$DATA Migration Technical Handbook (External)"
type: "decision"
slug: "decisions/story-rebrand-to-data"
freshness: "2026-06-23T05:35:00Z"
tags:
  - "infrastructure"
  - "rebrand"
  - "tokenomics"
owners: []
source_revision_ids:
  - "srcrev_b697e5f4f71b0774887c70a185e0965c"
conflict_state: "none"
---

# $DATA Migration Technical Handbook (External)

## Summary

Documentation of the Story to Data rebrand: token renaming, domain transitions, WDATA deployment, chain ID stability, guidelines for centralized exchanges, wallets, SDKs, and migration from WIP to WDATA.

## Claims

- Story is rebranding to Data to focus on training data infrastructure. `claim:rebrand_story_to_data` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- Story Foundation becomes Data Foundation. `claim:story_foundation_to_data_foundation` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- The Story L1 chain is renamed to Data Network. `claim:story_l1_to_data_network` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- The native network token $IP is renamed to $DATA. `claim:native_token_ip_to_data` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- A new ERC-20 wrapper of the native token, $WDATA, is deployed at address 0xD18a56346227f25D1410F98f78234305660bB877. `claim:wdata_deployment` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- Domains transition from story to data (e.g., storyrpc.io → datarpc.io, storyscan.io → datanetscan.io, storyapis.com → dataapis.io, story.foundation → datafdn.org), with both versions supported for 1 month. `claim:domain_transitions` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- New websites replace existing Story sites, and staking sites are updated to reflect the new brand. `claim:new_websites_and_staking` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- The GitHub repositories piplabs/story and piplabs/story-geth are archived; development continues in piplabs/data-network and piplabs/data-network-geth. `claim:github_repos_archived` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- Chain ID remains 1514; no hard fork is required for token or network name changes. `claim:chain_id_unchanged` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- The consensus client App ID 'story-1' and validator Bech32 prefix 'story' remain unchanged; the $IP reference in genesis JSON is kept to allow syncing from block 0. `claim:app_id_and_bech32_unchanged` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- WDATA is an independently deployed ERC-20 contract; WIP is immutable, so both tokens coexist as functional native token wrappers. `claim:wdata_and_wip_coexist` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- Centralized exchanges should update ticker to $DATA, network name to Data Network, RPC endpoints to datarpc.io (mainnet) and datarpc.io (aeneid), block explorer to datanetscan.io, and refresh all graphic materials. `claim:cex_actions` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- Proof of Creativity protocol and IP Portal require no changes; POC continues to support WIP. `claim:poc_and_ip_portal_no_change` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- SDK integrations require no modifications. `claim:sdk_no_change` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- Wallets and frontends should update ticker to $DATA, network name to Data Network, refresh graphics, and point to latest chain list commits. `claim:wallet_and_frontend_updates` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- Twitter handles change: @StoryProtocol → @datafdn, @StoryEcosystem → @data_ecosystem, @StoryEngs → @databuilders. `claim:twitter_handle_changes` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_0fcc2be20b139ba29e9642c792f5ae81` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-23T05:35:00Z`
- To migrate from WIP to WDATA programmatically: call WIP (0x1514000000000000000000000000000000000000) withdraw, then call WDATA (0xD18a56346227f25D1410F98f78234305660bB877) deposit, or send native tokens directly to the WDATA address. `claim:programmatic_migration` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_ff78e77b01e47f2f975be7b8f954689b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-23T05:35:00Z`
- A dedicated WDATA migration page is planned but the URL is not yet confirmed. `claim:wdata_migration_page_tbd` `confidence:0.80`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b697e5f4f71b0774887c70a185e0965c` `chunk_id=srcchunk_ff78e77b01e47f2f975be7b8f954689b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-23T05:35:00Z`

## Open Questions

- Handling of storyprotocol GitHub org repositories is not yet specified.
- The exact URL for the WDATA migration page is still to be confirmed.

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_b697e5f4f71b0774887c70a185e0965c`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
