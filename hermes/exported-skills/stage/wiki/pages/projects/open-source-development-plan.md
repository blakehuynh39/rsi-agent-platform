---
title: "Open Source Development Plan"
type: "project"
slug: "projects/open-source-development-plan"
freshness: "2023-07-03T21:08:00Z"
tags:
  - "api"
  - "developer-ecosystem"
  - "open-source"
  - "protocol"
owners: []
source_revision_ids:
  - "srcrev_ed2f86d9445e0c6e99af225738dc4bc9"
conflict_state: "none"
---

# Open Source Development Plan

## Summary

Plan for open-sourcing components of the Story Protocol backend, including core protocol APIs, franchise toolings, and internal toolings, with steps for repo structure, licensing, and community guidelines.

## Claims

- Open sourcing software creates transparency and builds trust with developer communities. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_058c2d8106521b2bc912a3bf3a78b336` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1` `source_timestamp=2023-07-03T21:08:00Z`
- The success of Story Protocol relies on an active developer ecosystem, and open sourcing is a proven way to build that ecosystem. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_058c2d8106521b2bc912a3bf3a78b336` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1` `source_timestamp=2023-07-03T21:08:00Z`
- Downsides of open sourcing include increased resource management, reduced competitive edge, and potential security vulnerabilities. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_058c2d8106521b2bc912a3bf3a78b336` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1` `source_timestamp=2023-07-03T21:08:00Z`
- Core protocol APIs (get franchise(s), get story(s), get object(s), get group(s), get content) should be open sourced to gain developers' trust. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_058c2d8106521b2bc912a3bf3a78b336` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1` `source_timestamp=2023-07-03T21:08:00Z`
- Franchise features and toolings include Merkle proof (minting), sign message (sign in), NFT gallery (pfp), profile page (data aggregation), content storage management, backstory editing, and AIGC story. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_058c2d8106521b2bc912a3bf3a78b336` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-1` `source_timestamp=2023-07-03T21:08:00Z`
- Internal toolings (database operations, data ETL workflows) should remain private. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_2b9ee3609fb00612fda58bcd323375e3` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2` `source_timestamp=2023-07-03T21:08:00Z`
- The license for open source projects will be MIT. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_2b9ee3609fb00612fda58bcd323375e3` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2` `source_timestamp=2023-07-03T21:08:00Z`
- The vision is to maintain a transparent API layer that everyone can contribute, host and find bugs, and developers should be able to see the API code they use. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_2b9ee3609fb00612fda58bcd323375e3` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2` `source_timestamp=2023-07-03T21:08:00Z`
- The repo name will be sp-api and the logo will be Story Protocol’s logo. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_2b9ee3609fb00612fda58bcd323375e3` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2` `source_timestamp=2023-07-03T21:08:00Z`
- The code of conduct will follow the Contributor Covenant. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_2b9ee3609fb00612fda58bcd323375e3` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2` `source_timestamp=2023-07-03T21:08:00Z`
- Contributing guidelines will be based on a template from nayafia/contributing-template. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_2b9ee3609fb00612fda58bcd323375e3` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2` `source_timestamp=2023-07-03T21:08:00Z`
- The project will evaluate monorepo vs polyrepo and git submodule vs subtree for repo structure. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_2b9ee3609fb00612fda58bcd323375e3` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2` `source_timestamp=2023-07-03T21:08:00Z`
- APIs related to the protocol should be open-sourced as protocol explorer, franchise toolings kept separate, and admin APIs kept private. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2) `source_document_id=srcdoc_3717dcab16e1d0c3dcb1ea7272277c81` `source_revision_id=srcrev_ed2f86d9445e0c6e99af225738dc4bc9` `chunk_id=srcchunk_2b9ee3609fb00612fda58bcd323375e3` `native_locator=https://www.notion.so/Open-Source-Development-WIP-ed7e70490e8d41a39f256d7576c567c0#chunk-2` `source_timestamp=2023-07-03T21:08:00Z`

## Related Pages

- `systems/back-end-designs`

## Sources

- `source_document_id`: `srcdoc_bf66d0017aa23a636760fde4a5588dbb`
- `source_revision_id`: `srcrev_36e732f25c3a33e47bdd466165ccd3c1`
- `source_url`: [Notion source](https://www.notion.so/Open-Source-Development-37287546f188484796c160483a1dd9be)
