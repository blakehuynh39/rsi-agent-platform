---
title: "Story NFT Contract Design"
type: "system"
slug: "systems/story-nft-contract-design"
freshness: "2024-10-01T23:19:00Z"
tags:
  - "badge"
  - "contract"
  - "factory"
  - "ipa"
  - "nft"
  - "template"
owners:
  - "Story Protocol Admin"
source_revision_ids:
  - "srcrev_ca51d169fee004c433e3a050489ca555"
conflict_state: "none"
---

# Story NFT Contract Design

## Summary

Design for the Story NFT Contract system, which uses Factory and Template patterns to allow ecosystem partners to deploy NFT contracts. It defines a Badge IPA structure with Root, Organization, and Badge levels, and outlines use cases for admins, partners, and end users.

## Claims

- The Badge NFT Contract supports both the Factory and Template patterns, allowing Story Ecosystem partners to deploy NFT contracts through the Factory. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-1) `source_document_id=srcdoc_e43861b48ace4141a4925884f1da5b1f` `source_revision_id=srcrev_ca51d169fee004c433e3a050489ca555` `chunk_id=srcchunk_75799d4535fa260ff91774e26553f93a` `native_locator=https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-1` `source_timestamp=2024-10-01T23:19:00Z`
- As a story protocol admin, I can allow only whitelisted partners to deploy Story NFT contracts through the NFT Factory. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-1) `source_document_id=srcdoc_e43861b48ace4141a4925884f1da5b1f` `source_revision_id=srcrev_ca51d169fee004c433e3a050489ca555` `chunk_id=srcchunk_75799d4535fa260ff91774e26553f93a` `native_locator=https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-1` `source_timestamp=2024-10-01T23:19:00Z`
- The Badge IPA structure consists of three levels: Root Badge IPA owned by Story, Organization IPA owned by each partner, and Badge NFT/IPA owned by end users. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-1) `source_document_id=srcdoc_e43861b48ace4141a4925884f1da5b1f` `source_revision_id=srcrev_ca51d169fee004c433e3a050489ca555` `chunk_id=srcchunk_75799d4535fa260ff91774e26553f93a` `native_locator=https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-1` `source_timestamp=2024-10-01T23:19:00Z`
- The deployStoryNft function verifies caller whitelisting, mints an Org NFT, registers an Org IPA, deploys an Org NFT contract via the Story Badge Template ERC721, and initializes it. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-2) `source_document_id=srcdoc_e43861b48ace4141a4925884f1da5b1f` `source_revision_id=srcrev_ca51d169fee004c433e3a050489ca555` `chunk_id=srcchunk_235cb7a6e24529fa904ca6f703cffb88` `native_locator=https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-2` `source_timestamp=2024-10-01T23:19:00Z`
- The registerNftTemplate function is a governance function callable only by Story Admin to register a new Story NFT Template. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-2) `source_document_id=srcdoc_e43861b48ace4141a4925884f1da5b1f` `source_revision_id=srcrev_ca51d169fee004c433e3a050489ca555` `chunk_id=srcchunk_235cb7a6e24529fa904ca6f703cffb88` `native_locator=https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-2` `source_timestamp=2024-10-01T23:19:00Z`
- The Badge Template NFT is a Soulbound token (ERC721 + ERC5192) where only end users with a partner's signature can mint, and it registers a Badge IPA upon minting. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-2) `source_document_id=srcdoc_e43861b48ace4141a4925884f1da5b1f` `source_revision_id=srcrev_ca51d169fee004c433e3a050489ca555` `chunk_id=srcchunk_235cb7a6e24529fa904ca6f703cffb88` `native_locator=https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-2` `source_timestamp=2024-10-01T23:19:00Z`
- The Base Story NFT Contract defines a common interface for all Story NFT Contracts, requiring ERC721 compliance, a contractURI function, and a mint function that registers an IPA. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-2) `source_document_id=srcdoc_e43861b48ace4141a4925884f1da5b1f` `source_revision_id=srcrev_ca51d169fee004c433e3a050489ca555` `chunk_id=srcchunk_235cb7a6e24529fa904ca6f703cffb88` `native_locator=https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce#chunk-2` `source_timestamp=2024-10-01T23:19:00Z`

## Sources

- `source_document_id`: `srcdoc_e43861b48ace4141a4925884f1da5b1f`
- `source_revision_id`: `srcrev_ca51d169fee004c433e3a050489ca555`
- `source_url`: [Notion source](https://www.notion.so/Story-NFT-Contract-Design-112051299a5480bda8dcc696867619ce)
