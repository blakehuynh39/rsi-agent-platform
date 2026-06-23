---
title: "Devnet Migration to AWS"
type: "project"
slug: "projects/devnet-migration-to-aws"
freshness: "2026-01-26T02:41:06Z"
tags:
  - "aws"
  - "devnet"
  - "migration"
owners:
  - "Tai (U07TNT9N4JC)"
  - "Yao Wang (U07KLPN0JN6)"
source_revision_ids:
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_18a91aae0f159f3c3a38c958876a5de8"
  - "srcrev_1a9a34e7011dad60e06a26f50442515b"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
conflict_state: "none"
---

# Devnet Migration to AWS

## Summary

Migration of the Story devnet environment from GCP to AWS, including provisioning infrastructure, access controls, and domain updates.

## Claims

- Yao Wang created the repository story-devnet-aws on GitHub to support the devnet migration from GCP to AWS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The new repository is based on previous internal-devnet provisioning code. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Yao asked about the history of genesis generation scripts, specifically the file update-genesis-hash.sh. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_1a9a34e7011dad60e06a26f50442515b` `chunk_id=srcchunk_ccd938e272141324d371faa5e0eac0a6` `native_locator=slack:C0547N89JUB:1768945631.194489:1768974983.803429` `source_timestamp=2026-01-21T05:56:23Z`
- Tai confirmed that update-genesis-hash.sh is a helper script with no special context requiring attention. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
- The planned updated domains for the devnet are: RPC at devnet0.storyrpc.io and Explorer at devnet0.storyscan.io. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The AWS Server Access Guide is documented at a Notion page: https://www.notion.so/storyprotocol/AWS-Server-Access-Guide-2ee051299a5480d998e2d3bb23abd3ea. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Yao requested Tai to review the created workflow for the story-devnet-aws repository. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_18a91aae0f159f3c3a38c958876a5de8` `chunk_id=srcchunk_f488ff86049c4185d9f242bc32bab9bf` `native_locator=slack:C0547N89JUB:1768945631.194489:1769395266.877009` `source_timestamp=2026-01-26T02:41:06Z`

## Related Pages

- `page_devnet-aws-ssm-access-for-yao`

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_18e7589fd1071086ab07f188e1af0512`
