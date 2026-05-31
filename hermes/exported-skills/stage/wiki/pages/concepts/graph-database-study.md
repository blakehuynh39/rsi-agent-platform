---
title: "Graph DB Study"
type: "concept"
slug: "concepts/graph-database-study"
freshness: "2024-09-11T04:10:00Z"
tags:
  - "database-comparison"
  - "decentralized-storage"
  - "graph-database"
owners: []
source_revision_ids:
  - "srcrev_f34e16fbb16b194bd91de418b78ea167"
conflict_state: "none"
---

# Graph DB Study

## Summary

A study comparing graph databases to relational databases, exploring decentralized graph DB concepts, and evaluating existing graph database technologies for potential use in blockchain and IP management applications.

## Claims

- Graph databases treat relationships between data as equally important to the data itself, using nodes, edges, and properties to represent and store data. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-1) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_53e4f0fef7631a5f358ffeaf0b6285cf` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-1` `source_timestamp=2024-09-11T04:10:00Z`
- Graph databases excel in handling complex relationships and offer flexibility in data modeling, but face scalability challenges and have a steeper learning curve compared to SQL. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-1) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_53e4f0fef7631a5f358ffeaf0b6285cf` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-1` `source_timestamp=2024-09-11T04:10:00Z`
- Relational databases offer maturity, reliability, and strong data integrity but perform poorly with highly connected data and have rigid schema designs. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-1) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_53e4f0fef7631a5f358ffeaf0b6285cf` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-1` `source_timestamp=2024-09-11T04:10:00Z`
- The Graph protocol uses Indexers that process blockchain logs into a PostgreSQL database with a GraphQL interface, not a native graph database. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-2) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_fafa87d5cf76599696645d60f6bfe61e` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-2` `source_timestamp=2024-09-11T04:10:00Z`
- GUN is a decentralized graph database that opts for eventual consistency, meaning conflicting updates resolve only after propagation over the network. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-2) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_fafa87d5cf76599696645d60f6bfe61e` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-2` `source_timestamp=2024-09-11T04:10:00Z`
- Dgraph features a sharded and distributed architecture with automatic data movement for shard rebalancing and supports distributed ACID transactions. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-3) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_70822236bfeb1f7a9762d0ede9a9c204` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-3` `source_timestamp=2024-09-11T04:10:00Z`
- Neo4j community edition operates as a single server without replication, while enterprise edition supports replicas. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-3) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_70822236bfeb1f7a9762d0ede9a9c204` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-3` `source_timestamp=2024-09-11T04:10:00Z`
- JanusGraph is a layer on top of other distributed databases and does not typically provide ACID transactions. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-3) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_70822236bfeb1f7a9762d0ede9a9c204` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-3` `source_timestamp=2024-09-11T04:10:00Z`
- Potential use cases for a graph database at Story include an IP legal graph, royalty graph, and semantic IP graph for classifying IP by AI models. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-2) `source_document_id=srcdoc_897e027c56c794b54ce103676f714223` `source_revision_id=srcrev_f34e16fbb16b194bd91de418b78ea167` `chunk_id=srcchunk_fafa87d5cf76599696645d60f6bfe61e` `native_locator=https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78#chunk-2` `source_timestamp=2024-09-11T04:10:00Z`

## Open Questions

- Could consensus be the 'cluster controller'? How do we do that?
- How do we verify the cluster? Could we leverage a distributed DB and add proofs?
- If we go with Polaris, could the graph DB be a cosmos module?
- Is the IPAssetStorage a precompile for the DB?

## Sources

- `source_document_id`: `srcdoc_897e027c56c794b54ce103676f714223`
- `source_revision_id`: `srcrev_f34e16fbb16b194bd91de418b78ea167`
- `source_url`: [Notion source](https://www.notion.so/Graph-DB-Study-d301730634e641de9313362b61877e78)
