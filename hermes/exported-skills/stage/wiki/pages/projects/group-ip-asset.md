---
title: "Group IP Asset"
type: "project"
slug: "projects/group-ip-asset"
freshness: "2024-10-19T02:37:00Z"
tags:
  - "feature"
  - "group"
  - "ip-asset"
  - "royalty"
owners: []
source_revision_ids:
  - "srcrev_4224da9b218769f384e820e3c58326ef"
conflict_state: "none"
---

# Group IP Asset

## Summary

Design for a feature enabling creation and management of groups of IP Assets with a shared royalty pool, common interface, and license management.

## Claims

- The feature enables the creation and management of groups of IP Assets, supporting a royalty pool for the group. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_f2bee24acf6ee082496b3ae62362658b` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1` `source_timestamp=2024-10-19T02:37:00Z`
- The IP Asset Group should function equivalently to a normal IP Asset, allowing attachment of license terms, creation of derivatives, execution with modules, and other interactions. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_f2bee24acf6ee082496b3ae62362658b` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1` `source_timestamp=2024-10-19T02:37:00Z`
- Group reward pool holds the reward tokens of the group and distributes the reward tokens to individual member IPAs. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_f2bee24acf6ee082496b3ae62362658b` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1` `source_timestamp=2024-10-19T02:37:00Z`
- The group reward are royalties that collected from children IPAs or directly pay to the Group through RoyaltyModule. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_f2bee24acf6ee082496b3ae62362658b` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1` `source_timestamp=2024-10-19T02:37:00Z`
- Ensure the IP Asset Group shares the same interface as individual IP Asset instances. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_f2bee24acf6ee082496b3ae62362658b` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1` `source_timestamp=2024-10-19T02:37:00Z`
- Enable dynamic addition, removal, and listing of IP Assets within the group. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_f2bee24acf6ee082496b3ae62362658b` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1` `source_timestamp=2024-10-19T02:37:00Z`
- Allow the IP Asset Group to have a common license applicable to all its sub IP Assets. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_f2bee24acf6ee082496b3ae62362658b` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-1` `source_timestamp=2024-10-19T02:37:00Z`
- A GroupNFT is represented as an ERC-721 NFT, standing for ownership of GroupIPAccount, and is minted by IPAssetRegistry when registering a new group. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_d3548177e7b60d1f5a32f961306362db` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2` `source_timestamp=2024-10-19T02:37:00Z`
- A new RoyaltyPoolPolicy will be introduced to distribute revenue evenly to all individual IPAs within the group. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_d3548177e7b60d1f5a32f961306362db` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2` `source_timestamp=2024-10-19T02:37:00Z`
- The RoyaltyPoolPolicy will implement two main functions: claimRevenue(address[] memberIPAs) and claimMySelf(address[] memberIPAs, address receiverAddress). `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_d3548177e7b60d1f5a32f961306362db` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2` `source_timestamp=2024-10-19T02:37:00Z`
- claimMySelf transfers tokens to a given address, and the caller must be the owner of the memberIPAs. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_d3548177e7b60d1f5a32f961306362db` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2` `source_timestamp=2024-10-19T02:37:00Z`
- The pool needs to track claim history of each IPA. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_d3548177e7b60d1f5a32f961306362db` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2` `source_timestamp=2024-10-19T02:37:00Z`
- The claim process: get claimable amount of token = total token / number of IPA, reduce amount already reduced, transfer token to each IPA. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2) `source_document_id=srcdoc_e4656780dd23f6370aa225ac5587dab5` `source_revision_id=srcrev_4224da9b218769f384e820e3c58326ef` `chunk_id=srcchunk_d3548177e7b60d1f5a32f961306362db` `native_locator=https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d#chunk-2` `source_timestamp=2024-10-19T02:37:00Z`

## Sources

- `source_document_id`: `srcdoc_e4656780dd23f6370aa225ac5587dab5`
- `source_revision_id`: `srcrev_4224da9b218769f384e820e3c58326ef`
- `source_url`: [Notion source](https://www.notion.so/Group-IP-Asset-49cf8608e0814d91827d1f7809ffcf2d)
