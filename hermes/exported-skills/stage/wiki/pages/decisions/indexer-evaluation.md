---
title: "Indexer Evaluation"
type: "decision"
slug: "decisions/indexer-evaluation"
freshness: "2024-02-19T20:27:00Z"
tags:
  - "data"
  - "indexer"
  - "infrastructure"
  - "vendor-comparison"
owners: []
source_revision_ids:
  - "srcrev_cb1e4f669125a0b57d245913fe63a804"
conflict_state: "none"
---

# Indexer Evaluation

## Summary

Evaluation of indexer vendors (Zetablocks and Goldsky) to power protocol APIs, on-chain analytics, and third-party developer APIs. Includes requirements, fee structure, and vendor comparison.

## Claims

- The indexer is needed to power protocol APIs and other first-party app APIs (P0 priority). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- The indexer is needed for on-chain analytics, including building metrics and dashboards to understand user activities (P1 priority). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- The indexer should allow other developers to build their API on top of the data, with a query fee (P2 priority). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- For app chain owners, the fee structure includes a maintenance fee (e.g., Goldsky ~$30K/year) and a usage fee based on queries and storage. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- Zetablocks current customers include Eigenlayer, Polygon, Sui, zkSync, Chainlink, and OpenZeppelin. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- Goldsky current customers include Zora, Arweave, Immutable, Optimism, 0x, and Berachain. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- Zetablocks offers SQL and GraphQL for developer experience, while Goldsky uses AssemblyScript and GraphQL via The Graph. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- Zetablocks supports easy on-chain analytics query and dashboard building; Goldsky does not offer this as a feature. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- Zetablocks claims 99.95% reliability; Goldsky reliability is not available and depends on the RPC provider. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`
- Zetablocks CEO admits focus on big L1/protocol customers but wants to enter the app chain market and showcase RSI as their first app chain partner. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58) `source_document_id=srcdoc_46a210a359965b0bbdc3ebe2cd259cc4` `source_revision_id=srcrev_cb1e4f669125a0b57d245913fe63a804` `chunk_id=srcchunk_dde89154fc52167c54a01cd14ad356a9` `native_locator=https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58` `source_timestamp=2024-02-19T20:27:00Z`

## Open Questions

- What additional data do we need? (AA data? 6551 data?)

## Sources

- `source_document_id`: `srcdoc_46a210a359965b0bbdc3ebe2cd259cc4`
- `source_revision_id`: `srcrev_cb1e4f669125a0b57d245913fe63a804`
- `source_url`: [Notion source](https://www.notion.so/Indexer-evaluation-In-Progress-ade05aad406945d882d82aa658be9e58)
