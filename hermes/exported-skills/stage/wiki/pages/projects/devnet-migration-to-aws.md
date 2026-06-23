---
title: "Devnet Migration to AWS"
type: "project"
slug: "projects/devnet-migration-to-aws"
freshness: "2026-02-02T09:57:44Z"
tags:
  - "aws"
  - "devnet"
  - "infrastructure"
  - "migration"
owners:
  - "yao.wang@piplabs.xyz"
source_revision_ids:
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_1ab9dc873a0a5b0cbf488ed01643cc4e"
  - "srcrev_601e41e9ed3b697e900e25209a80b017"
  - "srcrev_872ac962a8d920428d31ed1d51bd54d1"
  - "srcrev_927830e489716b6394b2f77580855ffe"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
  - "srcrev_a8df9f25874bfae9ae77b57d47fbd1a2"
  - "srcrev_ebb79a70a9425edc788a67443be7b77a"
  - "srcrev_f3d249a79629e8a22cafdef0bc303e75"
conflict_state: "none"
---

# Devnet Migration to AWS

## Summary

Migration of the devnet environment from GCP to AWS, using a new repository story-devnet-aws based on internal-devnet provisioning code, with updated public endpoints devnet0.storyrpc.io and devnet0.storyscan.io.

## Claims

- A new repository story-devnet-aws was created to support devnet migration from GCP to AWS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The new devnet will use domains devnet0.storyrpc.io (RPC) and devnet0.storyscan.io (Explorer). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- AWS server access guide is available at https://www.notion.so/storyprotocol/AWS-Server-Access-Guide-2ee051299a5480d998e2d3bb23abd3ea. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The genesis generation script update-genesis-hash.sh is a helper with no special context required for integration. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_1ab9dc873a0a5b0cbf488ed01643cc4e` `chunk_id=srcchunk_9b852392be582866ca56d149a08ade10` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977211.264509` `source_timestamp=2026-01-21T06:33:31Z`
- Yao (yao.wang@piplabs.xyz) was granted SystemAdministrator access to the story-devnet AWS account. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_a8df9f25874bfae9ae77b57d47fbd1a2` `chunk_id=srcchunk_c74548a12c0e0f60dd1398c519e32f5e` `native_locator=slack:C0547N89JUB:1768945631.194489:1768976762.443739` `source_timestamp=2026-01-21T06:26:02Z`
- Yao requested additional permissions to launch EC2 instances for full sync testing. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_f3d249a79629e8a22cafdef0bc303e75` `chunk_id=srcchunk_91937e509941fbad6c184e3d30759e15` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416327.392909` `source_timestamp=2026-01-26T08:36:40Z`
- An AWS-Organization PR #8 was created to adjust Yao's permissions, and it merged. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_872ac962a8d920428d31ed1d51bd54d1` `chunk_id=srcchunk_46c5df2c13efc22f0f3ecaa32d7746a5` `native_locator=slack:C0547N89JUB:1768945631.194489:1769653148.452189` `source_timestamp=2026-01-29T02:19:08Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_601e41e9ed3b697e900e25209a80b017` `chunk_id=srcchunk_dc7e6248f888153eb010c3d42a6b5679` `native_locator=slack:C0547N89JUB:1768945631.194489:1769655382.071069` `source_timestamp=2026-01-29T02:56:22Z`
- Despite the merge, Yao still encountered SSM permission errors when trying to start sessions to instances. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_ebb79a70a9425edc788a67443be7b77a` `chunk_id=srcchunk_07b687cd06bd70fb344a36bc9e56057f` `native_locator=slack:C0547N89JUB:1768945631.194489:1770026264.209509` `source_timestamp=2026-02-02T09:57:44Z`
- Yao was given maintain role on the story-devnet-aws repository to access GitHub Actions. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_927830e489716b6394b2f77580855ffe` `chunk_id=srcchunk_bf3cd60ca7c966b78445192dc380d362` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416521.546069` `source_timestamp=2026-01-26T08:35:21Z`

## Open Questions

- Are EC2 instance launch permissions for Yao fully resolved? A test was scheduled but outcome not recorded.

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_7d3e4512de69af7ea79665df3531433a`
