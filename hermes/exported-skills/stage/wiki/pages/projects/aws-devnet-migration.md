---
title: "AWS Devnet Migration"
type: "project"
slug: "projects/aws-devnet-migration"
freshness: "2026-02-02T19:49:38Z"
tags:
  - "aws"
  - "devnet"
  - "infrastructure"
  - "migration"
owners:
  - "Lucas"
  - "Yao"
source_revision_ids:
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_1a9a34e7011dad60e06a26f50442515b"
  - "srcrev_2f4c634ece1214f473aae27110313f7d"
  - "srcrev_5688ed7112b0849869fb0d8ffe52a7c6"
  - "srcrev_872ac962a8d920428d31ed1d51bd54d1"
  - "srcrev_927830e489716b6394b2f77580855ffe"
  - "srcrev_96e1c8dec0e48cf339fde5c656edfd81"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
  - "srcrev_a8df9f25874bfae9ae77b57d47fbd1a2"
  - "srcrev_b0d8c6211859790272dd6233da18c19b"
  - "srcrev_ebb79a70a9425edc788a67443be7b77a"
  - "srcrev_f3d249a79629e8a22cafdef0bc303e75"
conflict_state: "none"
---

# AWS Devnet Migration

## Summary

Migration of the devnet from GCP to AWS. Repository story-devnet-aws contains provisioning code. Domains updated to devnet0.storyrpc.io and devnet0.storyscan.io. Addressing access and permissions for team members to enable testing.

## Claims

- A new GitHub repository story-devnet-aws was created for devnet migration from GCP to AWS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Updated devnet domains: RPC at devnet0.storyrpc.io and Explorer at devnet0.storyscan.io. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- An AWS Server Access Guide is available at https://www.notion.so/storyprotocol/AWS-Server-Access-Guide-2ee051299a5480d998e2d3bb23abd3ea. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The genesis generation helper script `update-genesis-hash.sh` is only a helper and requires no special context for finalizing the workflow. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_1a9a34e7011dad60e06a26f50442515b` `chunk_id=srcchunk_ccd938e272141324d371faa5e0eac0a6` `native_locator=slack:C0547N89JUB:1768945631.194489:1768974983.803429` `source_timestamp=2026-01-21T05:56:23Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
- Yao was granted systemAdmin access to the devnet AWS account. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_a8df9f25874bfae9ae77b57d47fbd1a2` `chunk_id=srcchunk_c74548a12c0e0f60dd1398c519e32f5e` `native_locator=slack:C0547N89JUB:1768945631.194489:1768976762.443739` `source_timestamp=2026-01-21T06:26:02Z`
- Yao requested permission to launch EC2 instances for full sync testing. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_f3d249a79629e8a22cafdef0bc303e75` `chunk_id=srcchunk_91937e509941fbad6c184e3d30759e15` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416327.392909` `source_timestamp=2026-01-26T08:36:40Z`
- Yao requested maintain role on the GitHub repo for GitHub Actions access. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_b0d8c6211859790272dd6233da18c19b` `chunk_id=srcchunk_33adf021f969f72c414e1b4f6c1dea99` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416442.559699` `source_timestamp=2026-01-26T08:34:10Z`
- Lucas granted Yao maintain role on the GitHub repo. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_927830e489716b6394b2f77580855ffe` `chunk_id=srcchunk_bf3cd60ca7c966b78445192dc380d362` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416521.546069` `source_timestamp=2026-01-26T08:35:21Z`
- A PR (AWS-Organization #8) was created to add Yao's access permissions. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_872ac962a8d920428d31ed1d51bd54d1` `chunk_id=srcchunk_46c5df2c13efc22f0f3ecaa32d7746a5` `native_locator=slack:C0547N89JUB:1768945631.194489:1769653148.452189` `source_timestamp=2026-01-29T02:19:08Z`
- The PR was merged, but a Terraform workflow/security scan failure was observed; it was resolved. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_96e1c8dec0e48cf339fde5c656edfd81` `chunk_id=srcchunk_e65853fc7e3da45ffe21da8f70e6b8b1` `native_locator=slack:C0547N89JUB:1768945631.194489:1769673955.427289` `source_timestamp=2026-01-29T08:05:55Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_2f4c634ece1214f473aae27110313f7d` `chunk_id=srcchunk_fec44376ee4a1fccaa26319e9b3c6214` `native_locator=slack:C0547N89JUB:1768945631.194489:1769706890.638689` `source_timestamp=2026-01-29T17:14:50Z`
- Despite the access grant, Yao encountered AccessDeniedException when attempting to start an SSM session to an EC2 instance, using two different roles. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_ebb79a70a9425edc788a67443be7b77a` `chunk_id=srcchunk_07b687cd06bd70fb344a36bc9e56057f` `native_locator=slack:C0547N89JUB:1768945631.194489:1770026264.209509` `source_timestamp=2026-02-02T09:57:44Z`
- Lucas offered to run a quick test to resolve the SSM access issue. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_5688ed7112b0849869fb0d8ffe52a7c6` `chunk_id=srcchunk_f610ffe3195af412db97011365bf3dfc` `native_locator=slack:C0547N89JUB:1768945631.194489:1770061778.789199` `source_timestamp=2026-02-02T19:49:38Z`

## Open Questions

- Are there any remaining IAM policy gaps preventing instance launch?
- Is SSM access fully resolved to allow Yao to launch and access EC2 instances for full sync testing?

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_31f0e318883e40c6df62f9a539058c31`
