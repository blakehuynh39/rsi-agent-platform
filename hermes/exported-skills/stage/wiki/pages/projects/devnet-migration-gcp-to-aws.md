---
title: "Devnet Migration from GCP to AWS"
type: "project"
slug: "projects/devnet-migration-gcp-to-aws"
freshness: "2026-02-02T19:49:38Z"
tags:
  - "aws"
  - "devnet"
  - "infrastructure"
  - "migration"
owners: []
source_revision_ids:
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_1a9a34e7011dad60e06a26f50442515b"
  - "srcrev_2f4c634ece1214f473aae27110313f7d"
  - "srcrev_5688ed7112b0849869fb0d8ffe52a7c6"
  - "srcrev_601e41e9ed3b697e900e25209a80b017"
  - "srcrev_872ac962a8d920428d31ed1d51bd54d1"
  - "srcrev_96e1c8dec0e48cf339fde5c656edfd81"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
  - "srcrev_a8df9f25874bfae9ae77b57d47fbd1a2"
  - "srcrev_ebb79a70a9425edc788a67443be7b77a"
  - "srcrev_f3d249a79629e8a22cafdef0bc303e75"
conflict_state: "none"
---

# Devnet Migration from GCP to AWS

## Summary

Project to migrate the Story Protocol devnet from GCP to AWS, including a new repository, genesis generation scripts, domain updates, and AWS access grants.

## Claims

- A new repository `story-devnet-aws` was created to support the devnet migration from GCP to AWS, based on the previous internal-devnet provisioning code. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The genesis generation scripts include a helper script `update-genesis-hash.sh` that does not require special context or attention. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_1a9a34e7011dad60e06a26f50442515b` `chunk_id=srcchunk_ccd938e272141324d371faa5e0eac0a6` `native_locator=slack:C0547N89JUB:1768945631.194489:1768974983.803429` `source_timestamp=2026-01-21T05:56:23Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
- Updated domains for the migrated devnet are RPC at `devnet0.storyrpc.io` and Explorer at `devnet0.storyscan.io`. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Yao Wang was initially granted `SystemAdministrator` access to the devnet AWS account. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_a8df9f25874bfae9ae77b57d47fbd1a2` `chunk_id=srcchunk_c74548a12c0e0f60dd1398c519e32f5e` `native_locator=slack:C0547N89JUB:1768945631.194489:1768976762.443739` `source_timestamp=2026-01-21T06:26:02Z`
- Yao later required additional permissions to launch EC2 instances and use SSM for full sync testing. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_f3d249a79629e8a22cafdef0bc303e75` `chunk_id=srcchunk_91937e509941fbad6c184e3d30759e15` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416327.392909` `source_timestamp=2026-01-26T08:36:40Z`
- A pull request was created to adjust Yao’s IAM permissions, which was merged but encountered a Terraform workflow/security scan failure. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_872ac962a8d920428d31ed1d51bd54d1` `chunk_id=srcchunk_46c5df2c13efc22f0f3ecaa32d7746a5` `native_locator=slack:C0547N89JUB:1768945631.194489:1769653148.452189` `source_timestamp=2026-01-29T02:19:08Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_601e41e9ed3b697e900e25209a80b017` `chunk_id=srcchunk_dc7e6248f888153eb010c3d42a6b5679` `native_locator=slack:C0547N89JUB:1768945631.194489:1769655382.071069` `source_timestamp=2026-01-29T02:56:22Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_96e1c8dec0e48cf339fde5c656edfd81` `chunk_id=srcchunk_e65853fc7e3da45ffe21da8f70e6b8b1` `native_locator=slack:C0547N89JUB:1768945631.194489:1769673955.427289` `source_timestamp=2026-01-29T08:05:55Z`
- Despite the merged PR, Yao encountered `AccessDeniedException` when attempting `ssm:StartSession` on an EC2 instance using two different assumed roles (`SystemAdministrator` and `AWS-QA-Devnet-Access`). `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_ebb79a70a9425edc788a67443be7b77a` `chunk_id=srcchunk_07b687cd06bd70fb344a36bc9e56057f` `native_locator=slack:C0547N89JUB:1768945631.194489:1770026264.209509` `source_timestamp=2026-02-02T09:57:44Z`
- The access issue was reportedly resolved after a final adjustment; Yao was asked to test again. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_2f4c634ece1214f473aae27110313f7d` `chunk_id=srcchunk_fec44376ee4a1fccaa26319e9b3c6214` `native_locator=slack:C0547N89JUB:1768945631.194489:1769706890.638689` `source_timestamp=2026-01-29T17:14:50Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_5688ed7112b0849869fb0d8ffe52a7c6` `chunk_id=srcchunk_f610ffe3195af412db97011365bf3dfc` `native_locator=slack:C0547N89JUB:1768945631.194489:1770061778.789199` `source_timestamp=2026-02-02T19:49:38Z`
- An AWS Server Access Guide is available on Notion: https://www.notion.so/storyprotocol/AWS-Server-Access-Guide-2ee051299a5480d998e2d3bb23abd3ea `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`

## Open Questions

- Is the AWS access issue fully resolved for Yao, and does the current IAM policy grant all necessary permissions (EC2 launch, SSM) without further adjustments?
- What was the root cause of the Terraform workflow/security scan failure, and has it been fixed going forward?

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_803703432b6e74eefa9ad5615c95505f`
