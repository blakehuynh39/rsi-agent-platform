---
title: "NFT Metadata Reveal"
type: "system"
slug: "systems/nft-metadata-reveal"
freshness: "2026-05-05T05:40:53Z"
tags:
  - "asteri"
  - "metadata"
  - "nft"
  - "reveal"
owners: []
source_revision_ids:
  - "srcrev_4203a7381755fb230d6dce1508cc1be5"
conflict_state: "none"
---

# NFT Metadata Reveal

## Summary

Design for the Asteri NFT metadata reveal process, including metadata structure, generation rules, and two reveal methods.

## Claims

- Asteri NFT metadata includes attributes: Picture (SVG), Type (uint: Moon, Planet, Comet), Location (uint[] array of coordinates), Resources (string[] of resource names), Creatures (string[] of creature names), and Etchings (string of Etchings ID). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6) `source_document_id=srcdoc_7534a4ea15f340a16c4e621c3601858e` `source_revision_id=srcrev_4203a7381755fb230d6dce1508cc1be5` `chunk_id=srcchunk_cbe67b152acac13198ff1b3662586ffa` `native_locator=https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6` `source_timestamp=2026-05-05T05:40:53Z`
- Metadata generation uses existing scripts to produce all metadata based on rarity percentages. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6) `source_document_id=srcdoc_7534a4ea15f340a16c4e621c3601858e` `source_revision_id=srcrev_4203a7381755fb230d6dce1508cc1be5` `chunk_id=srcchunk_cbe67b152acac13198ff1b3662586ffa` `native_locator=https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6` `source_timestamp=2026-05-05T05:40:53Z`
- All 500 Asteri NFTs are generated on-chain in advance before the metadata reveal event, not during the reveal process. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6) `source_document_id=srcdoc_7534a4ea15f340a16c4e621c3601858e` `source_revision_id=srcrev_4203a7381755fb230d6dce1508cc1be5` `chunk_id=srcchunk_cbe67b152acac13198ff1b3662586ffa` `native_locator=https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6` `source_timestamp=2026-05-05T05:40:53Z`
- Reveal Method 1 (Per-NFT Reveal): User approves NFT to Mystery Box contract, calls open(tokenId), contract burns the NFT, requests Chainlink VRF random number, waits for callback, picks an Asteri NFT based on the random number, and transfers it back to the user. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6) `source_document_id=srcdoc_7534a4ea15f340a16c4e621c3601858e` `source_revision_id=srcrev_4203a7381755fb230d6dce1508cc1be5` `chunk_id=srcchunk_cbe67b152acac13198ff1b3662586ffa` `native_locator=https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6` `source_timestamp=2026-05-05T05:40:53Z`
- A second reveal method utilizing ERC... is mentioned but details are incomplete in the source. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6) `source_document_id=srcdoc_7534a4ea15f340a16c4e621c3601858e` `source_revision_id=srcrev_4203a7381755fb230d6dce1508cc1be5` `chunk_id=srcchunk_cbe67b152acac13198ff1b3662586ffa` `native_locator=https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6` `source_timestamp=2026-05-05T05:40:53Z`

## Open Questions

- How are the 500 Asteri NFTs pre-generated and stored on-chain?
- What are the specific rarity percentages for each metadata attribute?
- What is the full specification for Reveal Method 2 utilizing ERC...?

## Sources

- `source_document_id`: `srcdoc_7534a4ea15f340a16c4e621c3601858e`
- `source_revision_id`: `srcrev_4203a7381755fb230d6dce1508cc1be5`
- `source_url`: [Notion source](https://www.notion.so/NFT-Metadata-Reveal-519cc569c68a4b1ba813133a87ae23a6)
