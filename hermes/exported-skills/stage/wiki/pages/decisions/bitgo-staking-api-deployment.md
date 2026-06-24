---
title: "BitGo Staking API Deployment"
type: "decision"
slug: "decisions/bitgo-staking-api-deployment"
freshness: "2026-01-22T06:10:36Z"
tags:
  - "aeneid"
  - "bitgo"
  - "deployment"
  - "mainnet"
  - "staking-api"
owners:
  - "U079ZJ48D62"
  - "U07TNT9N4JC"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_1826c4c15a63d4afdd08136282dfb5ef"
  - "srcrev_252555536f7a062a9ac2f67fb5e5a04c"
  - "srcrev_380d03f47ee9951eb37c025f655b4079"
  - "srcrev_8024d3905858587877e3c882a3aa9bf0"
  - "srcrev_9d6d7ee6ddaee4e5a7d346fb0bbf4dae"
  - "srcrev_a1edce771bd8cf30a9c38d4f94ddc8f6"
  - "srcrev_cc30cc2a2191dbaa41f4eaf574241631"
  - "srcrev_f192d470fa3712101f03d1bac0bd6948"
conflict_state: "none"
---

# BitGo Staking API Deployment

## Summary

Decision on deploying a private fork of the Staking API for BitGo on a new AWS node, indexing from Mainnet genesis, while keeping the public PR open until further testing. The Aeneid node remains operational until the new one completes indexing.

## Claims

- The public Staking API is deployed via GitHub Actions upon merging to the main branch. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_9d6d7ee6ddaee4e5a7d346fb0bbf4dae` `chunk_id=srcchunk_80bff27b5cbee948e08d58674dba816e` `native_locator=slack:C0547N89JUB:1768970724.228439:1768971637.369419` `source_timestamp=2026-01-21T05:00:37Z`
- IP allow-listing is required for partners to access the Staking API. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_a1edce771bd8cf30a9c38d4f94ddc8f6` `chunk_id=srcchunk_ff806668bbf41a4de3dce31aa63919cf` `native_locator=slack:C0547N89JUB:1768970724.228439:1768974669.892979` `source_timestamp=2026-01-21T05:51:09Z`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_380d03f47ee9951eb37c025f655b4079` `chunk_id=srcchunk_e241922bca278c8c1a32ff1abefc9305` `native_locator=slack:C0547N89JUB:1768970724.228439:1768975775.380149` `source_timestamp=2026-01-21T06:09:35Z`
- A private fork of the Staking API exists for BitGo integration, and the plan is to deploy it on a new AWS node (without Kubernetes) in the same VPC as the mainnet archive node. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_1826c4c15a63d4afdd08136282dfb5ef` `chunk_id=srcchunk_8b01885cd955d8993f7f8e305b913b7c` `native_locator=slack:C0547N89JUB:1768970724.228439:1769060972.358269` `source_timestamp=2026-01-22T05:49:32Z`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_252555536f7a062a9ac2f67fb5e5a04c` `chunk_id=srcchunk_3c9c3ce786b1b78cb8827aa38b91a842` `native_locator=slack:C0547N89JUB:1768970724.228439:1769046178.989119` `source_timestamp=2026-01-22T01:42:58Z`
- The new BitGo Staking API node must index data from the mainnet genesis block. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_252555536f7a062a9ac2f67fb5e5a04c` `chunk_id=srcchunk_3c9c3ce786b1b78cb8827aa38b91a842` `native_locator=slack:C0547N89JUB:1768970724.228439:1769046178.989119` `source_timestamp=2026-01-22T01:42:58Z`
- GitHub PR #60 on the public staking-api repository should be left open but not merged until additional testing with the staking dashboard is completed. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_1826c4c15a63d4afdd08136282dfb5ef` `chunk_id=srcchunk_8b01885cd955d8993f7f8e305b913b7c` `native_locator=slack:C0547N89JUB:1768970724.228439:1769060972.358269` `source_timestamp=2026-01-22T05:49:32Z`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_cc30cc2a2191dbaa41f4eaf574241631` `chunk_id=srcchunk_a47f4b107440982a9d9838baff596b80` `native_locator=slack:C0547N89JUB:1768970724.228439:1769045534.014079` `source_timestamp=2026-01-22T05:49:28Z`
- The existing Aeneid Staking API node used by BitGo should not be replaced until the new node completes indexing. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_8024d3905858587877e3c882a3aa9bf0` `chunk_id=srcchunk_d545bcf709daa666b009731294bf75f0` `native_locator=slack:C0547N89JUB:1768970724.228439:1769062236.495879` `source_timestamp=2026-01-22T06:10:36Z`
- BitGo-specific APIs have been tested and are verified to work, but other APIs are still under test by another team member. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_f192d470fa3712101f03d1bac0bd6948` `chunk_id=srcchunk_20c31ff76314a0c892bb067f15ad570d` `native_locator=slack:C0547N89JUB:1768970724.228439:1769049651.379399` `source_timestamp=2026-01-22T02:40:51Z`

## Open Questions

- How will the migration from GCP to AWS affect the Aeneid Staking API service location?
- What is the timeline for completing testing of other APIs?
- Who will set up the new AWS node given that Jinn is fully booked?

## Sources

- `source_document_id`: `srcdoc_17b546bc7dc770ff362c024e87b6dca8`
- `source_revision_id`: `srcrev_380d03f47ee9951eb37c025f655b4079`
