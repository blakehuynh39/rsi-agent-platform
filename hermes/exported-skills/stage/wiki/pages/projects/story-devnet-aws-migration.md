---
title: "Devnet Migration from GCP to AWS"
type: "project"
slug: "projects/story-devnet-aws-migration"
freshness: "2026-02-02T09:57:44Z"
tags:
  - "aws"
  - "devnet"
  - "gcp"
  - "migration"
  - "story-devnet-aws"
owners:
  - "Lucas"
  - "Yao Wang"
source_revision_ids:
  - "srcrev_0ec51856e296cdc24889fa34ee9ae9dc"
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_1a9a34e7011dad60e06a26f50442515b"
  - "srcrev_2f4c634ece1214f473aae27110313f7d"
  - "srcrev_927830e489716b6394b2f77580855ffe"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
  - "srcrev_b0d8c6211859790272dd6233da18c19b"
  - "srcrev_ebb79a70a9425edc788a67443be7b77a"
  - "srcrev_f3d249a79629e8a22cafdef0bc303e75"
conflict_state: "none"
---

# Devnet Migration from GCP to AWS

## Summary

Migrating the story devnet from GCP to AWS using the new story-devnet-aws repository, with planned RPC and explorer domains.

## Claims

- Yao Wang created the repository story-devnet-aws (https://github.com/storyprotocol/story-devnet-aws) to support the devnet migration from GCP to AWS, based on previous internal-devnet provisioning code. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Planned devnet domains: RPC at devnet0.storyrpc.io, Explorer at devnet0.storyscan.io. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The AWS Server Access Guide is available at https://www.notion.so/storyprotocol/AWS-Server-Access-Guide-2ee051299a5480d998e2d3bb23abd3ea. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The repository contains a genesis generation helper script `update-genesis-hash.sh` which does not require special context to use. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_1a9a34e7011dad60e06a26f50442515b` `chunk_id=srcchunk_ccd938e272141324d371faa5e0eac0a6` `native_locator=slack:C0547N89JUB:1768945631.194489:1768974983.803429` `source_timestamp=2026-01-21T05:56:23Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
- Lucas requested and was granted maintainer role on the story-devnet-aws repository by Yao. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_b0d8c6211859790272dd6233da18c19b` `chunk_id=srcchunk_33adf021f969f72c414e1b4f6c1dea99` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416442.559699` `source_timestamp=2026-01-26T08:34:10Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_927830e489716b6394b2f77580855ffe` `chunk_id=srcchunk_bf3cd60ca7c966b78445192dc380d362` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416521.546069` `source_timestamp=2026-01-26T08:35:21Z`
- Yao requires ability to launch EC2 instances for full sync testing. However, he encountered AccessDeniedException for SSM StartSession with both SystemAdministrator and AWS-QA-Devnet-Access roles, indicating missing ssm:StartSession permission in the identity-based policies. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_f3d249a79629e8a22cafdef0bc303e75` `chunk_id=srcchunk_91937e509941fbad6c184e3d30759e15` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416327.392909` `source_timestamp=2026-01-26T08:36:40Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_ebb79a70a9425edc788a67443be7b77a` `chunk_id=srcchunk_07b687cd06bd70fb344a36bc9e56057f` `native_locator=slack:C0547N89JUB:1768945631.194489:1770026264.209509` `source_timestamp=2026-02-02T09:57:44Z`
- The permission policy for Yao was defined in terraform at https://github.com/storyprotocol/AWS-Organization/blob/0bab852b32a88f662402d060589d755bd0063e01/accounts/story-dev-ou/main.tf#L33 and was claimed to be resolved, but Yao still experienced issues after that. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_2f4c634ece1214f473aae27110313f7d` `chunk_id=srcchunk_fec44376ee4a1fccaa26319e9b3c6214` `native_locator=slack:C0547N89JUB:1768945631.194489:1769706890.638689` `source_timestamp=2026-01-29T17:14:50Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_0ec51856e296cdc24889fa34ee9ae9dc` `chunk_id=srcchunk_81c176a5a7bbde1142e8a3c13b09f6bb` `native_locator=slack:C0547N89JUB:1768945631.194489:1770023007.294819` `source_timestamp=2026-02-02T09:03:32Z`

## Open Questions

- Is Yao's AWS access issue fully resolved? As of the last message, he still could not create instances and was asked to coordinate a quick test.

## Related Pages

- `aws-devnet-access-guide`

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_fbe7ac61fbf25c03ef095f29b54b6386`
