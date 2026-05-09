---
title: "EU: API/SDK integration notes"
type: "project"
slug: "projects/eu-api-sdk-integration-notes"
freshness: "2023-10-27T10:38:00Z"
tags:
  - "api"
  - "emergence-universe"
  - "frontend"
  - "integration"
  - "sdk"
owners: []
source_revision_ids:
  - "srcrev_6b1fa48e0949cdcaab0a6a38d5e92484"
conflict_state: "none"
---

# EU: API/SDK integration notes

## Summary

Analysis of Emergence Universe features and data requirements, mapping which can be completed within the frontend, with the SDK, and which need additional API endpoints. Covers features for launch and speculative future releases.

## Claims

- The document analyzes EU features and data requirements, categorizing implementation methods as frontend, SDK, or API, with features labeled for launch being most likely required and others speculative. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- Connect Wallet feature is estimated for release/v2, implemented via frontend (possibly API), with open questions about whether wallet connection is needed for initial release and if wallet binding to backend data is required. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- Authentication for protecting against malicious requests is required for release, implemented via API, with an open question about what kind of authentication is needed. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- List Stories feature for the Landing/Stories page is required for launch, implemented via frontend, with a note that the SDK may not support scheduling content release, so unreleased story details should be defined in the frontend. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- Story data (title, image, address, author name & description, owner, excerpt, chapters, IP Assets) is required for launch and fetched via SDK, with an open question about whether author details come from a separate request. `claim:claim_2_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- Number of Reads feature (save a story as read, fetch number of reads) is required for launch, implemented via API, and should not be limited to connected users, with consideration of using IP address or cookies to avoid duplicates. `claim:claim_2_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- Reading Progress (get/save user's progress through a story) is required for launch, implemented via API and frontend, with local storage for unconnected users and comparison against DB when a wallet is connected. `claim:claim_2_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- Estimated Reading Time is required for launch, implemented via frontend, with a suggestion to do the logic in the frontend or predefine per story. `claim:claim_2_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- IP Asset detail page (Characters, Relics, Places) displaying name, image, address, description, creator info, and stories where the asset appears is required for launch and fetched via SDK. `claim:claim_2_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_057c905ebbfe8309222b4f5845130ceb` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-1` `source_timestamp=2023-10-27T10:38:00Z`
- The document suggests ideally returning author/creator info (name, description) for stories and elements, possibly via a new endpoint related to user profiles, and anticipates author/creator detail pages similar to character pages. `claim:claim_2_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-2) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_12443f675bfd9567840a431ff7f00c96` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-2` `source_timestamp=2023-10-27T10:38:00Z`
- An open question exists about whether the get character details endpoint can return the stories where the character is referenced. `claim:claim_2_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-2) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_12443f675bfd9567840a431ff7f00c96` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-2` `source_timestamp=2023-10-27T10:38:00Z`
- A suggestion was made that the 'read' counter for each story could potentially be inferred from the progress field. `claim:claim_2_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-2) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_12443f675bfd9567840a431ff7f00c96` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-2` `source_timestamp=2023-10-27T10:38:00Z`
- Design references are available for story detail pages (Figma and Notion feedback) and reading experience (Figma and Notion feedback). `claim:claim_2_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-2) `source_document_id=srcdoc_013e4fc605dc443c7aff205f14fbc61d` `source_revision_id=srcrev_6b1fa48e0949cdcaab0a6a38d5e92484` `chunk_id=srcchunk_12443f675bfd9567840a431ff7f00c96` `native_locator=https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7#chunk-2` `source_timestamp=2023-10-27T10:38:00Z`

## Open Questions

- Are there any features that require a wallet connection in the initial release?
- Are we able to return the stories where a character is referenced via the get character details endpoint?
- Can the 'read' counter be inferred from the progress field?
- Is local storage for reading progress suitable, with comparison against DB on wallet connection?
- Should read count not be limited to connected users? Do we use IP address or cookies to allow guests and avoid duplicates?
- What kind of authentication do we need?
- Will author details be fetched through a separate request? Does this come from SDK, API, or predefined?
- Will we need to bind a wallet to backend data or authenticate a connection?

## Related Pages

- `projects/emergence-universe`
- `projects/eu-pfp-nfts`
- `projects/eu-releases-tasks`
- `projects/platform-sdk-api-doc`

## Sources

- `source_document_id`: `srcdoc_013e4fc605dc443c7aff205f14fbc61d`
- `source_revision_id`: `srcrev_6b1fa48e0949cdcaab0a6a38d5e92484`
- `source_url`: [Notion source](https://www.notion.so/EU-API-SDK-integration-notes-34a20201ec534c7a924451c2a13133d7)
