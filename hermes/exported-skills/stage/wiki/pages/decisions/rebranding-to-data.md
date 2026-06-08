---
title: "Rebranding to Data"
type: "decision"
slug: "decisions/rebranding-to-data"
freshness: "2026-06-08T19:41:00Z"
tags:
  - "data-network"
  - "rebranding"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_b726db414a9997815e82b5779965f8e9"
conflict_state: "none"
---

# Rebranding to Data

## Summary

Overview of Story's rebranding to Data, including name changes for chain, foundation, and token.

## Claims

- Story Foundation becomes Data Foundation. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- Story, the L1 chain, becomes Data Network. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- Story native network token $IP renames to $DATA, the native token of Data Network. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- A new ERC20 wrapper of the native token will be deployed, $WDATA, with address 0xD18a56346227f25D1410F98f78234305660bB877. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- All domains with 'story' in them will transition to 'data' with 1 month supporting both versions. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- Github Repositories for network clients story (piplabs/story) and story-geth (piplabs/story-geth) will be archived, and work will continue in forks piplabs/data-network and piplabs/data-network-geth. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- The consensus client references 'story' in App Id 'story-1' and Validator Bech32 prefix 'story', but due to low visibility and usability impact, changing them via hardfork is deprioritized. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- Chain ID remains 1514. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`

## Open Questions

- Exact new RPC domains not finalized (placeholders '??' in source).

## Related Pages

- `data-migration-technical-runbook`

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_b726db414a9997815e82b5779965f8e9`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
