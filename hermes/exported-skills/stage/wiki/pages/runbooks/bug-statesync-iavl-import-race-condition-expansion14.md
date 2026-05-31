---
title: "Bug Report: Statesync IAVL Import Race Condition (expansion14 crash)"
type: "runbook"
slug: "runbooks/bug-statesync-iavl-import-race-condition-expansion14"
freshness: "2026-05-20T03:01:00Z"
tags:
  - "bug"
  - "cosmos-sdk"
  - "iavl"
  - "race-condition"
  - "statesync"
  - "story"
owners: []
source_revision_ids:
  - "srcrev_f106b01da02c019f6fc589bcf40c8178"
conflict_state: "none"
---

# Bug Report: Statesync IAVL Import Race Condition (expansion14 crash)

## Summary

Non-deterministic crash during statesync caused by a race condition in cosmos/iavl v1.2.2 import.go where an asynchronous batch write goroutine has not completed before LoadVersion() reads the tree. This results in an incomplete IAVL tree, causing FinalizeBlock to fail with 'validator does not exist' on first block execution. The bug is triggered when an IAVL store exceeds 10,000 nodes (e.g., staking store) and depends on goroutine scheduling timing.

## Claims

- The incident occurred on 2026-05-20 during a CDR devnet expansion from 15 to 17 validators. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_08a58bd5bb0614968db9a6bafa2d758e` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1` `source_timestamp=2026-05-20T03:01:00Z`
- Node expansion14 (20.46.167.28) running Story v1.6.3-stable failed with a permanent crash loop after statesync, reporting 'validator does not exist' at height 43001. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_08a58bd5bb0614968db9a6bafa2d758e` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1` `source_timestamp=2026-05-20T03:01:00Z`
- Node expansion15 (20.89.59.188) using the same snapshot and binary completed statesync successfully and operated normally. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_08a58bd5bb0614968db9a6bafa2d758e` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1` `source_timestamp=2026-05-20T03:01:00Z`
- The initial diagnosis attributing the crash to mid-flight statesync failure and an abnormally large application.db was incorrect; logs show all 6 chunks applied successfully and ABCI Info returned successfully. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_08a58bd5bb0614968db9a6bafa2d758e` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1` `source_timestamp=2026-05-20T03:01:00Z`
- The verified root cause is a race condition in cosmos/iavl v1.2.2 import.go where an asynchronous inflight goroutine writing a large batch of IAVL nodes to LevelDB has not completed before LoadVersion() reads the tree, resulting in an incomplete tree. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_cb41cc5706152ab5aac6f3a197875109` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2` `source_timestamp=2026-05-20T03:01:00Z`
- The race condition is non-deterministic; expansion15 succeeded because the inflight goroutine completed before LoadVersion(), while expansion14 failed because the goroutine was still writing. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_cb41cc5706152ab5aac6f3a197875109` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2` `source_timestamp=2026-05-20T03:01:00Z`
- This is the second occurrence of the bug; the first was on 2026-05-15 at height 11000 with 50,015 storyvaloper entries. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_cb41cc5706152ab5aac6f3a197875109` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2` `source_timestamp=2026-05-20T03:01:00Z`
- The bug is triggered when any IAVL store exceeds 10,000 nodes, which is typical for a staking store with multiple validators and delegations. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_cb41cc5706152ab5aac6f3a197875109` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2` `source_timestamp=2026-05-20T03:01:00Z`
- The bug location is in cosmos/iavl v1.2.2 import.go line 216, where Commit() fails to wait for the inflight goroutine before calling batch.WriteSync() and LoadVersion(). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_cb41cc5706152ab5aac6f3a197875109` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2` `source_timestamp=2026-05-20T03:01:00Z`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_08a58bd5bb0614968db9a6bafa2d758e` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-1` `source_timestamp=2026-05-20T03:01:00Z`
- The IAVL tree stores Merkle AVL tree nodes in LevelDB indexed by version and nonce; snapshot export uses logical ExportNode structures and import must reassign nonces, recompute hashes, and rebuild references, which introduced the async batch optimization and race condition. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-3) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_69cae60e6b80d5e22d8d44df7bc571b5` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-3` `source_timestamp=2026-05-20T03:01:00Z`
- The proposed IAVL-layer fix is to wait for the inflight goroutine before WriteSync() in Commit(), ensuring all nodes are flushed before LoadVersion() reads the tree. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_cb41cc5706152ab5aac6f3a197875109` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2` `source_timestamp=2026-05-20T03:01:00Z`
- Affected versions include all chains using cosmos/iavl v1.2.2 or similar versions with async batch import combined with statesync, and the severity is that statesync produces corrupt state silently. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2) `source_document_id=srcdoc_807f827cbcf149cf9840098fad924661` `source_revision_id=srcrev_f106b01da02c019f6fc589bcf40c8178` `chunk_id=srcchunk_cb41cc5706152ab5aac6f3a197875109` `native_locator=https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540#chunk-2` `source_timestamp=2026-05-20T03:01:00Z`

## Related Pages

- `iavl-import-async-batch-optimization`
- `statesync-recovery-runbook`
- `story-devnet-expansion-procedure`

## Sources

- `source_document_id`: `srcdoc_807f827cbcf149cf9840098fad924661`
- `source_revision_id`: `srcrev_f106b01da02c019f6fc589bcf40c8178`
- `source_url`: [Notion source](https://www.notion.so/Bug-Report-Statesync-IAVL-Import-Race-Condition-expansion14-crash-366051299a548119ad77c72f46813540)
