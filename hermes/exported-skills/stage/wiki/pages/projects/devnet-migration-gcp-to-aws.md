---
title: "Devnet Migration: GCP to AWS (story-devnet-aws)"
type: "project"
slug: "projects/devnet-migration-gcp-to-aws"
freshness: "2026-02-02T19:49:38Z"
tags:
  - "aws"
  - "devnet"
  - "gcp"
  - "infrastructure"
  - "migration"
owners: []
source_revision_ids:
  - "srcrev_0ec51856e296cdc24889fa34ee9ae9dc"
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_1ab9dc873a0a5b0cbf488ed01643cc4e"
  - "srcrev_2f4c634ece1214f473aae27110313f7d"
  - "srcrev_5688ed7112b0849869fb0d8ffe52a7c6"
  - "srcrev_601e41e9ed3b697e900e25209a80b017"
  - "srcrev_872ac962a8d920428d31ed1d51bd54d1"
  - "srcrev_927830e489716b6394b2f77580855ffe"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
  - "srcrev_a8df9f25874bfae9ae77b57d47fbd1a2"
  - "srcrev_ebb79a70a9425edc788a67443be7b77a"
conflict_state: "none"
---

# Devnet Migration: GCP to AWS (story-devnet-aws)

## Summary

Yao is leading the creation of a new infrastructure repository (story-devnet-aws) to migrate the devnet environment from GCP to AWS, reusing existing provisioning code. The setup involves updating domains, configuring AWS access permissions, and finalizing the workflow.

## Claims

- New repository story-devnet-aws created for devnet migration from GCP to AWS, based on previous internal-devnet provisioning code. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- AWS Server Access Guide available at https://www.notion.so/storyprotocol/AWS-Server-Access-Guide-2ee051299a5480d998e2d3bb23abd3ea. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Updated domains: RPC devnet0.storyrpc.io, Explorer devnet0.storyscan.io. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Genesis generation script `update-genesis-hash.sh` is a helper script requiring no special context; workflow can be finalized. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_1ab9dc873a0a5b0cbf488ed01643cc4e` `chunk_id=srcchunk_9b852392be582866ca56d149a08ade10` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977211.264509` `source_timestamp=2026-01-21T06:33:31Z`
- Kaito provided Yao systemAdmin access to the devnet AWS account. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_a8df9f25874bfae9ae77b57d47fbd1a2` `chunk_id=srcchunk_c74548a12c0e0f60dd1398c519e32f5e` `native_locator=slack:C0547N89JUB:1768945631.194489:1768976762.443739` `source_timestamp=2026-01-21T06:26:02Z`
- Yao was granted maintain role on the story-devnet-aws repo for GHA access. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_927830e489716b6394b2f77580855ffe` `chunk_id=srcchunk_bf3cd60ca7c966b78445192dc380d362` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416521.546069` `source_timestamp=2026-01-26T08:35:21Z`
- A PR was created (https://github.com/storyprotocol/AWS-Organization/pull/8) and merged for Yao's additional AWS access. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_872ac962a8d920428d31ed1d51bd54d1` `chunk_id=srcchunk_46c5df2c13efc22f0f3ecaa32d7746a5` `native_locator=slack:C0547N89JUB:1768945631.194489:1769653148.452189` `source_timestamp=2026-01-29T02:19:08Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_601e41e9ed3b697e900e25209a80b017` `chunk_id=srcchunk_dc7e6248f888153eb010c3d42a6b5679` `native_locator=slack:C0547N89JUB:1768945631.194489:1769655382.071069` `source_timestamp=2026-01-29T02:56:22Z`
- Yao still encountered ssm:StartSession permission denied errors despite being granted SystemAdministrator and AWS-QA-Devnet-Access roles. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_ebb79a70a9425edc788a67443be7b77a` `chunk_id=srcchunk_07b687cd06bd70fb344a36bc9e56057f` `native_locator=slack:C0547N89JUB:1768945631.194489:1770026264.209509` `source_timestamp=2026-02-02T09:57:44Z`
- Kaito resolved permissions, pointing to the policy at https://github.com/storyprotocol/AWS-Organization/blob/0bab852b32a88f662402d060589d755bd0063e01/accounts/story-dev-ou/main.tf#L33. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_2f4c634ece1214f473aae27110313f7d` `chunk_id=srcchunk_fec44376ee4a1fccaa26319e9b3c6214` `native_locator=slack:C0547N89JUB:1768945631.194489:1769706890.638689` `source_timestamp=2026-01-29T17:14:50Z`
- Yao reported still unable to create EC2 instances after permissions resolution. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_0ec51856e296cdc24889fa34ee9ae9dc` `chunk_id=srcchunk_81c176a5a7bbde1142e8a3c13b09f6bb` `native_locator=slack:C0547N89JUB:1768945631.194489:1770023007.294819` `source_timestamp=2026-02-02T09:03:32Z`
- Kaito offered to test permissions directly with Yao when online. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_5688ed7112b0849869fb0d8ffe52a7c6` `chunk_id=srcchunk_f610ffe3195af412db97011365bf3dfc` `native_locator=slack:C0547N89JUB:1768945631.194489:1770061778.789199` `source_timestamp=2026-02-02T19:49:38Z`

## Open Questions

- What specific EC2 permissions are needed for full sync testing?
- Why were permissions still insufficient after granting?

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_3e8a3e0609b87d2faba709a7f5c8ad5c`
