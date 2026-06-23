---
title: "Node Removal from GKE Clusters and CloudSQL"
type: "runbook"
slug: "runbooks/node-removal-gke-cloudsql"
freshness: "2026-03-16T05:25:13Z"
tags:
  - "cloudsql"
  - "gke"
  - "maintenance"
  - "node-removal"
owners: []
source_revision_ids:
  - "srcrev_e84b125314054d70fbb2222813008446"
conflict_state: "none"
---

# Node Removal from GKE Clusters and CloudSQL

## Summary

Removal of all nodes from Stage and Production GKE clusters and CloudSQL, with a final backup for CloudSQL. Users should report any issues.

## Claims

- All nodes in Stage and Production GKE clusters and CloudSQL are being removed, with a final backup to be performed for CloudSQL. `claim:claim_7_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b853172082170c2d6c8b40804fa03731` `source_revision_id=srcrev_e84b125314054d70fbb2222813008446` `chunk_id=srcchunk_ba785ddcd6bb095d3e7d105e0fdf5e38` `native_locator=slack:C0547N89JUB:1773638713.701879:1773638713.701879` `source_timestamp=2026-03-16T05:25:13Z`

## Open Questions

- What impact will this have on services?
- What is the timeline for node removal and restoration?

## Sources

- `source_document_id`: `srcdoc_b853172082170c2d6c8b40804fa03731`
- `source_revision_id`: `srcrev_97c339aa56c30de445f0c4ee3c0d58ce`
