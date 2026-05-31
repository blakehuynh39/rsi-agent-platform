---
title: "SP Precompile Candidates"
type: "open_question"
slug: "open-questions/sp-precompile-candidates"
freshness: "2024-04-26T06:58:00Z"
tags:
  - "precompiles"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_519c2753e1e7b3066012ee3436e3d62a"
conflict_state: "none"
---

# SP Precompile Candidates

## Summary

Brainstorming potential precompile candidates for Story Protocol, focusing on high gas cost, high frequency operations and unique SP features like licensing registration, royalty checks, and NFT swap tax.

## Claims

- Heuristic for selecting precompile candidates: map protocol operations with high gas cost and high frequency in normal operations. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6) `source_document_id=srcdoc_e001652815e9aae7f7058b6e261be3b5` `source_revision_id=srcrev_519c2753e1e7b3066012ee3436e3d62a` `chunk_id=srcchunk_3e75de034c0a927c8dc885af490a954d` `native_locator=https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6` `source_timestamp=2024-04-26T06:58:00Z`
- SP has unique enshrinement opportunities: Licensing, Registration, License compatibility check. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6) `source_document_id=srcdoc_e001652815e9aae7f7058b6e261be3b5` `source_revision_id=srcrev_519c2753e1e7b3066012ee3436e3d62a` `chunk_id=srcchunk_3e75de034c0a927c8dc885af490a954d` `native_locator=https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6` `source_timestamp=2024-04-26T06:58:00Z`
- Royalty checks involve determining if a node is an ancestor of another node; options include a tree-searching algorithm or a compressed fingerprint per node containing ancestor data. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6) `source_document_id=srcdoc_e001652815e9aae7f7058b6e261be3b5` `source_revision_id=srcrev_519c2753e1e7b3066012ee3436e3d62a` `chunk_id=srcchunk_3e75de034c0a927c8dc885af490a954d` `native_locator=https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6` `source_timestamp=2024-04-26T06:58:00Z`
- NFT swap tax is a candidate precompile but its motivation is unclear; tax avoidance workarounds include tweaking NFT contracts or bridging out to other chains. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6) `source_document_id=srcdoc_e001652815e9aae7f7058b6e261be3b5` `source_revision_id=srcrev_519c2753e1e7b3066012ee3436e3d62a` `chunk_id=srcchunk_3e75de034c0a927c8dc885af490a954d` `native_locator=https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6` `source_timestamp=2024-04-26T06:58:00Z`

## Open Questions

- Can NFT contracts be tweaked to avoid swap tax, and is bridging a viable workaround?
- How can license compatibility check be optimized as a precompile?
- Is a tree-searching algorithm or fingerprint-based approach more efficient for royalty checks?
- Should swap tax precompile be considered post-L1 launch?
- Which specific operations should be mapped for high gas cost and frequency?
- Would NFT swap tax motivate or hinder swap volume on Renaissance?

## Sources

- `source_document_id`: `srcdoc_e001652815e9aae7f7058b6e261be3b5`
- `source_revision_id`: `srcrev_519c2753e1e7b3066012ee3436e3d62a`
- `source_url`: [Notion source](https://www.notion.so/SP-Precompiles-Candidates-Brainstorming-2438ba0841a847a89478a659d1de36b6)
