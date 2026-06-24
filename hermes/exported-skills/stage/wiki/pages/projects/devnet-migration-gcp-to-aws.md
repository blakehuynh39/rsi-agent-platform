---
title: "Devnet Migration from GCP to AWS"
type: "project"
slug: "projects/devnet-migration-gcp-to-aws"
freshness: "2026-01-26T02:41:06Z"
tags:
  - "AWS"
  - "devnet"
  - "GCP"
  - "genesis"
  - "migration"
owners:
  - "Lucas"
  - "Yao Wang"
source_revision_ids:
  - "srcrev_17ed84ab4ec41f1d3df49691402e3f3e"
  - "srcrev_18a91aae0f159f3c3a38c958876a5de8"
  - "srcrev_1a9a34e7011dad60e06a26f50442515b"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
conflict_state: "none"
---

# Devnet Migration from GCP to AWS

## Summary

Migration of Story Protocol devnet infrastructure from GCP to AWS, including a new repository, genesis scripts, and domain updates.

## Claims

- A new GitHub repository story-devnet-aws was created to support the devnet migration from GCP to AWS, based on previous internal-devnet provisioning code. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The migration will use new domains: RPC at devnet0.storyrpc.io and Explorer at devnet0.storyscan.io. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- An AWS Server Access Guide was shared on Notion for team reference. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- The repository contains a genesis generation helper script named update-genesis-hash.sh. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_1a9a34e7011dad60e06a26f50442515b` `chunk_id=srcchunk_ccd938e272141324d371faa5e0eac0a6` `native_locator=slack:C0547N89JUB:1768945631.194489:1768974983.803429` `source_timestamp=2026-01-21T05:56:23Z`
- Lucas confirmed the script is only a helper and requires no special context for the migration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_17ed84ab4ec41f1d3df49691402e3f3e` `chunk_id=srcchunk_1dfc5630af92e29896acd61dd54a5c38` `native_locator=slack:C0547N89JUB:1768945631.194489:1768977183.385329` `source_timestamp=2026-01-21T06:33:03Z`
- Yao finalized the repository workflow and requested review. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_18a91aae0f159f3c3a38c958876a5de8` `chunk_id=srcchunk_f488ff86049c4185d9f242bc32bab9bf` `native_locator=slack:C0547N89JUB:1768945631.194489:1769395266.877009` `source_timestamp=2026-01-26T02:41:06Z`

## Related Pages

- `aws-devnet-iam-permissions`

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_3c9e974b4ec3eae4b61d59bcf6c0bb26`
