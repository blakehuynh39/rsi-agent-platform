---
title: "Data Migration Technical Runbook"
type: "runbook"
slug: "runbooks/data-migration-technical-runbook"
freshness: "2026-06-08T19:41:00Z"
tags:
  - "data-network"
  - "exchange"
  - "migration"
  - "token"
  - "wallet"
owners: []
source_revision_ids:
  - "srcrev_b726db414a9997815e82b5779965f8e9"
conflict_state: "none"
---

# Data Migration Technical Runbook

## Summary

Technical steps required for exchanges, wallets, and integrations during the Data rebranding.

## Claims

- For centralized exchanges, no action regarding chain id is required; it remains 1514. `claim:claim_2_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- RPC domain will change from mainnet.storyrpc.io and aeneid.storyrpc.io to mainnet.datanetworkrpc.io and aeneid.datanetworkrpc.io (exact URLs pending confirmation). `claim:claim_2_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_548bce5069cc47f5ecbcff48a211b178` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-1` `source_timestamp=2026-06-08T19:41:00Z`
- Update ticker to $DATA in backend and frontend. `claim:claim_2_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_51654300ef036f6508fa9a945ff10ecc` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T19:41:00Z`
- Update network name from 'Story' to 'Data Network' in backend and frontend. `claim:claim_2_4` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_51654300ef036f6508fa9a945ff10ecc` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T19:41:00Z`
- Update graphic materials for project and token. `claim:claim_2_5` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_51654300ef036f6508fa9a945ff10ecc` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T19:41:00Z`
- Update domain for block explorer to datanetwokscan?? (exact domain TBD). `claim:claim_2_6` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_51654300ef036f6508fa9a945ff10ecc` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T19:41:00Z`
- WDATA can be bridged to BSC using the new Layer Zero bridge for WDATA available at stargate.finance (URL with placeholders, contract addresses TBD). `claim:claim_2_7` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_51654300ef036f6508fa9a945ff10ecc` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T19:41:00Z`
- Proof of Creativity protocol and IP Portal require no action; they still support WIP, and SDK integrations require no changes. `claim:claim_2_8` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_51654300ef036f6508fa9a945ff10ecc` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T19:41:00Z`
- Balances of native tokens are not affected. `claim:claim_2_9` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_51654300ef036f6508fa9a945ff10ecc` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T19:41:00Z`
- If using chain list repos from ethereum-lists/chains, wevm/viem, or DefiLlama/chainlist, point to the latest commit. `claim:claim_2_10` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_b726db414a9997815e82b5779965f8e9` `chunk_id=srcchunk_51654300ef036f6508fa9a945ff10ecc` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893#chunk-2` `source_timestamp=2026-06-08T19:41:00Z`

## Open Questions

- Exact new RPC URLs and block explorer domain still have placeholders.
- WDATA bridge contract addresses (BSC side WDATA OTF, Story side WDATA OTFAdapter) not specified.

## Related Pages

- `rebranding-to-data`

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_b726db414a9997815e82b5779965f8e9`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
