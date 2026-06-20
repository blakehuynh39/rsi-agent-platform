---
title: "Grafana Loki Query Timeout Troubleshooting"
type: "runbook"
slug: "runbooks/grafana-loki-query-timeout"
freshness: "2026-04-27T03:48:16Z"
tags:
  - "grafana"
  - "logging"
  - "loki"
  - "timeout"
  - "troubleshooting"
owners:
  - "U07KLPN0JN6"
source_revision_ids:
  - "srcrev_268252ed1da019c03a82a22cc262fd3b"
  - "srcrev_8655e4071ef87305e75d61581afb8638"
conflict_state: "none"
---

# Grafana Loki Query Timeout Troubleshooting

## Summary

Grafana queries to Loki datasources may timeout due to requesting excessive raw log lines over large time ranges. Mitigation involves reducing the time window, applying a line limit, and enabling pagination.

## Claims

- Grafana queries to Loki can timeout when requesting too many raw log lines, such as up to 5,000 lines across a 6-hour time window. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_268252ed1da019c03a82a22cc262fd3b` `chunk_id=srcchunk_67f4a1fcbe9472a69e37b3b4c1a9be2a` `native_locator=slack:C0547N89JUB:1777253413.479709:1777261696.961859` `source_timestamp=2026-04-27T03:48:16Z`
- A specific failing query targeted hostname jpe-aeneid-validator1 and service_name cosmovisor.service from the loki-blockchain datasource. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_268252ed1da019c03a82a22cc262fd3b` `chunk_id=srcchunk_67f4a1fcbe9472a69e37b3b4c1a9be2a` `native_locator=slack:C0547N89JUB:1777253413.479709:1777261696.961859` `source_timestamp=2026-04-27T03:48:16Z`
- Mitigation: use a shorter time range (e.g., 30 minutes), set a maxLines limit (e.g., 3000), and enable log pagination in Grafana Explore. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_268252ed1da019c03a82a22cc262fd3b` `chunk_id=srcchunk_67f4a1fcbe9472a69e37b3b4c1a9be2a` `native_locator=slack:C0547N89JUB:1777253413.479709:1777261696.961859` `source_timestamp=2026-04-27T03:48:16Z`
- Underlying Loki ring instability may contribute to intermittent query delays, potentially exacerbating timeout symptoms. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bbd1403456612e7354334747960ed325` `source_revision_id=srcrev_8655e4071ef87305e75d61581afb8638` `chunk_id=srcchunk_4619c10dc2c53848be0c066cd2b66505` `native_locator=slack:C0547N89JUB:1777253413.479709:1777253572.548999` `source_timestamp=2026-04-27T01:32:52Z`

## Related Pages

- `loki-logging-system`

## Sources

- `source_document_id`: `srcdoc_bbd1403456612e7354334747960ed325`
- `source_revision_id`: `srcrev_a859c5f6b1ade3b3b55000c71d7e4d43`
