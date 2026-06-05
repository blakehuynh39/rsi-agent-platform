---
title: "$DATA Token and Data Network Rebranding"
type: "project"
slug: "projects/data-token-rebrand"
freshness: "2026-06-05T18:42:00Z"
tags:
  - "migration"
  - "rebranding"
  - "token"
owners: []
source_revision_ids:
  - "srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a"
conflict_state: "none"
---

# $DATA Token and Data Network Rebranding

## Summary

Rebranding of Story Foundation, L1 chain, and token from $IP to $DATA, with technical migration steps for exchanges and node operators.

## Claims

- Story Foundation is renamed to Data Foundation. `claim:claim_1_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`
- The Story L1 blockchain is renamed to Data Network. `claim:claim_1_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`
- The native network token $IP is renamed to $DATA, and a new ERC20 wrapper token $WDATA is deployed. `claim:claim_1_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`
- All domains containing 'story' (e.g., storyrpc.io) will transition to 'data' (e.g., datarpc.io) with a 1-month period supporting both versions. `claim:claim_1_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`
- The GitHub repositories story and story-geth are archived, and development continues in the forks data-network and data-network-geth. `claim:claim_1_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`
- No hardfork is required because the EVM execution layer does not reference the native token ticker or network name, and the chain ID remains 1514. `claim:claim_1_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`
- References to 'story' remain in the Cosmos consensus client (app id 'story-1' and validator bech32 prefix 'story') and in the genesis JSON; changing them is deprioritized due to complexity and low impact. `claim:claim_1_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`
- WDATA is an independently deployed ERC20 contract, and WIP is immutable, allowing both wrapper tokens to coexist without breaking changes. `claim:claim_1_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`
- Centralized exchanges must update the token ticker to $DATA, network name to Data Network, and RPC domain (e.g., from mainnet.storyrpc.io to a new datanetworkrpc.io address). `claim:claim_1_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a` `chunk_id=srcchunk_518bea93e0801fd733098176577f708b` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T18:42:00Z`

## Open Questions

- What are the final, confirmed RPC domain names for mainnet and aeneid (the chunk shows '??' indicating uncertainty)?

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_2cc82fef0d3e319ce9bcf632a3f8f75a`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
