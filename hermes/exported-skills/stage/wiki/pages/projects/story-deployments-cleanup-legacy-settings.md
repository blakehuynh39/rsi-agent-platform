---
title: "Story Deployments Cleanup (Legacy Settings)"
type: "project"
slug: "projects/story-deployments-cleanup-legacy-settings"
freshness: "2026-05-20T02:33:46Z"
tags:
  - "cleanup"
  - "deployment"
  - "mainnet"
  - "private-fork"
  - "staking"
  - "storyprotocol"
owners:
  - "U04VDFP1YQ5"
  - "U0772SH7BRA"
  - "U079ZJ48D62"
source_revision_ids:
  - "srcrev_3e1bb4252d7e3a75df7bc73f59b2e94c"
  - "srcrev_b32762c65d5f3bc8a037c0607749c95a"
  - "srcrev_b46f93cc0d8c73e37206bfcb02fdb660"
conflict_state: "none"
---

# Story Deployments Cleanup (Legacy Settings)

## Summary

Discussion regarding cleaning up legacy settings in story-deployments, focusing on whether to keep the staking-api-private-fork mainnet deployment based on use1-prod.yaml while BitGo team still uses it, pending migration to new deployments.

## Claims

- A request was made to clean up story-deployments with legacy settings via PR #356. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cd5313f93d0532e49e87d54ac7d34bbf` `source_revision_id=srcrev_b32762c65d5f3bc8a037c0607749c95a` `chunk_id=srcchunk_e87b62ac19d0629d9a30706c3cbdafbc` `native_locator=slack:C0547N89JUB:1779242381.212589:1779242381.212589` `source_timestamp=2026-05-20T01:59:41Z`
- The staking-api-private-fork for mainnet is deployed based on the use1-prod.yaml configuration. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cd5313f93d0532e49e87d54ac7d34bbf` `source_revision_id=srcrev_3e1bb4252d7e3a75df7bc73f59b2e94c` `chunk_id=srcchunk_a5f6aa901500291aed5f481f2b433740` `native_locator=slack:C0547N89JUB:1779242381.212589:1779243701.385649` `source_timestamp=2026-05-20T02:21:41Z`
- The mainnet staking-api-private-fork is currently used by the BitGo team and should be retained until the new deployments replace it. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cd5313f93d0532e49e87d54ac7d34bbf` `source_revision_id=srcrev_3e1bb4252d7e3a75df7bc73f59b2e94c` `chunk_id=srcchunk_a5f6aa901500291aed5f481f2b433740` `native_locator=slack:C0547N89JUB:1779242381.212589:1779243701.385649` `source_timestamp=2026-05-20T02:21:41Z`
- The new deployments are in the final testing phase and may soon replace the existing ones. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cd5313f93d0532e49e87d54ac7d34bbf` `source_revision_id=srcrev_3e1bb4252d7e3a75df7bc73f59b2e94c` `chunk_id=srcchunk_a5f6aa901500291aed5f481f2b433740` `native_locator=slack:C0547N89JUB:1779242381.212589:1779243701.385649` `source_timestamp=2026-05-20T02:21:41Z`
- The private fork (likely staging) only uses the use1-stage configuration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cd5313f93d0532e49e87d54ac7d34bbf` `source_revision_id=srcrev_b46f93cc0d8c73e37206bfcb02fdb660` `chunk_id=srcchunk_20dc70a5f31acda7ecde2bbfcc269133` `native_locator=slack:C0547N89JUB:1779242381.212589:1779244426.040749` `source_timestamp=2026-05-20T02:33:46Z`
- The staking-api-private-fork will be removed after the migration is completed. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cd5313f93d0532e49e87d54ac7d34bbf` `source_revision_id=srcrev_b46f93cc0d8c73e37206bfcb02fdb660` `chunk_id=srcchunk_20dc70a5f31acda7ecde2bbfcc269133` `native_locator=slack:C0547N89JUB:1779242381.212589:1779244426.040749` `source_timestamp=2026-05-20T02:33:46Z`

## Open Questions

- When will the migration and removal of the private fork be completed?

## Sources

- `source_document_id`: `srcdoc_cd5313f93d0532e49e87d54ac7d34bbf`
- `source_revision_id`: `srcrev_f52332207e039e90862688bb3f70ddaf`
