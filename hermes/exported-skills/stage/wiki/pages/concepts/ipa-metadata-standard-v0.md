---
title: "IPA Metadata Standard (v0)"
type: "concept"
slug: "concepts/ipa-metadata-standard-v0"
freshness: "2025-08-27T15:16:00Z"
tags:
  - "ipa"
  - "metadata"
  - "standard"
owners: []
source_revision_ids:
  - "srcrev_5e101873e2d9deb8bc8828b45aa944ae"
conflict_state: "none"
---

# IPA Metadata Standard (v0)

## Summary

Off-chain and on-chain metadata standard for IP Assets (IPA) in the RSI protocol, defining JSON structure, required fields, and immutability process.

## Claims

- The off-chain metadata should not hold redundant protocol information if it can already be queried on-chain (e.g., licensing information from LicensingModule/Registry contracts). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-1) `source_document_id=srcdoc_4b1f34186fc066087eed013ead4179a6` `source_revision_id=srcrev_5e101873e2d9deb8bc8828b45aa944ae` `chunk_id=srcchunk_272523133da0805550d158b781da2139` `native_locator=https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-1` `source_timestamp=2025-08-27T15:16:00Z`
- The metadata is stored off-chain in JSON format, mutable by default, with a 2-step process to make it immutable. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-1) `source_document_id=srcdoc_4b1f34186fc066087eed013ead4179a6` `source_revision_id=srcrev_5e101873e2d9deb8bc8828b45aa944ae` `chunk_id=srcchunk_272523133da0805550d158b781da2139` `native_locator=https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-1` `source_timestamp=2025-08-27T15:16:00Z`
- The off-chain metadata includes required properties for the Portal: title, description, dateTime, image, and creators. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-1) `source_document_id=srcdoc_4b1f34186fc066087eed013ead4179a6` `source_revision_id=srcrev_5e101873e2d9deb8bc8828b45aa944ae` `chunk_id=srcchunk_272523133da0805550d158b781da2139` `native_locator=https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-1` `source_timestamp=2025-08-27T15:16:00Z`
- The standard defines structured types: Relationship (ipId, type), Creator (name, address, description, image, socialMedia, role, contributionPercent), Media (name, url, mimeType), Attribute (key, value), and App (id, name, website). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-2) `source_document_id=srcdoc_4b1f34186fc066087eed013ead4179a6` `source_revision_id=srcrev_5e101873e2d9deb8bc8828b45aa944ae` `chunk_id=srcchunk_f1bc13a607e3031f68048083adcced4f` `native_locator=https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-2` `source_timestamp=2025-08-27T15:16:00Z`
- Example use-cases include a Harry Potter book (literature), a simple IP character from a PFP, and a Damien Hirst physical painting. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-2) `source_document_id=srcdoc_4b1f34186fc066087eed013ead4179a6` `source_revision_id=srcrev_5e101873e2d9deb8bc8828b45aa944ae` `chunk_id=srcchunk_f1bc13a607e3031f68048083adcced4f` `native_locator=https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-2` `source_timestamp=2025-08-27T15:16:00Z`
- On-chain metadata includes tokenURI, metadataURI, tokenURIHash, and metadataURIHash, all mutable until the IPOwner sets an immutable flag. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-3) `source_document_id=srcdoc_4b1f34186fc066087eed013ead4179a6` `source_revision_id=srcrev_5e101873e2d9deb8bc8828b45aa944ae` `chunk_id=srcchunk_3a642051db10021bd6a73ba87a2694d9` `native_locator=https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9#chunk-3` `source_timestamp=2025-08-27T15:16:00Z`

## Open Questions

- How could DeSo protocols be integrated and used for creator attribution?
- In what cases would we want to use an alternative image for the IPA?
- Should metadata mutability depend on licensing/policy terms?
- Who should be able to set the IPA metadata? (NFT owner, original creator, or both?)

## Sources

- `source_document_id`: `srcdoc_4b1f34186fc066087eed013ead4179a6`
- `source_revision_id`: `srcrev_5e101873e2d9deb8bc8828b45aa944ae`
- `source_url`: [Notion source](https://www.notion.so/IPA-Metadata-Standard-v0-538a5c7c0e6e4045a85105efd85196b9)
