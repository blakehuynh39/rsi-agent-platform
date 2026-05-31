---
title: "Serverless Design to Expose Network Status"
type: "concept"
slug: "concepts/serverless-design-expose-network-status"
freshness: "2024-09-18T17:43:00Z"
tags: []
owners: []
source_revision_ids:
  - "srcrev_de3be74aa099d8a01197dc2e744f0e86"
  - "srcrev_f3b6d4630b437a690b16689ca897ebff"
conflict_state: "none"
---

# Serverless Design to Expose Network Status

## Summary

A serverless design pattern for exposing network status, originally a Golang project on EC2, proposed to be re-architected with two AWS Lambda functions and DynamoDB for scalability and cost efficiency.

## Claims

- A serverless design exists to expose network status. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/System-Design-and-Optimization-10ce1e020d4c4d779042e70e3dec87ea) `source_document_id=srcdoc_7f5ec5cd941f2bc6249207c551c39730` `source_revision_id=srcrev_de3be74aa099d8a01197dc2e744f0e86` `chunk_id=srcchunk_803606ee533efaa29aa7c1d76016c9a0` `native_locator=https://www.notion.so/System-Design-and-Optimization-10ce1e020d4c4d779042e70e3dec87ea` `source_timestamp=2024-08-15T21:57:00Z`
- The existing system is a Golang-based project running on a dedicated EC2 server to expose network block heights and health status. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1) `source_document_id=srcdoc_561584823fc534174665582c193d3c4b` `source_revision_id=srcrev_f3b6d4630b437a690b16689ca897ebff` `chunk_id=srcchunk_24a9892e80ac83815ee66a9350f2b6f3` `native_locator=https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1` `source_timestamp=2024-09-18T17:43:00Z`
- The current system lacks a CD pipeline, has scaling concerns, stores state in memory (making sync across servers challenging), and may be overkill for the project's needs. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1) `source_document_id=srcdoc_561584823fc534174665582c193d3c4b` `source_revision_id=srcrev_f3b6d4630b437a690b16689ca897ebff` `chunk_id=srcchunk_24a9892e80ac83815ee66a9350f2b6f3` `native_locator=https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1` `source_timestamp=2024-09-18T17:43:00Z`
- The proposed architecture uses two Lambda functions: one as a data scraper that fetches metrics from geth and cometbft endpoints, processes them, and pushes to DynamoDB; the other likely serves the status via API Gateway. `claim:claim_2_3` `confidence:0.90`
  - citation: [Notion source](https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1) `source_document_id=srcdoc_561584823fc534174665582c193d3c4b` `source_revision_id=srcrev_f3b6d4630b437a690b16689ca897ebff` `chunk_id=srcchunk_24a9892e80ac83815ee66a9350f2b6f3` `native_locator=https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1` `source_timestamp=2024-09-18T17:43:00Z`
  - citation: [Notion source](https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-2) `source_document_id=srcdoc_561584823fc534174665582c193d3c4b` `source_revision_id=srcrev_f3b6d4630b437a690b16689ca897ebff` `chunk_id=srcchunk_c07ab87467d9bd7c6b9777272d3288f4` `native_locator=https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-2` `source_timestamp=2024-09-18T17:43:00Z`
- The data scraper Lambda is triggered by AWS EventBridge every 10 seconds. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1) `source_document_id=srcdoc_561584823fc534174665582c193d3c4b` `source_revision_id=srcrev_f3b6d4630b437a690b16689ca897ebff` `chunk_id=srcchunk_24a9892e80ac83815ee66a9350f2b6f3` `native_locator=https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1` `source_timestamp=2024-09-18T17:43:00Z`
- DynamoDB is extremely affordable; storing a record like {"consensus_block_height":249817,"execution_block_height":249816,"status":"Normal"} costs about $3 per month. `claim:claim_2_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1) `source_document_id=srcdoc_561584823fc534174665582c193d3c4b` `source_revision_id=srcrev_f3b6d4630b437a690b16689ca897ebff` `chunk_id=srcchunk_24a9892e80ac83815ee66a9350f2b6f3` `native_locator=https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52#chunk-1` `source_timestamp=2024-09-18T17:43:00Z`

## Sources

- `source_document_id`: `srcdoc_561584823fc534174665582c193d3c4b`
- `source_revision_id`: `srcrev_f3b6d4630b437a690b16689ca897ebff`
- `source_url`: [Notion source](https://www.notion.so/Serverless-Design-to-Expose-Network-Status-727719d60df442908b935c54fc1b6d52)
