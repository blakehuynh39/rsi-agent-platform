---
title: "BitGo Staking API Deployment Decision"
type: "decision"
slug: "decisions/bitgo-staking-api-deployment"
freshness: "2026-01-22T06:10:36Z"
tags:
  - "bitgo"
  - "deployment"
  - "infrastructure"
  - "staking-api"
owners: []
source_revision_ids:
  - "srcrev_0bc696b7abf879183248cdc560df651a"
  - "srcrev_1826c4c15a63d4afdd08136282dfb5ef"
  - "srcrev_252555536f7a062a9ac2f67fb5e5a04c"
  - "srcrev_380d03f47ee9951eb37c025f655b4079"
  - "srcrev_8024d3905858587877e3c882a3aa9bf0"
  - "srcrev_a1edce771bd8cf30a9c38d4f94ddc8f6"
  - "srcrev_aade0f2936b6e96443141bf8305550ad"
  - "srcrev_cc30cc2a2191dbaa41f4eaf574241631"
conflict_state: "none"
---

# BitGo Staking API Deployment Decision

## Summary

We decided to deploy a private staking API node for BitGo using a private fork, indexing from genesis on mainnet, and not merge the public PR until integration tests with the staking dashboard are complete.

## Claims

- BitGo's Staking API should be deployed from a private fork rather than the public repository. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_252555536f7a062a9ac2f67fb5e5a04c` `chunk_id=srcchunk_3c9c3ce786b1b78cb8827aa38b91a842` `native_locator=slack:C0547N89JUB:1768970724.228439:1769046178.989119` `source_timestamp=2026-01-22T01:42:58Z`
- The deployment for the private staking API is triggered by merging the staging branch into main on the private repository. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_252555536f7a062a9ac2f67fb5e5a04c` `chunk_id=srcchunk_3c9c3ce786b1b78cb8827aa38b91a842` `native_locator=slack:C0547N89JUB:1768970724.228439:1769046178.989119` `source_timestamp=2026-01-22T01:42:58Z`
- BitGo's IP addresses must be allow-listed for the staking API service. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_a1edce771bd8cf30a9c38d4f94ddc8f6` `chunk_id=srcchunk_ff806668bbf41a4de3dce31aa63919cf` `native_locator=slack:C0547N89JUB:1768970724.228439:1768974669.892979` `source_timestamp=2026-01-21T05:51:09Z`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_380d03f47ee9951eb37c025f655b4079` `chunk_id=srcchunk_e241922bca278c8c1a32ff1abefc9305` `native_locator=slack:C0547N89JUB:1768970724.228439:1768975775.380149` `source_timestamp=2026-01-21T06:09:35Z`
- For mainnet, the indexer service must re-run from the genesis block after code merge. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_aade0f2936b6e96443141bf8305550ad` `chunk_id=srcchunk_0a58b98bda76087c8e0d61d46f3c137a` `native_locator=slack:C0547N89JUB:1768970724.228439:1769045482.487979` `source_timestamp=2026-01-22T01:31:22Z`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_252555536f7a062a9ac2f67fb5e5a04c` `chunk_id=srcchunk_3c9c3ce786b1b78cb8827aa38b91a842` `native_locator=slack:C0547N89JUB:1768970724.228439:1769046178.989119` `source_timestamp=2026-01-22T01:42:58Z`
- A public pull request (piplabs/story-staking-api#60) was opened but must not be merged until additional testing with the staking dashboard is completed. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_cc30cc2a2191dbaa41f4eaf574241631` `chunk_id=srcchunk_a47f4b107440982a9d9838baff596b80` `native_locator=slack:C0547N89JUB:1768970724.228439:1769045534.014079` `source_timestamp=2026-01-22T05:49:28Z`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_1826c4c15a63d4afdd08136282dfb5ef` `chunk_id=srcchunk_8b01885cd955d8993f7f8e305b913b7c` `native_locator=slack:C0547N89JUB:1768970724.228439:1769060972.358269` `source_timestamp=2026-01-22T05:49:32Z`
- The priority is to stand up a new AWS node (without Kubernetes) using the private fork, indexing from the mainnet archive node in the same VPC. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_1826c4c15a63d4afdd08136282dfb5ef` `chunk_id=srcchunk_8b01885cd955d8993f7f8e305b913b7c` `native_locator=slack:C0547N89JUB:1768970724.228439:1769060972.358269` `source_timestamp=2026-01-22T05:49:32Z`
- BitGo currently uses an Aeneid deployment of the staking API, and any redeployment should consider keeping the old node until the new one completes indexing. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_8024d3905858587877e3c882a3aa9bf0` `chunk_id=srcchunk_d545bcf709daa666b009731294bf75f0` `native_locator=slack:C0547N89JUB:1768970724.228439:1769062236.495879` `source_timestamp=2026-01-22T06:10:36Z`
- Indexing from genesis for mainnet is estimated to take about 2-3 days, based on Aeneid experience. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_252555536f7a062a9ac2f67fb5e5a04c` `chunk_id=srcchunk_3c9c3ce786b1b78cb8827aa38b91a842` `native_locator=slack:C0547N89JUB:1768970724.228439:1769046178.989119` `source_timestamp=2026-01-22T01:42:58Z`
- The private fork code contains no mention of BitGo, making it safe to eventually make the repository public. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_17b546bc7dc770ff362c024e87b6dca8` `source_revision_id=srcrev_0bc696b7abf879183248cdc560df651a` `chunk_id=srcchunk_343616bb05fa1db2cb1cf62219c201b9` `native_locator=slack:C0547N89JUB:1768970724.228439:1769047534.050689` `source_timestamp=2026-01-22T02:05:34Z`

## Open Questions

- How will the Aeneid Staking API node be handled during AWS migration?
- What is the exact timeline for indexing the mainnet node from genesis, and how will the switchover from Aeneid to mainnet be coordinated for BitGo?
- When will the staking dashboard integration testing be complete, allowing the public PR to be merged?

## Sources

- `source_document_id`: `srcdoc_17b546bc7dc770ff362c024e87b6dca8`
- `source_revision_id`: `srcrev_cc30cc2a2191dbaa41f4eaf574241631`
