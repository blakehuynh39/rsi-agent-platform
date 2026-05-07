---
title: "IP Graph API"
type: "system"
slug: "systems/ip-graph-api"
freshness: "2024-03-22T18:37:00Z"
tags:
  - "api"
  - "graph-traversal"
  - "ip-graph"
owners: []
source_revision_ids:
  - "srcrev_9ae17164bb1141d54403f22a3d933c14"
conflict_state: "none"
---

# IP Graph API

## Summary

API design for retrieving IP graph data, including root and child node queries with depth control and relationship traversal.

## Claims

- The IP Graph API provides an endpoint to get all IPs on SP. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3) `source_document_id=srcdoc_fa74d3dc4ab2c4fb82b851e875c48dac` `source_revision_id=srcrev_9ae17164bb1141d54403f22a3d933c14` `chunk_id=srcchunk_96e48569c7e17487989900be4f8b9440` `native_locator=https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3` `source_timestamp=2024-03-22T18:37:00Z`
- The graph can be retrieved from an IP Asset node, which can be either root or children. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3) `source_document_id=srcdoc_fa74d3dc4ab2c4fb82b851e875c48dac` `source_revision_id=srcrev_9ae17164bb1141d54403f22a3d933c14` `chunk_id=srcchunk_96e48569c7e17487989900be4f8b9440` `native_locator=https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3` `source_timestamp=2024-03-22T18:37:00Z`
- Root queries allow choosing depth levels with a max depth limit, and each node indicates if it still has children. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3) `source_document_id=srcdoc_fa74d3dc4ab2c4fb82b851e875c48dac` `source_revision_id=srcrev_9ae17164bb1141d54403f22a3d933c14` `chunk_id=srcchunk_96e48569c7e17487989900be4f8b9440` `native_locator=https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3` `source_timestamp=2024-03-22T18:37:00Z`
- Children queries return parents for certain levels, similar to root, and a flag can be set to return only children or only parents. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3) `source_document_id=srcdoc_fa74d3dc4ab2c4fb82b851e875c48dac` `source_revision_id=srcrev_9ae17164bb1141d54403f22a3d933c14` `chunk_id=srcchunk_96e48569c7e17487989900be4f8b9440` `native_locator=https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3` `source_timestamp=2024-03-22T18:37:00Z`
- The API can return an unformatted array of IP objects with id, parentIpIds, and childIpIds. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3) `source_document_id=srcdoc_fa74d3dc4ab2c4fb82b851e875c48dac` `source_revision_id=srcrev_9ae17164bb1141d54403f22a3d933c14` `chunk_id=srcchunk_96e48569c7e17487989900be4f8b9440` `native_locator=https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3` `source_timestamp=2024-03-22T18:37:00Z`
- The frontend graph formatted response includes ipId objects with tokenAddress, tokenId, childIpIds, parentIpIds, hasChild, hasParent, and a relationships array of source-target pairs. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3) `source_document_id=srcdoc_fa74d3dc4ab2c4fb82b851e875c48dac` `source_revision_id=srcrev_9ae17164bb1141d54403f22a3d933c14` `chunk_id=srcchunk_96e48569c7e17487989900be4f8b9440` `native_locator=https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3` `source_timestamp=2024-03-22T18:37:00Z`
- Graph traversal is performed by filtering relationships: to find parents of a node, loop relationships where target equals the node; to find children, loop where source equals the node. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3) `source_document_id=srcdoc_fa74d3dc4ab2c4fb82b851e875c48dac` `source_revision_id=srcrev_9ae17164bb1141d54403f22a3d933c14` `chunk_id=srcchunk_96e48569c7e17487989900be4f8b9440` `native_locator=https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3` `source_timestamp=2024-03-22T18:37:00Z`

## Open Questions

- What are the Simple Queries to get info?
- What is the SignWithEthereum API?

## Sources

- `source_document_id`: `srcdoc_fa74d3dc4ab2c4fb82b851e875c48dac`
- `source_revision_id`: `srcrev_9ae17164bb1141d54403f22a3d933c14`
- `source_url`: [Notion source](https://www.notion.so/Hub-APIs-40711db41bd845deb8302dd5654ccfb3)
