---
title: "Iliad Code Walkthrough"
type: "concept"
slug: "concepts/iliad-code-walkthrough"
freshness: "2026-05-05T06:36:33Z"
tags:
  - "abci"
  - "cosmos-sdk"
  - "engine-api"
  - "halo"
  - "rewards"
  - "staking"
owners: []
source_revision_ids:
  - "srcrev_83b42ed45f37bbcb759e3ef5937adab6"
conflict_state: "none"
---

# Iliad Code Walkthrough

## Summary

Overview of the Iliad codebase structure, covering Cosmos SDK modules, ABCI functions, the Engine API, and the Halo application lifecycle including staking and reward distribution.

## Claims

- The Iliad code walkthrough covers Cosmos SDK structures (modules, keepers), core ABCI functions, and the Engine API. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171) `source_document_id=srcdoc_3a32031a4c38458d0f5e76629e4ca8a0` `source_revision_id=srcrev_83b42ed45f37bbcb759e3ef5937adab6` `chunk_id=srcchunk_03e2e9cdd74d18f99c9f257c98b0dc40` `native_locator=https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171` `source_timestamp=2026-05-05T06:36:33Z`
- Halo is a Cosmos application that connects ABCI to the Engine API, including the application lifecycle (node start, InitGenesis, different block phases) and a new module added. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171) `source_document_id=srcdoc_3a32031a4c38458d0f5e76629e4ca8a0` `source_revision_id=srcrev_83b42ed45f37bbcb759e3ef5937adab6` `chunk_id=srcchunk_03e2e9cdd74d18f99c9f257c98b0dc40` `native_locator=https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171` `source_timestamp=2026-05-05T06:36:33Z`
- The walkthrough covers different staking functions: deposit, withdraw, and createValidator. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171) `source_document_id=srcdoc_3a32031a4c38458d0f5e76629e4ca8a0` `source_revision_id=srcrev_83b42ed45f37bbcb759e3ef5937adab6` `chunk_id=srcchunk_03e2e9cdd74d18f99c9f257c98b0dc40` `native_locator=https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171` `source_timestamp=2026-05-05T06:36:33Z`
- The walkthrough covers reward distribution and the Cosmos API. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171) `source_document_id=srcdoc_3a32031a4c38458d0f5e76629e4ca8a0` `source_revision_id=srcrev_83b42ed45f37bbcb759e3ef5937adab6` `chunk_id=srcchunk_03e2e9cdd74d18f99c9f257c98b0dc40` `native_locator=https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171` `source_timestamp=2026-05-05T06:36:33Z`
- In the Cosmos SDK, callers include CometBFT through ABCI calls (baseApp), RPC calls (msg server and query server in modules), and internal calls that go directly to the keeper, while external calls go through msg server and query server. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171) `source_document_id=srcdoc_3a32031a4c38458d0f5e76629e4ca8a0` `source_revision_id=srcrev_83b42ed45f37bbcb759e3ef5937adab6` `chunk_id=srcchunk_03e2e9cdd74d18f99c9f257c98b0dc40` `native_locator=https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171` `source_timestamp=2026-05-05T06:36:33Z`
- The Context in the Cosmos SDK stores references to states and transactions. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171) `source_document_id=srcdoc_3a32031a4c38458d0f5e76629e4ca8a0` `source_revision_id=srcrev_83b42ed45f37bbcb759e3ef5937adab6` `chunk_id=srcchunk_03e2e9cdd74d18f99c9f257c98b0dc40` `native_locator=https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171` `source_timestamp=2026-05-05T06:36:33Z`

## Sources

- `source_document_id`: `srcdoc_3a32031a4c38458d0f5e76629e4ca8a0`
- `source_revision_id`: `srcrev_83b42ed45f37bbcb759e3ef5937adab6`
- `source_url`: [Notion source](https://www.notion.so/Iliad-code-walkthrough-37e3df6c2b9d4a4db631bf1c40e61171)
