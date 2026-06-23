---
title: "AWS Devnet Migration"
type: "project"
slug: "projects/aws-devnet-migration"
freshness: "2026-02-02T09:57:44Z"
tags:
  - "aws"
  - "devnet"
  - "infrastructure"
  - "migration"
owners:
  - "devnet-team"
source_revision_ids:
  - "srcrev_0ec51856e296cdc24889fa34ee9ae9dc"
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_1a9a34e7011dad60e06a26f50442515b"
  - "srcrev_2f4c634ece1214f473aae27110313f7d"
  - "srcrev_927830e489716b6394b2f77580855ffe"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
  - "srcrev_a8df9f25874bfae9ae77b57d47fbd1a2"
  - "srcrev_ebb79a70a9425edc788a67443be7b77a"
  - "srcrev_f3d249a79629e8a22cafdef0bc303e75"
conflict_state: "disputed"
---

# AWS Devnet Migration

## Summary

Project to migrate Story Protocol devnet from GCP to AWS, including provisioning code, access guides, and permission setup.

## Claims

- A new GitHub repository story-devnet-aws was created based on previous internal-devnet provisioning code for the GCP to AWS migration. `claim:repo_created` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- An AWS Server Access Guide is available at https://www.notion.so/storyprotocol/AWS-Server-Access-Guide-2ee051299a5480d998e2d3bb23abd3ea `claim:aws_access_guide` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Updated public domains will be: RPC at devnet0.storyrpc.io, Explorer at devnet0.storyscan.io. `claim:planned_domains` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The genesis generation script (update-genesis-hash.sh) is a helper script and does not require special attention for the migration workflow. `claim:genesis_script_context` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_1a9a34e7011dad60e06a26f50442515b` `chunk_id=srcchunk_ccd938e272141324d371faa5e0eac0a6` `native_locator=slack:C0547N89JUB:1768945631.194489:1768974983.803429` `source_timestamp=2026-01-21T05:56:23Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
- Yao (U07KLPN0JN6) was granted systemAdmin access to the AWS devnet account. `claim:yao_systemadmin_access` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_a8df9f25874bfae9ae77b57d47fbd1a2` `chunk_id=srcchunk_c74548a12c0e0f60dd1398c519e32f5e` `native_locator=slack:C0547N89JUB:1768945631.194489:1768976762.443739` `source_timestamp=2026-01-21T06:26:02Z`
- Yao received maintain role on the story-devnet-aws GitHub repository to access GitHub Actions. `claim:yao_repo_maintain_role` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_927830e489716b6394b2f77580855ffe` `chunk_id=srcchunk_bf3cd60ca7c966b78445192dc380d362` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416521.546069` `source_timestamp=2026-01-26T08:35:21Z`
- Yao needed additional IAM permissions to launch EC2 instances for full sync testing, specifically the ssm:StartSession action. `claim:ec2_launch_permission_needed` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_f3d249a79629e8a22cafdef0bc303e75` `chunk_id=srcchunk_91937e509941fbad6c184e3d30759e15` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416327.392909` `source_timestamp=2026-01-26T08:36:40Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_ebb79a70a9425edc788a67443be7b77a` `chunk_id=srcchunk_07b687cd06bd70fb344a36bc9e56057f` `native_locator=slack:C0547N89JUB:1768945631.194489:1770026264.209509` `source_timestamp=2026-02-02T09:57:44Z`
- The IAM permission issue for ssm:StartSession was claimed resolved on 2026-01-31, but as of 2026-02-01 Yao reported still being unable to create instances, indicating the issue may not be fully resolved. `claim:permission_resolution_status` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_2f4c634ece1214f473aae27110313f7d` `chunk_id=srcchunk_fec44376ee4a1fccaa26319e9b3c6214` `native_locator=slack:C0547N89JUB:1768945631.194489:1769706890.638689` `source_timestamp=2026-01-29T17:14:50Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_0ec51856e296cdc24889fa34ee9ae9dc` `chunk_id=srcchunk_81c176a5a7bbde1142e8a3c13b09f6bb` `native_locator=slack:C0547N89JUB:1768945631.194489:1770023007.294819` `source_timestamp=2026-02-02T09:03:32Z`

## Conflicts

- Contradictory status on whether the EC2 instance creation permission issue is resolved. One source says resolved on 2026-01-31, another says still unresolved on 2026-02-01. `claim:permission_resolution_status`
  - conflict citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_2f4c634ece1214f473aae27110313f7d` `chunk_id=srcchunk_fec44376ee4a1fccaa26319e9b3c6214` `native_locator=slack:C0547N89JUB:1768945631.194489:1769706890.638689` `source_timestamp=2026-01-29T17:14:50Z`
  - conflict citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_0ec51856e296cdc24889fa34ee9ae9dc` `chunk_id=srcchunk_81c176a5a7bbde1142e8a3c13b09f6bb` `native_locator=slack:C0547N89JUB:1768945631.194489:1770023007.294819` `source_timestamp=2026-02-02T09:03:32Z`

## Open Questions

- Have the test instances been created and full sync testing started?
- Is the EC2 instance launch permission for Yao fully functional now?
- What is the final status of the GCP to AWS migration?

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_55d554cbe6b19b319e600f12c3e98182`
