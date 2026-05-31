---
title: "Productionize the Backend"
type: "project"
slug: "projects/productionize-backend"
freshness: "2024-09-25T23:26:00Z"
tags:
  - "backend"
  - "cicd"
  - "monitoring"
  - "production"
owners: []
source_revision_ids:
  - "srcrev_035ca9fb17a6989fd28f75f811bd267c"
conflict_state: "none"
---

# Productionize the Backend

## Summary

Plan to productionize the backend service including monitoring, alerting, CI/CD pipeline improvements, and addressing live update issues.

## Claims

- Production cluster setup includes adding Prometheus and Grafana dashboard, Sentry, testing admin API keys for protocol APIs, testing dynamics integration, updating API Gateway, performing DB migration, and updating CI/CD flow. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238) `source_document_id=srcdoc_901d669c235fb6624bb6247966a14958` `source_revision_id=srcrev_035ca9fb17a6989fd28f75f811bd267c` `chunk_id=srcchunk_4e290dfac1d94392d1ada9095860fecc` `native_locator=https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238` `source_timestamp=2024-09-25T23:26:00Z`
- Staging cluster setup includes hooking Sentry and Grafana to PagerDuty (especially for Hub APIs), monitoring error rate and latency in Grafana, configuring API Gateway, and implementing CI/CD. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238) `source_document_id=srcdoc_901d669c235fb6624bb6247966a14958` `source_revision_id=srcrev_035ca9fb17a6989fd28f75f811bd267c` `chunk_id=srcchunk_4e290dfac1d94392d1ada9095860fecc` `native_locator=https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238` `source_timestamp=2024-09-25T23:26:00Z`
- Current staging CI/CD: PR against staging branch triggers test, lint, and deploy; merging to staging triggers deployment. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238) `source_document_id=srcdoc_901d669c235fb6624bb6247966a14958` `source_revision_id=srcrev_035ca9fb17a6989fd28f75f811bd267c` `chunk_id=srcchunk_4e290dfac1d94392d1ada9095860fecc` `native_locator=https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238` `source_timestamp=2024-09-25T23:26:00Z`
- Correct CI/CD process: PR to staging branch runs lint and test; merge to staging triggers deploy to staging; PR to main runs test again; merge to main triggers deploy to production. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238) `source_document_id=srcdoc_901d669c235fb6624bb6247966a14958` `source_revision_id=srcrev_035ca9fb17a6989fd28f75f811bd267c` `chunk_id=srcchunk_4e290dfac1d94392d1ada9095860fecc` `native_locator=https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238` `source_timestamp=2024-09-25T23:26:00Z`
- Current problems: live backend code updates break the app; Zettablock lacks staging and production environments, causing app breakage on live updates; need to discuss with Zettablock to avoid live updates. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238) `source_document_id=srcdoc_901d669c235fb6624bb6247966a14958` `source_revision_id=srcrev_035ca9fb17a6989fd28f75f811bd267c` `chunk_id=srcchunk_4e290dfac1d94392d1ada9095860fecc` `native_locator=https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238` `source_timestamp=2024-09-25T23:26:00Z`

## Sources

- `source_document_id`: `srcdoc_901d669c235fb6624bb6247966a14958`
- `source_revision_id`: `srcrev_035ca9fb17a6989fd28f75f811bd267c`
- `source_url`: [Notion source](https://www.notion.so/Productionize-the-backend-10c051299a5480b1af6ade7d4a668238)
