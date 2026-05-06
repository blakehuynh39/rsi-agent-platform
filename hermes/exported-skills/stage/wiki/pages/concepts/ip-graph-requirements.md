---
title: "IP Graph Requirements"
type: "concept"
slug: "concepts/ip-graph-requirements"
freshness: "2026-05-05T06:36:26Z"
tags:
  - "ip-graph"
  - "precompile"
  - "requirements"
owners: []
source_revision_ids:
  - "srcrev_97af278f1741ae1dc3cef71f27e57f88"
conflict_state: "none"
---

# IP Graph Requirements

## Summary

Requirements for the IP Graph precompile, including support for multiple parents, large ancestor counts, scalability, ACL, and dynamic gas costs.

## Claims

- The IP Graph must support up to 8 parent nodes (P0 priority). `claim:claim_ipgraph_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa) `source_document_id=srcdoc_8e252c1e515cf07abd1a2bb883c80f80` `source_revision_id=srcrev_97af278f1741ae1dc3cef71f27e57f88` `chunk_id=srcchunk_c51eb9ba9fb784d5bd7e3b7d16a62e3a` `native_locator=https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa` `source_timestamp=2026-05-05T06:36:26Z`
- The IP Graph must support tracing more than 1000 ancestor nodes (P0 priority). `claim:claim_ipgraph_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa) `source_document_id=srcdoc_8e252c1e515cf07abd1a2bb883c80f80` `source_revision_id=srcrev_97af278f1741ae1dc3cef71f27e57f88` `chunk_id=srcchunk_c51eb9ba9fb784d5bd7e3b7d16a62e3a` `native_locator=https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa` `source_timestamp=2026-05-05T06:36:26Z`
- The protocol should scale to handle 1 million IP Assets (IPA) efficiently (P1 priority). `claim:claim_ipgraph_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa) `source_document_id=srcdoc_8e252c1e515cf07abd1a2bb883c80f80` `source_revision_id=srcrev_97af278f1741ae1dc3cef71f27e57f88` `chunk_id=srcchunk_c51eb9ba9fb784d5bd7e3b7d16a62e3a` `native_locator=https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa` `source_timestamp=2026-05-05T06:36:26Z`
- Benchmark testing is required to validate performance and scalability, including testing load (CPU, IO) on validators (P1 priority). `claim:claim_ipgraph_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa) `source_document_id=srcdoc_8e252c1e515cf07abd1a2bb883c80f80` `source_revision_id=srcrev_97af278f1741ae1dc3cef71f27e57f88` `chunk_id=srcchunk_c51eb9ba9fb784d5bd7e3b7d16a62e3a` `native_locator=https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa` `source_timestamp=2026-05-05T06:36:26Z`
- Access Control List (ACL) support is planned: only whitelisted addresses/accounts can access the precompile, to be added via governance, with security audit needed (P1 priority). `claim:claim_ipgraph_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa) `source_document_id=srcdoc_8e252c1e515cf07abd1a2bb883c80f80` `source_revision_id=srcrev_97af278f1741ae1dc3cef71f27e57f88` `chunk_id=srcchunk_c51eb9ba9fb784d5bd7e3b7d16a62e3a` `native_locator=https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa` `source_timestamp=2026-05-05T06:36:26Z`
- Dynamic gas cost calculation for IP Graph precompile functions is required, with reference to a Polygon EIP (P1 priority). `claim:claim_ipgraph_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa) `source_document_id=srcdoc_8e252c1e515cf07abd1a2bb883c80f80` `source_revision_id=srcrev_97af278f1741ae1dc3cef71f27e57f88` `chunk_id=srcchunk_c51eb9ba9fb784d5bd7e3b7d16a62e3a` `native_locator=https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa` `source_timestamp=2026-05-05T06:36:26Z`

## Related Pages

- `concepts/ip-graph-node-verification-merkle-tree`
- `projects/l1-devnet-june-2025-deliverables`

## Sources

- `source_document_id`: `srcdoc_8e252c1e515cf07abd1a2bb883c80f80`
- `source_revision_id`: `srcrev_97af278f1741ae1dc3cef71f27e57f88`
- `source_url`: [Notion source](https://www.notion.so/IP-Graph-dfd3e08ae70d4ba69244a9c425a2e1aa)
