---
title: "Monitoring Stack"
type: "system"
slug: "systems/monitoring-stack"
freshness: "2024-10-27T18:42:00Z"
tags:
  - "alertmanager"
  - "grafana"
  - "loki"
  - "monitoring"
  - "observability"
  - "pagerduty"
  - "prometheus"
owners: []
source_revision_ids:
  - "srcrev_63044eb02327896049b768b6cf22894f"
conflict_state: "none"
---

# Monitoring Stack

## Summary

Design and implementation overview of the monitoring stack, covering aggressive and passive monitoring strategies, core components (Node Exporter, Prometheus, Grafana, Alertmanager, PagerDuty, Loki), and automation requirements.

## Claims

- The monitoring strategy includes an aggressive approach (scraping hardware/application metrics via Prometheus, visualized in Grafana with active alerting) and a passive approach (publishing basic metrics to third-party services like Better Stack or Uptime). `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_7bbebd9d33fe9164e019228c43b40225` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1` `source_timestamp=2024-10-27T18:42:00Z`
- Node Exporter is used to collect hardware and OS metrics from servers, exposed for Prometheus scraping. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_7bbebd9d33fe9164e019228c43b40225` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1` `source_timestamp=2024-10-27T18:42:00Z`
- Prometheus is the core metrics collection and time-series database, scraping data from sources like Node Exporter. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_7bbebd9d33fe9164e019228c43b40225` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1` `source_timestamp=2024-10-27T18:42:00Z`
- Grafana is used to visualize metrics collected by Prometheus through customizable dashboards. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_7bbebd9d33fe9164e019228c43b40225` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1` `source_timestamp=2024-10-27T18:42:00Z`
- Alertmanager handles deduplication, grouping, and routing of Prometheus alerts to destinations like email, PagerDuty, or Slack. `claim:claim_2_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_7bbebd9d33fe9164e019228c43b40225` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1` `source_timestamp=2024-10-27T18:42:00Z`
- PagerDuty is integrated as the incident management tool for handling alerts and escalations from Alertmanager. `claim:claim_2_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_7bbebd9d33fe9164e019228c43b40225` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1` `source_timestamp=2024-10-27T18:42:00Z`
- Loki is the log aggregation system used with Grafana, indexing logs by metadata rather than content for cost efficiency. `claim:claim_2_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_7bbebd9d33fe9164e019228c43b40225` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-1` `source_timestamp=2024-10-27T18:42:00Z`
- The monitoring stack deployment should be fully automated via a GitHub Actions workflow to maintain alerting rules, dashboards, historical data, role management, and PagerDuty integration after each deployment. `claim:claim_2_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-2) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_74c209b93e34d0ade8c0b2d0bc49ae46` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-2` `source_timestamp=2024-10-27T18:42:00Z`
- Reference documentation includes Prometheus Node Exporter guide, CometBFT metrics documentation, and the Story Protocol node-launcher Grafana files. `claim:claim_2_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-2) `source_document_id=srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2` `source_revision_id=srcrev_63044eb02327896049b768b6cf22894f` `chunk_id=srcchunk_74c209b93e34d0ade8c0b2d0bc49ae46` `native_locator=https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31#chunk-2` `source_timestamp=2024-10-27T18:42:00Z`

## Related Pages

- `concepts/learning-center`
- `runbooks/deploy-and-config-monitoring-stack`

## Sources

- `source_document_id`: `srcdoc_d8ad61df9040ef0963d7e30a1c7f87b2`
- `source_revision_id`: `srcrev_63044eb02327896049b768b6cf22894f`
- `source_url`: [Notion source](https://www.notion.so/Monitoring-Stack-12a051299a548046ba16d606411cbb31)
