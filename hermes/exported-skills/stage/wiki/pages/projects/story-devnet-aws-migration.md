---
title: "Devnet Migration from GCP to AWS"
type: "project"
slug: "projects/story-devnet-aws-migration"
freshness: "2026-02-02T09:57:44Z"
tags:
  - "aws"
  - "devnet"
  - "migration"
  - "permissions"
owners:
  - "Lucas"
  - "Yao Wang"
source_revision_ids:
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_601e41e9ed3b697e900e25209a80b017"
  - "srcrev_872ac962a8d920428d31ed1d51bd54d1"
  - "srcrev_927830e489716b6394b2f77580855ffe"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
  - "srcrev_ebb79a70a9425edc788a67443be7b77a"
conflict_state: "none"
---

# Devnet Migration from GCP to AWS

## Summary

Project to migrate Story Protocol's devnet from GCP to AWS using the story-devnet-aws repository. Involves provisioning, permissions setup, and domain updates.

## Claims

- Yao Wang created a new repository story-devnet-aws based on previous internal-devnet provisioning code to support the devnet migration from GCP to AWS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Updated domains will be devnet0.storyrpc.io for RPC and devnet0.storyscan.io for the explorer. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The genesis generation script update-genesis-hash.sh is a helper script that does not involve any context requiring special attention. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
- Lucas was given maintain role in the story-devnet-aws repository to gain GitHub Actions (GHA) access. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_927830e489716b6394b2f77580855ffe` `chunk_id=srcchunk_bf3cd60ca7c966b78445192dc380d362` `native_locator=slack:C0547N89JUB:1768945631.194489:1769416521.546069` `source_timestamp=2026-01-26T08:35:21Z`
- Yao was initially granted SystemAdministrator and AWS-QA-Devnet-Access roles in the story-devnet AWS account, but encountered permission errors for ssm:StartSession during instance connection attempts. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_ebb79a70a9425edc788a67443be7b77a` `chunk_id=srcchunk_07b687cd06bd70fb344a36bc9e56057f` `native_locator=slack:C0547N89JUB:1768945631.194489:1770026264.209509` `source_timestamp=2026-02-02T09:57:44Z`
- A pull request in the AWS-Organization repository was created and merged to adjust Yao's permissions for launching instances, though final verification is pending. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_872ac962a8d920428d31ed1d51bd54d1` `chunk_id=srcchunk_46c5df2c13efc22f0f3ecaa32d7746a5` `native_locator=slack:C0547N89JUB:1768945631.194489:1769653148.452189` `source_timestamp=2026-01-29T02:19:08Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_601e41e9ed3b697e900e25209a80b017` `chunk_id=srcchunk_dc7e6248f888153eb010c3d42a6b5679` `native_locator=slack:C0547N89JUB:1768945631.194489:1769655382.071069` `source_timestamp=2026-01-29T02:56:22Z`

## Open Questions

- Are the SSM permissions fully operational? Final test with Lucas awaited as of last message.
- What is the overall status of the devnet migration to AWS?

## Related Pages

- `aws-organisation-iam`
- `story-devnet`
- `story-node-provisioning`

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_872ac962a8d920428d31ed1d51bd54d1`
