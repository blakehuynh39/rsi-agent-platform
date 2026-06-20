---
title: "Temporal Self-Hosting Migration Plan"
type: "project"
slug: "projects/temporal-self-hosting-migration"
freshness: "2026-05-21T00:56:57Z"
tags:
  - "cost-optimization"
  - "migration"
  - "self-hosting"
  - "temporal"
owners:
  - "Aiwei"
  - "Blake"
source_revision_ids:
  - "srcrev_7bd4453ad6371a64cb792e487089d987"
  - "srcrev_cac38205da1a6ca17f6951b06d206a8c"
conflict_state: "none"
---

# Temporal Self-Hosting Migration Plan

## Summary

Proposal to migrate from Temporal Cloud to self-hosted Temporal on EKS to reduce orchestration costs, with a break-even point estimated at 30–50M Actions/month.

## Claims

- Migration to self-hosted Temporal involves deploying the Temporal server on EKS via Helm, switching 5 Temporal Cloud namespaces to internal endpoints, and reissuing mTLS certificates. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_cac38205da1a6ca17f6951b06d206a8c` `chunk_id=srcchunk_41ed086ff26680f335659f5bb7631fec` `native_locator=slack:C0547N89JUB:1779322928.231139:1779325017.786919` `source_timestamp=2026-05-21T00:56:57Z`
- Break-even for self-hosting is approximately 30–50 million Actions per month. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_cac38205da1a6ca17f6951b06d206a8c` `chunk_id=srcchunk_41ed086ff26680f335659f5bb7631fec` `native_locator=slack:C0547N89JUB:1779322928.231139:1779325017.786919` `source_timestamp=2026-05-21T00:56:57Z`
- The migration plan consists of 4 phases: deploy server, staging validation, staggered production cutover, decommission Cloud. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_cac38205da1a6ca17f6951b06d206a8c` `chunk_id=srcchunk_41ed086ff26680f335659f5bb7631fec` `native_locator=slack:C0547N89JUB:1779322928.231139:1779325017.786919` `source_timestamp=2026-05-21T00:56:57Z`
- Currently, Temporal Cloud costs are borne by the Temporal account, while the worker infrastructure is managed internally. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f053c43795827a147f78c95534d4ccc9` `source_revision_id=srcrev_7bd4453ad6371a64cb792e487089d987` `chunk_id=srcchunk_e546de99beb0c05799884b49ff2c4f4d` `native_locator=slack:C0547N89JUB:1779322928.231139:1779324096.094899` `source_timestamp=2026-05-21T00:41:36Z`

## Open Questions

- What are the infrastructure costs for running the self-hosted Temporal server?
- What is the current monthly Actions volume to confirm break-even?
- What is the timeline for each phase?
- Who will lead the migration effort?

## Related Pages

- `story-orchestration-service`

## Sources

- `source_document_id`: `srcdoc_f053c43795827a147f78c95534d4ccc9`
- `source_revision_id`: `srcrev_a0d061bafa97d39fd2bc09e63e0c91bb`
