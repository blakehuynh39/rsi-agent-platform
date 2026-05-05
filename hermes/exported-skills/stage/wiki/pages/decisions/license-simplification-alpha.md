---
title: "License Simplification for Alpha"
type: "decision"
slug: "decisions/license-simplification-alpha"
freshness: "2026-05-05T06:28:17Z"
tags:
  - "alpha"
  - "license"
  - "nft"
  - "udl"
owners: []
source_revision_ids:
  - "srcrev_c1bf39dc720dc2b12c66ee6a5575f975"
conflict_state: "none"
---

# License Simplification for Alpha

## Summary

Decisions to simplify the license module for the alpha release, focusing on tradable LNFT tokens, UDL-based terms, and minimal interface changes.

## Claims

- Each LNFT representing a license deal should be a tradable NFT token. No 'bound' licenses that are mere metadata in the root token; all licenses are tradeable or at least separate tokens by default. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- Metadata of the LNFT represents the license term or parameters. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- Licensor can mint the LNFT for free and list it on NFT marketplace employing various pricing strategies. Licensee can also mint the LNFT based on the condition set by the licensor, such as a predetermined minting price. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- Licensors may have different configuration for different LNFT, such as movie or merch may have very different license parameters. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- We will only support the UDL that Ben was drafting. All license terms shall come from the UDL. Add a link to the UDL in LNFT. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- For alpha, we will probably only use a few terms, not all the ones possible from UDL, to keep things simple. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- License Term has no need to be associated with any legal text. All legal text should be in one document like UDL, with conditional parameters within the UDL. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- No terms that are not enforceable on-chain or do not have any on-chain impacts. No 'pure text' terms. If it is a term, it needs to have on-chain impact. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- We need to support a few predefined type of license term: boolean, number, multiple choices. License terms can inherit the license term type to exhibit different behaviors. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`
- We don’t want to touch other part of the existing license module due to the time constraints. We shall also minimize the impact to the interface and downstream works. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290) `source_document_id=srcdoc_41dd3ddac29434a168650814d02c7de7` `source_revision_id=srcrev_c1bf39dc720dc2b12c66ee6a5575f975` `chunk_id=srcchunk_fd7f0bd1c0cb799dfb3b542bc89c278e` `native_locator=https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290` `source_timestamp=2026-05-05T06:28:17Z`

## Sources

- `source_document_id`: `srcdoc_41dd3ddac29434a168650814d02c7de7`
- `source_revision_id`: `srcrev_c1bf39dc720dc2b12c66ee6a5575f975`
- `source_url`: [Notion source](https://www.notion.so/License-Simplification-for-Alpha-8d98267d5a4d414b8d8db4f9f82c0290)
