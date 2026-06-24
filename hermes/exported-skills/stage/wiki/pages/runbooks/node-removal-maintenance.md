---
title: "Node Removal Maintenance"
type: "runbook"
slug: "runbooks/node-removal-maintenance"
freshness: "2026-03-16T05:25:13Z"
tags:
  - "cloudsql"
  - "gke"
  - "maintenance"
  - "production"
  - "stage"
owners: []
source_revision_ids:
  - "srcrev_e84b125314054d70fbb2222813008446"
conflict_state: "none"
---

# Node Removal Maintenance

## Summary

Procedure and communication for removing nodes from Stage/Production GKE and CloudSQL.

## Claims

- During maintenance, all nodes in Stage and Production GKE clusters and CloudSQL are removed, with a final backup for CloudSQL. Users should report any issues. `claim:claim_6_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_e84b125314054d70fbb2222813008446` `chunk_id=srcchunk_ba785ddcd6bb095d3e7d105e0fdf5e38` `native_locator=slack:C0547N89JUB:1773638713.701879:1773638713.701879` `source_timestamp=2026-03-16T05:25:13Z`

## Sources

- `source_document_id`: `srcdoc_b853172082170c2d6c8b40804fa03731`
- `source_revision_id`: `srcrev_e108af2fe4f8cd79a31c55c17dca3e86`
