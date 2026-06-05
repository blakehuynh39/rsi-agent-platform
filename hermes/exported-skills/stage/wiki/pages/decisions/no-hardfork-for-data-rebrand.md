---
title: "No Hardfork for Data Rebrand"
type: "decision"
slug: "decisions/no-hardfork-for-data-rebrand"
freshness: "2026-06-05T21:39:00Z"
tags:
  - "chain-id"
  - "hardfork"
  - "token"
owners: []
source_revision_ids:
  - "srcrev_387e1c774ce3439b600a5ab109f7d463"
conflict_state: "none"
---

# No Hardfork for Data Rebrand

## Summary

A hardfork to change the native token ticker or network name was deprioritized because the EVM client does not reference them, consensus client references have low visibility, and preserving chain history would require block-height-dependent changes.

## Claims

- Data Network uses an EVM execution environment (data-network-geth) which does not reference the native token ticker or network name at all. `claim:claim_2_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`
- The chain ID remains 1514. `claim:claim_2_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`
- The Cosmos-based consensus client references 'story' in two places: App ID 'story-1' and validator Bech32 address prefix 'story'. `claim:claim_2_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`
- A reference to $IP exists in the genesis JSON, but must be kept to allow clients to sync from block 0. `claim:claim_2_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`
- The values are only seen by developers, are not upgradeable via the normal handler, and changes would require a block-height-dependent hardfork; complexity and low user impact led to deprioritization. `claim:claim_2_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`
- WDATA is an independently deployed ERC20 contract, and WIP is immutable; both tokens will coexist as native token wrappers without a hardfork. `claim:claim_2_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_387e1c774ce3439b600a5ab109f7d463`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
