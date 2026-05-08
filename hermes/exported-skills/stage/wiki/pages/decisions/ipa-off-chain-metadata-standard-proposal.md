---
title: "IPA Off-Chain Metadata Standard Proposal"
type: "decision"
slug: "decisions/ipa-off-chain-metadata-standard-proposal"
freshness: "2024-03-24T07:49:00Z"
tags:
  - "ipa"
  - "metadata"
  - "nft"
  - "off-chain"
  - "standard"
owners: []
source_revision_ids:
  - "srcrev_7a8c625de687d935ae0348a320ae8585"
conflict_state: "none"
---

# IPA Off-Chain Metadata Standard Proposal

## Summary

Proposes a metadata standard for IP Assets (IPAs) stored off-chain (e.g., IPFS, Arweave) to supplement on-chain data. Two alternative JSON schemas are under consideration: one using generic attributes and another using typed extensions.

## Claims

- IPA metadata is intended for frontend display and to supplement information not queryable on-chain (e.g., from modules, registries). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9) `source_document_id=srcdoc_cf55e615ab53c18906146cc2dab1233e` `source_revision_id=srcrev_7a8c625de687d935ae0348a320ae8585` `chunk_id=srcchunk_d14c01a2a5ef9fff305c749bca82ddf3` `native_locator=https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9` `source_timestamp=2024-03-24T07:49:00Z`
- Metadata related to licensing, the underlying NFT, and royalties should be queried via their respective contract functions (e.g., uri(), tokenUri()). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9) `source_document_id=srcdoc_cf55e615ab53c18906146cc2dab1233e` `source_revision_id=srcrev_7a8c625de687d935ae0348a320ae8585` `chunk_id=srcchunk_d14c01a2a5ef9fff305c749bca82ddf3` `native_locator=https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9` `source_timestamp=2024-03-24T07:49:00Z`
- The IPA metadata is assumed to be stored off-chain (e.g., IPFS, Arweave), similar to NFT metadata, but could also be stored on-chain. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9) `source_document_id=srcdoc_cf55e615ab53c18906146cc2dab1233e` `source_revision_id=srcrev_7a8c625de687d935ae0348a320ae8585` `chunk_id=srcchunk_d14c01a2a5ef9fff305c749bca82ddf3` `native_locator=https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9` `source_timestamp=2024-03-24T07:49:00Z`
- Two alternative JSON schemas are proposed for the IPA metadata: one using a generic 'attributes' array for arbitrary key-value pairs, and another using typed extensions (e.g., LIT for literature, C2PA for AIGC). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9) `source_document_id=srcdoc_cf55e615ab53c18906146cc2dab1233e` `source_revision_id=srcrev_7a8c625de687d935ae0348a320ae8585` `chunk_id=srcchunk_d14c01a2a5ef9fff305c749bca82ddf3` `native_locator=https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9` `source_timestamp=2024-03-24T07:49:00Z`
- The generic attributes proposal includes fields for title, description, creation date, creators with social media links, media/content array, and an app identifier. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9) `source_document_id=srcdoc_cf55e615ab53c18906146cc2dab1233e` `source_revision_id=srcrev_7a8c625de687d935ae0348a320ae8585` `chunk_id=srcchunk_d14c01a2a5ef9fff305c749bca82ddf3` `native_locator=https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9` `source_timestamp=2024-03-24T07:49:00Z`
- The extensions proposal includes fields for title, image, creation date, creators, content array, and supports domain-specific extensions such as literature (ISBN, Publisher, Genre), AIGC (C2PA, model, contentHash), film credits, and platform-specific metadata (e.g., NetflixMovieID). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9) `source_document_id=srcdoc_cf55e615ab53c18906146cc2dab1233e` `source_revision_id=srcrev_7a8c625de687d935ae0348a320ae8585` `chunk_id=srcchunk_d14c01a2a5ef9fff305c749bca82ddf3` `native_locator=https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9` `source_timestamp=2024-03-24T07:49:00Z`

## Open Questions

- How should collisions between IPA metadata and underlying NFT metadata be handled?
- Should a Story watermark be enforced in the metadata?
- Should the metadata link to the underlying NFT via an 'ipMetadata' field?
- Should the metadata standard use generic attributes or typed extensions?

## Sources

- `source_document_id`: `srcdoc_cf55e615ab53c18906146cc2dab1233e`
- `source_revision_id`: `srcrev_7a8c625de687d935ae0348a320ae8585`
- `source_url`: [Notion source](https://www.notion.so/Discussion-3-23-ac6c3004097a450c87e6c3599991e1c9)
