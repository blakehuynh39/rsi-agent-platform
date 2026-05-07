---
title: "v2 Monitoring and CI/CD Design"
type: "decision"
slug: "decisions/v2-monitoring-cicd-design"
freshness: "2023-06-20T10:08:00Z"
tags:
  - "cicd"
  - "infrastructure"
  - "monitoring"
  - "observability"
owners: []
source_revision_ids:
  - "srcrev_498642f02f08a2c2eecbfafb4947df37"
conflict_state: "none"
---

# v2 Monitoring and CI/CD Design

## Summary

Design for the v2 monitoring and CI/CD system to address limitations of the v1 setup, introducing centralized observability with Thanos, multi-cluster ArgoCD, consolidated infrastructure-as-code, and automated promotions.

## Claims

- v1 monitoring used Grafana Cloud with kube-state-metrics and node_exporter, replacing deprecated AWS Managed Grafana and Prometheus due to high cost. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-1) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_bfa9f9949b2ef4eac35024dcc6549000` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-1` `source_timestamp=2023-06-20T10:08:00Z`
- v1 CI/CD used GitHub Actions for CI, a single-cluster ArgoCD instance for CD, and manual promotion processes. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-1) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_bfa9f9949b2ef4eac35024dcc6549000` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-1` `source_timestamp=2023-06-20T10:08:00Z`
- v1 limitations included mixing operations and application logic in the same cluster, no global monitoring view across environments, no multi-cluster CI/CD, and no streamlined promotion process. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-1) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_bfa9f9949b2ef4eac35024dcc6549000` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-1` `source_timestamp=2023-06-20T10:08:00Z`
- v2 will introduce Thanos, using remote-write from per-environment Prometheus instances to a central sharded Thanos querier for lower-latency, centralized metrics. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_74e7765488b0b5d6baf5ce4556dbff41` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2` `source_timestamp=2023-06-20T10:08:00Z`
- v2 will switch from per-environment ArgoCD instances to a central ArgoCD with per-environment workers, enabling management from a single global interface. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_74e7765488b0b5d6baf5ce4556dbff41` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2` `source_timestamp=2023-06-20T10:08:00Z`
- v2 will consolidate all infrastructure and application state code into a single iac repository with per-environment folder separation. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_74e7765488b0b5d6baf5ce4556dbff41` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2` `source_timestamp=2023-06-20T10:08:00Z`
- v2 will leverage ArgoCD post-sync hooks to automate CI/CD promotion, modeling streamlined approval-based promotions. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_74e7765488b0b5d6baf5ce4556dbff41` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2` `source_timestamp=2023-06-20T10:08:00Z`
- v2 will create an operations management cluster in the cicd account to separate application development from monitoring and deployments. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_74e7765488b0b5d6baf5ce4556dbff41` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2` `source_timestamp=2023-06-20T10:08:00Z`
- Completion of the v2 system will deprecate several existing documentation pages. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2) `source_document_id=srcdoc_fa13b656f0b6caceff0132ffc078790b` `source_revision_id=srcrev_498642f02f08a2c2eecbfafb4947df37` `chunk_id=srcchunk_74e7765488b0b5d6baf5ce4556dbff41` `native_locator=https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5#chunk-2` `source_timestamp=2023-06-20T10:08:00Z`

## Open Questions

- Migration plan for the v2 system is yet to be determined.

## Sources

- `source_document_id`: `srcdoc_fa13b656f0b6caceff0132ffc078790b`
- `source_revision_id`: `srcrev_498642f02f08a2c2eecbfafb4947df37`
- `source_url`: [Notion source](https://www.notion.so/Design-Doc-v2-Monitoring-CICD-9a34825c70dd436ebd764cb2924480c5)
