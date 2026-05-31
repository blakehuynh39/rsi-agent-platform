---
title: "Hub v1 Backend"
type: "project"
slug: "projects/hub-v1-backend"
freshness: "2024-05-30T19:24:00Z"
tags:
  - "backend"
  - "data"
  - "hub"
  - "infrastructure"
owners: []
source_revision_ids:
  - "srcrev_c18eadc0f819580b076a9d3c85064036"
conflict_state: "none"
---

# Hub v1 Backend

## Summary

The Hub v1 will have a dedicated backend, preferably co-located on the same Kubernetes cluster. Data storage will split between Arweave (for NFT and IPA metadata) and a Web2 database. Specific details about the User IPA NFT Collection/Group and the web2 database need clarification.

## Claims

- The Hub v1 will have its own backend. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5) `source_document_id=srcdoc_b4469259082dfbaf9904766d817b804e` `source_revision_id=srcrev_c18eadc0f819580b076a9d3c85064036` `chunk_id=srcchunk_4dbc0afeaa4ecb56e74a8809c699d3bd` `native_locator=https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5` `source_timestamp=2024-05-30T19:24:00Z`
- Preferred to run the backend within the same Kubernetes cluster to reduce operational overhead. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5) `source_document_id=srcdoc_b4469259082dfbaf9904766d817b804e` `source_revision_id=srcrev_c18eadc0f819580b076a9d3c85064036` `chunk_id=srcchunk_4dbc0afeaa4ecb56e74a8809c699d3bd` `native_locator=https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5` `source_timestamp=2024-05-30T19:24:00Z`
- Will discuss backend setup with Andy. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5) `source_document_id=srcdoc_b4469259082dfbaf9904766d817b804e` `source_revision_id=srcrev_c18eadc0f819580b076a9d3c85064036` `chunk_id=srcchunk_4dbc0afeaa4ecb56e74a8809c699d3bd` `native_locator=https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5` `source_timestamp=2024-05-30T19:24:00Z`
- Need to determine what User IPA NFT Collection/Group is. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5) `source_document_id=srcdoc_b4469259082dfbaf9904766d817b804e` `source_revision_id=srcrev_c18eadc0f819580b076a9d3c85064036` `chunk_id=srcchunk_4dbc0afeaa4ecb56e74a8809c699d3bd` `native_locator=https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5` `source_timestamp=2024-05-30T19:24:00Z`
- NFT metadata and IPA metadata will be stored on Arweave. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5) `source_document_id=srcdoc_b4469259082dfbaf9904766d817b804e` `source_revision_id=srcrev_c18eadc0f819580b076a9d3c85064036` `chunk_id=srcchunk_4dbc0afeaa4ecb56e74a8809c699d3bd` `native_locator=https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5` `source_timestamp=2024-05-30T19:24:00Z`
- A Web2 database will be used for other data. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5) `source_document_id=srcdoc_b4469259082dfbaf9904766d817b804e` `source_revision_id=srcrev_c18eadc0f819580b076a9d3c85064036` `chunk_id=srcchunk_4dbc0afeaa4ecb56e74a8809c699d3bd` `native_locator=https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5` `source_timestamp=2024-05-30T19:24:00Z`

## Open Questions

- What is the expected schema for the Arweave metadata?
- What is the Kubernetes environment?
- What is the User IPA NFT Collection/Group?
- What other data will go in the Web2 database?
- Which specific Web2 database will be used?

## Sources

- `source_document_id`: `srcdoc_b4469259082dfbaf9904766d817b804e`
- `source_revision_id`: `srcrev_c18eadc0f819580b076a9d3c85064036`
- `source_url`: [Notion source](https://www.notion.so/Hub-v1-Backend-bf558f9cac4747068a78db85d4d287d5)
