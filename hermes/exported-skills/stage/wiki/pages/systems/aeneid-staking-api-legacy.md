---
title: "Aeneid Staking API (Legacy)"
type: "system"
slug: "systems/aeneid-staking-api-legacy"
freshness: "2026-04-30T02:57:45Z"
tags:
  - "aeneid"
  - "api"
  - "incident"
  - "staking"
owners:
  - "Hans"
  - "Jinn"
  - "Yao"
source_revision_ids:
  - "srcrev_14a01a7942017db54a53e7d02a8855e3"
  - "srcrev_2728d23768b8e1821cecd8ff750656f3"
  - "srcrev_2ac99138e4399e14decfc5845d277793"
  - "srcrev_65b6ec7dc13eae3ea33df0d08d5961aa"
  - "srcrev_95fa865d27b30b72cff83d84118cadd2"
  - "srcrev_bdb4c326930f23bd13b3916aba4f5824"
conflict_state: "none"
---

# Aeneid Staking API (Legacy)

## Summary

The legacy staking API service at staking-aeneid.storyapis.com serves the Aeneid devnet staking dashboard. On 2026-04-30, it was found to be misconfigured with mainnet archive node endpoints, causing the dashboard to display a persistent 'Down' banner. The issue was resolved by updating the Kubernetes configuration in story-deployments (PR #225) to point to the correct Aeneid endpoints. Additionally, PR #17 introduced an Indexing status to reduce alarm for future degraded states.

## Claims

- The staking dashboard banner is driven by the staking API's `network_status` endpoint, not a static config. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_315f4833bd5b8ab877d2a2e17d358787` `source_revision_id=srcrev_14a01a7942017db54a53e7d02a8855e3` `chunk_id=srcchunk_5e37f9a880e99ae2cdd010a7215e3286` `native_locator=slack:C0547N89JUB:1777515975.366939:1777516230.422679` `source_timestamp=2026-04-30T02:30:31Z`
- The legacy staking-aeneid.storyapis.com service was configured with mainnet archive node endpoints (IP 54.234.38.229, chain id story-1), which is incorrect for Aeneid devnet. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_315f4833bd5b8ab877d2a2e17d358787` `source_revision_id=srcrev_95fa865d27b30b72cff83d84118cadd2` `chunk_id=srcchunk_4f43cfdb11d126e18b084f45d6ccbc5f` `native_locator=slack:C0547N89JUB:1777515975.366939:1777517483.350249` `source_timestamp=2026-04-30T02:51:27Z`
- Due to the misconfiguration, the `/api/network_status` endpoint returned `status: "Down"`, causing the red banner on the dashboard even though the block heights were live. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_315f4833bd5b8ab877d2a2e17d358787` `source_revision_id=srcrev_65b6ec7dc13eae3ea33df0d08d5961aa` `chunk_id=srcchunk_80f7bb169e810a99c6fa1a7c1a846ef0` `native_locator=slack:C0547N89JUB:1777515975.366939:1777516614.279729` `source_timestamp=2026-04-30T02:36:54Z`
- The correct endpoints for the Aeneid devnet staking API are CL 3.224.230.67:26657 and EL 3.224.230.67:8545, with chain id devnet-1. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_315f4833bd5b8ab877d2a2e17d358787` `source_revision_id=srcrev_95fa865d27b30b72cff83d84118cadd2` `chunk_id=srcchunk_4f43cfdb11d126e18b084f45d6ccbc5f` `native_locator=slack:C0547N89JUB:1777515975.366939:1777517483.350249` `source_timestamp=2026-04-30T02:51:27Z`
- Jinn resolved the misconfiguration by opening PR #225 in the story-deployments repository, updating the old staking-aeneid config to use the correct Aeneid endpoints. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_315f4833bd5b8ab877d2a2e17d358787` `source_revision_id=srcrev_2728d23768b8e1821cecd8ff750656f3` `chunk_id=srcchunk_a14d991d8efb950153cd7e960a807d4c` `native_locator=slack:C0547N89JUB:1777515975.366939:1777517828.530439` `source_timestamp=2026-04-30T02:57:45Z`
- PR #17 in story-staking-dashboard introduces an `Indexing` status with a blue banner, so that future degraded states due to indexing show `Indexing` instead of the alarming `Down`. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_315f4833bd5b8ab877d2a2e17d358787` `source_revision_id=srcrev_2ac99138e4399e14decfc5845d277793` `chunk_id=srcchunk_e171ddcbfff0a9a934fcb7757eaa4723` `native_locator=slack:C0547N89JUB:1777515975.366939:1777516471.060939` `source_timestamp=2026-04-30T02:34:31Z`
  - citation: `source_document_id=srcdoc_315f4833bd5b8ab877d2a2e17d358787` `source_revision_id=srcrev_bdb4c326930f23bd13b3916aba4f5824` `chunk_id=srcchunk_0e50b11e1e56b8cd645746b5081c39a2` `native_locator=slack:C0547N89JUB:1777515975.366939:1777516523.427849` `source_timestamp=2026-04-30T02:35:23Z`

## Open Questions

- Should the legacy staking-aeneid.storyapis.com service be deprecated and replaced by the newer staging-staking-aeneid service?

## Sources

- `source_document_id`: `srcdoc_315f4833bd5b8ab877d2a2e17d358787`
- `source_revision_id`: `srcrev_4c6825d3cdd3e3f292665143971eae73`
