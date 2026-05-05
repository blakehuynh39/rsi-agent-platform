---
title: "Twitter Oracle Backend Server"
type: "system"
slug: "systems/twitter-oracle-backend-server"
freshness: "2026-05-05T06:27:45Z"
tags:
  - "backend"
  - "ethereum"
  - "oracle"
  - "twitter"
  - "web3"
owners: []
source_revision_ids:
  - "srcrev_40b8411cb37d9f1463e82d4504a691ef"
conflict_state: "none"
---

# Twitter Oracle Backend Server

## Summary

A backend server that listens for events from the AsyncTwitterUserHook contract, retrieves Twitter follower counts via the Twitter API, and calls back the requester contract with the follower count.

## Claims

- The Twitter Oracle Backend Server is being developed as part of the Async Twitter Oracle Hook project. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b) `source_document_id=srcdoc_cea12f43cbb1d6fc2059161dc8c35240` `source_revision_id=srcrev_40b8411cb37d9f1463e82d4504a691ef` `chunk_id=srcchunk_616a632a4fc71b2d37026f623df1d1fe` `native_locator=https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b` `source_timestamp=2026-05-05T06:27:45Z`
- The backend server listens for the TwitterFollowerNumRequest event emitted by the AsyncTwitterUserHook contract. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b) `source_document_id=srcdoc_cea12f43cbb1d6fc2059161dc8c35240` `source_revision_id=srcrev_40b8411cb37d9f1463e82d4504a691ef` `chunk_id=srcchunk_616a632a4fc71b2d37026f623df1d1fe` `native_locator=https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b` `source_timestamp=2026-05-05T06:27:45Z`
- The TwitterFollowerNumRequest event provides the requestId, requester address, username, callback address, and callback function signature. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b) `source_document_id=srcdoc_cea12f43cbb1d6fc2059161dc8c35240` `source_revision_id=srcrev_40b8411cb37d9f1463e82d4504a691ef` `chunk_id=srcchunk_616a632a4fc71b2d37026f623df1d1fe` `native_locator=https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b` `source_timestamp=2026-05-05T06:27:45Z`
- The callback function signature expects two parameters: bytes32 requestId and uint256 followerCount. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b) `source_document_id=srcdoc_cea12f43cbb1d6fc2059161dc8c35240` `source_revision_id=srcrev_40b8411cb37d9f1463e82d4504a691ef` `chunk_id=srcchunk_616a632a4fc71b2d37026f623df1d1fe` `native_locator=https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b` `source_timestamp=2026-05-05T06:27:45Z`
- The backend server retrieves the Twitter user's follower count using the Twitter API endpoint GET /2/users/by/username/:username, specifically the public_metrics.followers_count field. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b) `source_document_id=srcdoc_cea12f43cbb1d6fc2059161dc8c35240` `source_revision_id=srcrev_40b8411cb37d9f1463e82d4504a691ef` `chunk_id=srcchunk_616a632a4fc71b2d37026f623df1d1fe` `native_locator=https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b` `source_timestamp=2026-05-05T06:27:45Z`
- The backend server calls the specified callback function on the callback address, passing the requestId and followerCount as encoded parameters. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b) `source_document_id=srcdoc_cea12f43cbb1d6fc2059161dc8c35240` `source_revision_id=srcrev_40b8411cb37d9f1463e82d4504a691ef` `chunk_id=srcchunk_616a632a4fc71b2d37026f623df1d1fe` `native_locator=https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b` `source_timestamp=2026-05-05T06:27:45Z`
- The sample implementation uses web3.js to encode the bytes32 requestId and uint256 followerCount before calling the contract method. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b) `source_document_id=srcdoc_cea12f43cbb1d6fc2059161dc8c35240` `source_revision_id=srcrev_40b8411cb37d9f1463e82d4504a691ef` `chunk_id=srcchunk_616a632a4fc71b2d37026f623df1d1fe` `native_locator=https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b` `source_timestamp=2026-05-05T06:27:45Z`

## Sources

- `source_document_id`: `srcdoc_cea12f43cbb1d6fc2059161dc8c35240`
- `source_revision_id`: `srcrev_40b8411cb37d9f1463e82d4504a691ef`
- `source_url`: [Notion source](https://www.notion.so/Twitter-Oracle-Backend-Server-071a84657ab04b04984f7922a74cfe6b)
