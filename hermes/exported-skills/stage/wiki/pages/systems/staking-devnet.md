---
title: "Staking Devnet"
type: "system"
slug: "systems/staking-devnet"
freshness: "2026-06-04T06:45:01Z"
tags:
  - "deprecated"
  - "devnet"
  - "infrastructure"
  - "staking"
owners: []
source_revision_ids:
  - "srcrev_178e94f378dcf57901362c0d5cd2ba0d"
  - "srcrev_33e1b2364c5a4ba1eef7afe94fb39e7d"
  - "srcrev_4b002bab8324b8e9da0a0c106db148f6"
  - "srcrev_63ea01675589bbab8c27707ee3a9da6b"
  - "srcrev_c23b4f86e3fcd11b63cc48cbb40a1d00"
conflict_state: "none"
---

# Staking Devnet

## Summary

The staking-devnet repository (story-deployments/staking-devnet/staking-api-devnet) appears to be a deprecated deployment for staking API testing. It is associated with a legacy container image and is being proposed for removal. Current testing uses a different environment (devnet0 with AWS EC2).

## Claims

- The repository is named story-deployments/staking-devnet/staking-api-devnet. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c182e72915b752f62ed83b5182e7c5a2` `source_revision_id=srcrev_33e1b2364c5a4ba1eef7afe94fb39e7d` `chunk_id=srcchunk_127ae50150cb3f7b293e59ad62d639cb` `native_locator=slack:C0547N89JUB:1780554992.192819:1780555190.391379` `source_timestamp=2026-06-04T06:39:50Z`
- The repository is associated with a legacy container image: gcr.io/story-stage-product/staking-api. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c182e72915b752f62ed83b5182e7c5a2` `source_revision_id=srcrev_63ea01675589bbab8c27707ee3a9da6b` `chunk_id=srcchunk_85add12e325756341a5d456170231a90` `native_locator=slack:C0547N89JUB:1780554992.192819:1780555211.975549` `source_timestamp=2026-06-04T06:40:11Z`
- It is speculated that @U0A7JJMU5T2 used this repository for testing staking-api with devnet. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c182e72915b752f62ed83b5182e7c5a2` `source_revision_id=srcrev_178e94f378dcf57901362c0d5cd2ba0d` `chunk_id=srcchunk_d875a95f98eea0c24dadade95badc238` `native_locator=slack:C0547N89JUB:1780554992.192819:1780555271.628579` `source_timestamp=2026-06-04T06:41:11Z`
- Another team member tested staking-api on a devnet0 environment using AWS EC2 deployment, indicating staking-devnet may not be the primary testing environment. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c182e72915b752f62ed83b5182e7c5a2` `source_revision_id=srcrev_c23b4f86e3fcd11b63cc48cbb40a1d00` `chunk_id=srcchunk_69272da927257a6e947a4a13025fb968` `native_locator=slack:C0547N89JUB:1780554992.192819:1780555408.224459` `source_timestamp=2026-06-04T06:43:28Z`
- The repository points to a previous image and is proposed for removal. A PR for domain removal exists: https://github.com/piplabs/cloudflare/pull/212. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c182e72915b752f62ed83b5182e7c5a2` `source_revision_id=srcrev_4b002bab8324b8e9da0a0c106db148f6` `chunk_id=srcchunk_e1525324767cfd3e1d6c190bea1475b4` `native_locator=slack:C0547N89JUB:1780554992.192819:1780555501.041899` `source_timestamp=2026-06-04T06:45:01Z`

## Open Questions

- Is it safe to remove the associated Cloudflare domain per PR #212?
- Is the staking-devnet repository still in use?
- What is the exact purpose and ownership of this repository?

## Sources

- `source_document_id`: `srcdoc_c182e72915b752f62ed83b5182e7c5a2`
- `source_revision_id`: `srcrev_4b002bab8324b8e9da0a0c106db148f6`
