---
title: "Story Orchestration Service (SOS)"
type: "system"
slug: "systems/story-orchestration-service"
freshness: "2026-05-21T00:56:57Z"
tags:
  - "blockchain"
  - "etl"
  - "orchestration"
  - "temporal"
owners: []
source_revision_ids:
  - "srcrev_7bd4453ad6371a64cb792e487089d987"
  - "srcrev_cac38205da1a6ca17f6951b06d206a8c"
conflict_state: "none"
---

# Story Orchestration Service (SOS)

## Summary

SOS is the orchestration layer for ETL workflows that extract on-chain events and compute derivative graphs, built on Temporal Cloud.

## Claims

- SOS uses Temporal Cloud as its SaaS orchestration engine. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_cac38205da1a6ca17f6951b06d206a8c` `chunk_id=srcchunk_41ed086ff26680f335659f5bb7631fec` `native_locator=slack:C0547N89JUB:1779322928.231139:1779325017.786919` `source_timestamp=2026-05-21T00:56:57Z`
- Approximately six worker deployments run in the EKS cluster and long-poll Temporal Cloud for ETL tasks. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_cac38205da1a6ca17f6951b06d206a8c` `chunk_id=srcchunk_41ed086ff26680f335659f5bb7631fec` `native_locator=slack:C0547N89JUB:1779322928.231139:1779325017.786919` `source_timestamp=2026-05-21T00:56:57Z`
- Manager workflows poll the chain for new blocks and spawn Processor workflows per batch, which extract on-chain events and load into PostgreSQL databases (blockchain-db, royalty-graph-db). `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_cac38205da1a6ca17f6951b06d206a8c` `chunk_id=srcchunk_41ed086ff26680f335659f5bb7631fec` `native_locator=slack:C0547N89JUB:1779322928.231139:1779325017.786919` `source_timestamp=2026-05-21T00:56:57Z`
- Downstream aggregation workflows compute derived IP and royalty graphs. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_cac38205da1a6ca17f6951b06d206a8c` `chunk_id=srcchunk_41ed086ff26680f335659f5bb7631fec` `native_locator=slack:C0547N89JUB:1779322928.231139:1779325017.786919` `source_timestamp=2026-05-21T00:56:57Z`
- Temporal Cloud costs are billed to the Temporal account, separate from infrastructure costs; workers run in our own infrastructure and poll the Temporal cloud for tasks. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_7bd4453ad6371a64cb792e487089d987` `chunk_id=srcchunk_e546de99beb0c05799884b49ff2c4f4d` `native_locator=slack:C0547N89JUB:1779322928.231139:1779324096.094899` `source_timestamp=2026-05-21T00:41:36Z`

## Open Questions

- A detailed cost versus features table for Temporal Cloud usage has not been generated yet (request pending).

## Related Pages

- `temporal-self-hosting-migration`

## Sources

- `source_document_id`: `srcdoc_f053c43795827a147f78c95534d4ccc9`
- `source_revision_id`: `srcrev_a0d061bafa97d39fd2bc09e63e0c91bb`
