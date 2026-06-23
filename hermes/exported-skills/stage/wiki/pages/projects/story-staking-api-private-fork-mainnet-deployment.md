---
title: "Story Staking API Private Fork Mainnet Deployment"
type: "project"
slug: "projects/story-staking-api-private-fork-mainnet-deployment"
freshness: "2026-01-13T03:50:03Z"
tags:
  - "deployment"
  - "mainnet"
  - "staking-api"
owners: []
source_revision_ids:
  - "srcrev_022f245d3ac6a9d0f6821f9cfab415c9"
  - "srcrev_192bc0eee26fd3bd0115554491de6173"
  - "srcrev_54b11acf7c453892731fab08e11effcd"
  - "srcrev_74175231c7d529f13df1a336cf96e949"
  - "srcrev_75f4e7452b8cc505a2c75edb6597ab39"
  - "srcrev_7b6538c068dff70aeca554583733993d"
  - "srcrev_83ba3bfca951a53734687b25e9fc4502"
  - "srcrev_bc31a8aa6d392bab8fcb4ffe06cec1ac"
  - "srcrev_cd7b4b0314d5d3a9a56772213cff6fd8"
conflict_state: "none"
---

# Story Staking API Private Fork Mainnet Deployment

## Summary

Deployment of the story-staking-api-private-fork to mainnet, initially requested on 2026-01-13, to be performed with newly synced archive node. After a brief postponement, the deployment was accelerated early due to urgency from the BitGo team, who needed it for investor-facing work.

## Claims

- A deployment of the story-staking-api-private-fork to mainnet was requested on 2026-01-13. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_54b11acf7c453892731fab08e11effcd` `chunk_id=srcchunk_96d5636f18147a5bc249211f830b23cb` `native_locator=slack:C0547N89JUB:1768273154.814879:1768273154.814879` `source_timestamp=2026-01-13T02:59:14Z`
- The deployment must use the newly synced archive node. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_54b11acf7c453892731fab08e11effcd` `chunk_id=srcchunk_96d5636f18147a5bc249211f830b23cb` `native_locator=slack:C0547N89JUB:1768273154.814879:1768273154.814879` `source_timestamp=2026-01-13T02:59:14Z`
- All archive nodes have been updated, so the existing endpoint can be used as is. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_54b11acf7c453892731fab08e11effcd` `chunk_id=srcchunk_96d5636f18147a5bc249211f830b23cb` `native_locator=slack:C0547N89JUB:1768273154.814879:1768273154.814879` `source_timestamp=2026-01-13T02:59:14Z`
- The deployment was initially proposed to be postponed to the following week due to bandwidth constraints. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_83ba3bfca951a53734687b25e9fc4502` `chunk_id=srcchunk_33aca4b72d747fc9bca1959baa124e84` `native_locator=slack:C0547N89JUB:1768273154.814879:1768273423.485199` `source_timestamp=2026-01-13T03:03:43Z`
- The requestor agreed that postponing to the next week was acceptable. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_192bc0eee26fd3bd0115554491de6173` `chunk_id=srcchunk_6a43ab2fcf36cc0de6ccb30935aae9b8` `native_locator=slack:C0547N89JUB:1768273154.814879:1768273494.400749` `source_timestamp=2026-01-13T03:04:54Z`
- Later, a suggestion was made to start the deployment earlier to allow indexing to complete, as it might be needed by the BitGo team. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_7b6538c068dff70aeca554583733993d` `chunk_id=srcchunk_78b9ebdaad498369e0c017c758a857c8` `native_locator=slack:C0547N89JUB:1768273154.814879:1768273608.353609` `source_timestamp=2026-01-13T03:06:48Z`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_74175231c7d529f13df1a336cf96e949` `chunk_id=srcchunk_dfe34f05023a56fea58846d9da662db5` `native_locator=slack:C0547N89JUB:1768273154.814879:1768273850.802989` `source_timestamp=2026-01-13T03:10:50Z`
- Investors are pressuring the BitGo team. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_022f245d3ac6a9d0f6821f9cfab415c9` `chunk_id=srcchunk_cbd6939fffa3ccfa73a04ee06b564684` `native_locator=slack:C0547N89JUB:1768273154.814879:1768273913.218969` `source_timestamp=2026-01-13T03:11:53Z`
- Deploying the mainnet environment early is beneficial to allow the BitGo team to prepare after testing on Aeneid. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_75f4e7452b8cc505a2c75edb6597ab39` `chunk_id=srcchunk_a3e46cbe805f91b55e0eaaf661621f3c` `native_locator=slack:C0547N89JUB:1768273154.814879:1768274295.200799` `source_timestamp=2026-01-13T03:18:15Z`
- A team member (U09M2SPUTSL) agreed to liaise with the BitGo team on the deployment. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_cd7b4b0314d5d3a9a56772213cff6fd8` `chunk_id=srcchunk_0b9042825e5e25120910cc7ee077650a` `native_locator=slack:C0547N89JUB:1768273154.814879:1768276177.861089` `source_timestamp=2026-01-13T03:49:37Z`
- Appreciation was expressed for U09M2SPUTSL's liaison role. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4a7ce59a30ad54516f0ac7d868f659ed` `source_revision_id=srcrev_bc31a8aa6d392bab8fcb4ffe06cec1ac` `chunk_id=srcchunk_d7b4e0d5fdb8e55e80f9dcd8d992f89e` `native_locator=slack:C0547N89JUB:1768273154.814879:1768276203.695899` `source_timestamp=2026-01-13T03:50:03Z`

## Sources

- `source_document_id`: `srcdoc_4a7ce59a30ad54516f0ac7d868f659ed`
- `source_revision_id`: `srcrev_192bc0eee26fd3bd0115554491de6173`
