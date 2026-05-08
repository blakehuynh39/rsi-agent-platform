---
title: "L1 Infra Roadmap (Sep 30)"
type: "project"
slug: "projects/l1-infra-roadmap-sep-30"
freshness: "2024-09-30T18:32:00Z"
tags:
  - "infrastructure"
  - "l1"
  - "roadmap"
owners: []
source_revision_ids:
  - "srcrev_5b351a74a66db67d6024bfa66201f3da"
  - "srcrev_7bd58ea4cb4f63271f13352213088fd4"
conflict_state: "none"
---

# L1 Infra Roadmap (Sep 30)

## Summary

High-level roadmap for L1 infrastructure covering cost control, network automation, development/testing/release infrastructure, and incident response/recovery. Also includes detailed phases from an earlier Infra Roadmap (June-August 2023) covering IAM, IaC, monitoring, CI/CD, secrets management, network optimization, and more.

## Claims

- Core Infra Cost control includes network config automation. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Infra-roadmap-Sep-30-111051299a5480f2a022deea2b1b0a34) `source_document_id=srcdoc_8e8fba82c4567ce6b2ccc55582b07919` `source_revision_id=srcrev_7bd58ea4cb4f63271f13352213088fd4` `chunk_id=srcchunk_8a7b842e2663b6e31002bf7bbdf537c5` `native_locator=https://www.notion.so/L1-Infra-roadmap-Sep-30-111051299a5480f2a022deea2b1b0a34` `source_timestamp=2024-09-30T18:32:00Z`
- Development, Testing and Release infra covers local test and CI test. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Infra-roadmap-Sep-30-111051299a5480f2a022deea2b1b0a34) `source_document_id=srcdoc_8e8fba82c4567ce6b2ccc55582b07919` `source_revision_id=srcrev_7bd58ea4cb4f63271f13352213088fd4` `chunk_id=srcchunk_8a7b842e2663b6e31002bf7bbdf537c5` `native_locator=https://www.notion.so/L1-Infra-roadmap-Sep-30-111051299a5480f2a022deea2b1b0a34` `source_timestamp=2024-09-30T18:32:00Z`
- Incident response/recovery includes observability (monitoring), alerting, and recovery process. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Infra-roadmap-Sep-30-111051299a5480f2a022deea2b1b0a34) `source_document_id=srcdoc_8e8fba82c4567ce6b2ccc55582b07919` `source_revision_id=srcrev_7bd58ea4cb4f63271f13352213088fd4` `chunk_id=srcchunk_8a7b842e2663b6e31002bf7bbdf537c5` `native_locator=https://www.notion.so/L1-Infra-roadmap-Sep-30-111051299a5480f2a022deea2b1b0a34` `source_timestamp=2024-09-30T18:32:00Z`
- Phase 1 (Initial Scoping) runs from June 14 to June 28, 2023, and covers IAM, IaC, and Monitoring. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4#chunk-1) `source_document_id=srcdoc_bb27066ab6ea663e91586cae1ccb9b7f` `source_revision_id=srcrev_5b351a74a66db67d6024bfa66201f3da` `chunk_id=srcchunk_424c1d20b0fc909ecd3d1548e5ec28cb` `native_locator=https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4#chunk-1` `source_timestamp=2023-07-10T07:31:00Z`
- Phase 1 goals include creating operations documentation for account management, reviewing all IaC code, and setting up Grafana for staging and prod environments. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4#chunk-1) `source_document_id=srcdoc_bb27066ab6ea663e91586cae1ccb9b7f` `source_revision_id=srcrev_5b351a74a66db67d6024bfa66201f3da` `chunk_id=srcchunk_424c1d20b0fc909ecd3d1548e5ec28cb` `native_locator=https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4#chunk-1` `source_timestamp=2023-07-10T07:31:00Z`
- Phase 4 (Network Optimization + WF Integration) runs from July 21 to August 14, 2023, and covers API Infrastructure, Network Optimization, Data Storage/Caching, and Smart Contract/Protocol. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4#chunk-2) `source_document_id=srcdoc_bb27066ab6ea663e91586cae1ccb9b7f` `source_revision_id=srcrev_5b351a74a66db67d6024bfa66201f3da` `chunk_id=srcchunk_4143eeb4809aa1402436ecb68beb9a44` `native_locator=https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4#chunk-2` `source_timestamp=2023-07-10T07:31:00Z`
- Phase 4 goals include ensuring APIs are defined with OpenAPI and CI, creating a v0 spec for istio or alternative ingress controllers, and rolling out FE deployments from Vercel to AWS as a PoC. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4#chunk-2) `source_document_id=srcdoc_bb27066ab6ea663e91586cae1ccb9b7f` `source_revision_id=srcrev_5b351a74a66db67d6024bfa66201f3da` `chunk_id=srcchunk_4143eeb4809aa1402436ecb68beb9a44` `native_locator=https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4#chunk-2` `source_timestamp=2023-07-10T07:31:00Z`

## Sources

- `source_document_id`: `srcdoc_bb27066ab6ea663e91586cae1ccb9b7f`
- `source_revision_id`: `srcrev_5b351a74a66db67d6024bfa66201f3da`
- `source_url`: [Notion source](https://www.notion.so/Infra-Roadmap-f7a432a059be4b57a4dd4607e1d23ce4)
