---
title: "Subnet Management Console API Design"
type: "system"
slug: "systems/subnet-management-console-api"
freshness: "2025-11-04T04:37:00Z"
tags:
  - "api"
  - "poseidon"
  - "subnet"
owners: []
source_revision_ids:
  - "srcrev_7c518adb1a10e30ac1947fc6784f03e1"
conflict_state: "none"
---

# Subnet Management Console API Design

## Summary

Design for the Subnet Management Console API, a REST data layer aggregating blockchain events for monitoring workflows, activities, workers, and task queues.

## Claims

- The Subnet Management Console API serves as a read-only data layer for developers and operators to monitor workflows, activities, workers, and task queues in the subnet. `claim:claim_smc_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-1) `source_document_id=srcdoc_c7eb629d764e4785381685f0da17ab63` `source_revision_id=srcrev_7c518adb1a10e30ac1947fc6784f03e1` `chunk_id=srcchunk_41d11bd8e31192f103aa078228f459ef` `native_locator=https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-1` `source_timestamp=2025-11-04T04:37:00Z`
- API base URL is /api/v1, authentication was previously under discussion (shown as struck-through Bearer Token JWT), pagination via ?page and ?pageSize, filtering via query params, timestamps in ISO 8601, data freshness near real-time. `claim:claim_smc_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-1) `source_document_id=srcdoc_c7eb629d764e4785381685f0da17ab63` `source_revision_id=srcrev_7c518adb1a10e30ac1947fc6784f03e1` `chunk_id=srcchunk_41d11bd8e31192f103aa078228f459ef` `native_locator=https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-1` `source_timestamp=2025-11-04T04:37:00Z`
- Data flow: Smart Contracts emit events ingested by an Indexer Backend into an Aggregated DB, then an API Layer exposes read-only REST/GraphQL endpoints for the Developer Console UI. `claim:claim_smc_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-3) `source_document_id=srcdoc_c7eb629d764e4785381685f0da17ab63` `source_revision_id=srcrev_7c518adb1a10e30ac1947fc6784f03e1` `chunk_id=srcchunk_99a8ff090d213865143b96aab6f31258` `native_locator=https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-3` `source_timestamp=2025-11-04T04:37:00Z`
- The aggregator is designed to pre-compute dashboard metrics (workflow counters, success rates, durations, active workers, total staked) for fast lookups instead of scanning raw events. `claim:claim_smc_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-5) `source_document_id=srcdoc_c7eb629d764e4785381685f0da17ab63` `source_revision_id=srcrev_7c518adb1a10e30ac1947fc6784f03e1` `chunk_id=srcchunk_6c7096988d32dc7681fc77f84887f419` `native_locator=https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-5` `source_timestamp=2025-11-04T04:37:00Z`
- Workflow and activity views should use snapshot/status tables to avoid reconstructing state from multiple event tables on each request. `claim:claim_smc_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-5) `source_document_id=srcdoc_c7eb629d764e4785381685f0da17ab63` `source_revision_id=srcrev_7c518adb1a10e30ac1947fc6784f03e1` `chunk_id=srcchunk_6c7096988d32dc7681fc77f84887f419` `native_locator=https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-5` `source_timestamp=2025-11-04T04:37:00Z`
- Open questions include: handling of multiple subnets, full-text/indexed search support, worker metadata beyond address, exposing WebSocket endpoints, and indefinite vs. pruned event retention. `claim:claim_smc_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-5) `source_document_id=srcdoc_c7eb629d764e4785381685f0da17ab63` `source_revision_id=srcrev_7c518adb1a10e30ac1947fc6784f03e1` `chunk_id=srcchunk_6c7096988d32dc7681fc77f84887f419` `native_locator=https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131#chunk-5` `source_timestamp=2025-11-04T04:37:00Z`

## Open Questions

- Event retention and pruning policy
- Full-text/indexed search requirements for workflows, activities, workers
- Multi-subnet support scope
- WebSocket stream exposure from indexer
- Worker off-chain metadata tracking needs

## Related Pages

- `projects/poseidon-engineering-home`

## Sources

- `source_document_id`: `srcdoc_c7eb629d764e4785381685f0da17ab63`
- `source_revision_id`: `srcrev_7c518adb1a10e30ac1947fc6784f03e1`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Management-Console-API-Design-288051299a54805baaebebddc1807131)
