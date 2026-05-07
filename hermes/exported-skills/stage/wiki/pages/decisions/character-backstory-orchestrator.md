---
title: "Character Backstory Orchestrator"
type: "decision"
slug: "decisions/character-backstory-orchestrator"
freshness: "2023-06-30T15:40:00Z"
tags:
  - "architecture"
  - "backstory"
  - "character"
  - "nft"
  - "orchestrator"
owners: []
source_revision_ids:
  - "srcrev_efefbef9c2fd06306e95e48a5cbe9cd9"
conflict_state: "none"
---

# Character Backstory Orchestrator

## Summary

Design options for linking character NFTs to their backstory StoryNFTs, with trade-offs between on-chain and off-chain metadata.

## Claims

- Use cases include: creating a character and its backstory, updating the backstory of the user's own character (strikethrough, possibly deprecated), and retrieving the backstory of any character. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-1) `source_document_id=srcdoc_64c5d89d5661604683d49acb02619757` `source_revision_id=srcrev_efefbef9c2fd06306e95e48a5cbe9cd9` `chunk_id=srcchunk_185cb13511e3d01cb7616f9933abf8c9` `native_locator=https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-1` `source_timestamp=2023-06-30T15:40:00Z`
- Use Case 2 (update backstory of own character) is struck through, suggesting it may be deprecated or not implemented. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-1) `source_document_id=srcdoc_64c5d89d5661604683d49acb02619757` `source_revision_id=srcrev_efefbef9c2fd06306e95e48a5cbe9cd9` `chunk_id=srcchunk_185cb13511e3d01cb7616f9933abf8c9` `native_locator=https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-1` `source_timestamp=2023-06-30T15:40:00Z`
- Option 1: Backstory is a normal StoryNFT. The Character NFT links to its backstory via a `back_story` attribute in its off-chain metadata. This requires two calls from the frontend when creating a character and its backstory, because the backend needs the newly created StoryNFT information before creating the Character NFT metadata. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-1) `source_document_id=srcdoc_64c5d89d5661604683d49acb02619757` `source_revision_id=srcrev_efefbef9c2fd06306e95e48a5cbe9cd9` `chunk_id=srcchunk_185cb13511e3d01cb7616f9933abf8c9` `native_locator=https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-1` `source_timestamp=2023-06-30T15:40:00Z`
- Option 2 is presented as an alternative design, but its textual description is not available in the source; only a screenshot image is referenced. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-1) `source_document_id=srcdoc_64c5d89d5661604683d49acb02619757` `source_revision_id=srcrev_efefbef9c2fd06306e95e48a5cbe9cd9` `chunk_id=srcchunk_185cb13511e3d01cb7616f9933abf8c9` `native_locator=https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-1` `source_timestamp=2023-06-30T15:40:00Z`
  - citation: [Notion source](https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-2) `source_document_id=srcdoc_64c5d89d5661604683d49acb02619757` `source_revision_id=srcrev_efefbef9c2fd06306e95e48a5cbe9cd9` `chunk_id=srcchunk_ba4316df3b6e71ef31f73e749eda94c3` `native_locator=https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019#chunk-2` `source_timestamp=2023-06-30T15:40:00Z`

## Open Questions

- What is the design of Option 2 for linking character and backstory?

## Sources

- `source_document_id`: `srcdoc_64c5d89d5661604683d49acb02619757`
- `source_revision_id`: `srcrev_efefbef9c2fd06306e95e48a5cbe9cd9`
- `source_url`: [Notion source](https://www.notion.so/Character-Backstory-Orchestrator-0389b0a727f848dbaa70024f5864a019)
