---
title: "AWS Server Access Guide"
type: "runbook"
slug: "runbooks/aws-server-access-guide"
freshness: "2026-02-02T19:49:38Z"
tags:
  - "access"
  - "aws"
  - "devnet"
  - "ssm"
owners:
  - "U07TNT9N4JC"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_0ec51856e296cdc24889fa34ee9ae9dc"
  - "srcrev_2f4c634ece1214f473aae27110313f7d"
  - "srcrev_5688ed7112b0849869fb0d8ffe52a7c6"
  - "srcrev_872ac962a8d920428d31ed1d51bd54d1"
  - "srcrev_9b5fc1bd2a4f8cffae503abba78c15e9"
  - "srcrev_ebb79a70a9425edc788a67443be7b77a"
conflict_state: "none"
---

# AWS Server Access Guide

## Summary

Guide and troubleshooting for accessing AWS devnet servers, including SSM setup and permission policies.

## Claims

- The AWS Server Access Guide is located at https://www.notion.so/storyprotocol/AWS-Server-Access-Guide-2ee051299a5480d998e2d3bb23abd3ea. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_9b5fc1bd2a4f8cffae503abba78c15e9` `chunk_id=srcchunk_6fcd0e19afe419fd13159584446b72b9` `native_locator=slack:C0547N89JUB:1768945631.194489:1768945631.194489` `source_timestamp=2026-01-20T21:47:11Z`
- Users attempting to start an SSM session may encounter AccessDeniedException due to missing ssm:StartSession permission. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_ebb79a70a9425edc788a67443be7b77a` `chunk_id=srcchunk_07b687cd06bd70fb344a36bc9e56057f` `native_locator=slack:C0547N89JUB:1768945631.194489:1770026264.209509` `source_timestamp=2026-02-02T09:57:44Z`
- Tony provided a permission policy for Lucas via PR https://github.com/storyprotocol/AWS-Organization/pull/8. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_872ac962a8d920428d31ed1d51bd54d1` `chunk_id=srcchunk_46c5df2c13efc22f0f3ecaa32d7746a5` `native_locator=slack:C0547N89JUB:1768945631.194489:1769653148.452189` `source_timestamp=2026-01-29T02:19:08Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_2f4c634ece1214f473aae27110313f7d` `chunk_id=srcchunk_fec44376ee4a1fccaa26319e9b3c6214` `native_locator=slack:C0547N89JUB:1768945631.194489:1769706890.638689` `source_timestamp=2026-01-29T17:14:50Z`
- Despite the PR being merged, Lucas still encountered permission errors, and a live test was planned to verify resolution. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_0ec51856e296cdc24889fa34ee9ae9dc` `chunk_id=srcchunk_81c176a5a7bbde1142e8a3c13b09f6bb` `native_locator=slack:C0547N89JUB:1768945631.194489:1770023007.294819` `source_timestamp=2026-02-02T09:03:32Z`
  - citation: `source_document_id=srcdoc_21912519959226e3397c0d1ec8c0dd39` `source_revision_id=srcrev_5688ed7112b0849869fb0d8ffe52a7c6` `chunk_id=srcchunk_f610ffe3195af412db97011365bf3dfc` `native_locator=slack:C0547N89JUB:1768945631.194489:1770061778.789199` `source_timestamp=2026-02-02T19:49:38Z`

## Related Pages

- `devnet-migration-to-aws`

## Sources

- `source_document_id`: `srcdoc_21912519959226e3397c0d1ec8c0dd39`
- `source_revision_id`: `srcrev_613cc8458f07bcfc20858840270db1be`
