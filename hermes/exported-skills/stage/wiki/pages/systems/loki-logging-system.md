---
title: "Loki Logging System"
type: "system"
slug: "systems/loki-logging-system"
freshness: "2026-04-27T02:45:28Z"
tags:
  - "grafana"
  - "ip-exhaustion"
  - "logging"
  - "loki"
  - "network"
  - "observability"
owners:
  - "U07TNT9N4JC"
  - "U0AKJV8710S"
source_revision_ids:
  - "srcrev_2bda50d41e474feb68ace0f13b06b23b"
  - "srcrev_8655e4071ef87305e75d61581afb8638"
conflict_state: "none"
---

# Loki Logging System

## Summary

Centralized logging system using Loki. As of 2026-04-27, experiencing memberlist ring instability due to subnet IP exhaustion on 10.0.101.0/24, causing intra-component communication failures and potential intermittent query latency.

## Claims

- Loki's memberlist ring is unstable, with ingesters and queriers constantly marking each other as suspect and failed. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_8655e4071ef87305e75d61581afb8638` `chunk_id=srcchunk_4619c10dc2c53848be0c066cd2b66505` `native_locator=slack:C0547N89JUB:1777253413.479709:1777253572.548999` `source_timestamp=2026-04-27T01:32:52Z`
- All Loki ingesters reside on the 10.0.101.x subnet, which is experiencing IP exhaustion, leading to resource insufficiency. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_8655e4071ef87305e75d61581afb8638` `chunk_id=srcchunk_4619c10dc2c53848be0c066cd2b66505` `native_locator=slack:C0547N89JUB:1777253413.479709:1777253572.548999` `source_timestamp=2026-04-27T01:32:52Z`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_2bda50d41e474feb68ace0f13b06b23b` `chunk_id=srcchunk_e36b6f4777fcfe64bcebe1afddff3fd4` `native_locator=slack:C0547N89JUB:1777253413.479709:1777257928.628009` `source_timestamp=2026-04-27T02:45:28Z`
- TCP transport EOF errors are occurring between Loki components on the 10.0.101.x subnet. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_8655e4071ef87305e75d61581afb8638` `chunk_id=srcchunk_4619c10dc2c53848be0c066cd2b66505` `native_locator=slack:C0547N89JUB:1777253413.479709:1777253572.548999` `source_timestamp=2026-04-27T01:32:52Z`
- All Loki pods (ingesters, queriers, schedulers) are in Running/Ready state, but the system is degraded internally due to the ring instability. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_8655e4071ef87305e75d61581afb8638` `chunk_id=srcchunk_4619c10dc2c53848be0c066cd2b66505` `native_locator=slack:C0547N89JUB:1777253413.479709:1777253572.548999` `source_timestamp=2026-04-27T01:32:52Z`
- Despite the ring instability, Loki is still serving queries, but may experience intermittent latency or log gaps. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_8655e4071ef87305e75d61581afb8638` `chunk_id=srcchunk_4619c10dc2c53848be0c066cd2b66505` `native_locator=slack:C0547N89JUB:1777253413.479709:1777253572.548999` `source_timestamp=2026-04-27T01:32:52Z`
- Validator node jpe-aeneid-validator1 is successfully sending cosmovisor.service logs to Loki, with the latest entry at 01:32 UTC. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_8655e4071ef87305e75d61581afb8638` `chunk_id=srcchunk_4619c10dc2c53848be0c066cd2b66505` `native_locator=slack:C0547N89JUB:1777253413.479709:1777253572.548999` `source_timestamp=2026-04-27T01:32:52Z`

## Related Pages

- `grafana-loki-query-timeout`

## Sources

- `source_document_id`: `srcdoc_bbd1403456612e7354334747960ed325`
- `source_revision_id`: `srcrev_a859c5f6b1ade3b3b55000c71d7e4d43`
