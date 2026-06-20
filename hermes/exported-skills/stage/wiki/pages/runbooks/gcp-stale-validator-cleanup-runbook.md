---
title: "GCP Stale Validator Resource Cleanup Runbook"
type: "runbook"
slug: "runbooks/gcp-stale-validator-cleanup-runbook"
freshness: "2026-05-20T22:13:28Z"
tags:
  - "cleanup"
  - "cost-optimization"
  - "gcp"
  - "validators"
owners: []
source_revision_ids:
  - "srcrev_61cbdbd12af99b7b4f8b754a512d7cc4"
  - "srcrev_626499a076b2cc149a8591f63df71fb5"
  - "srcrev_9652917dd94f0303bc628f408613641f"
  - "srcrev_d10e6d54449750d1dd6e62a1ba0df0e1"
  - "srcrev_d281d41c291cb0d07ef1e128971a86ee"
  - "srcrev_fc476854d698c98308d2cb60c11785d5"
conflict_state: "none"
---

# GCP Stale Validator Resource Cleanup Runbook

## Summary

Procedure to clean up stale GCP resources (SSD volumes and snapshots) from employee-managed validator nodes that have been moved to AWS, saving approximately $2k per month.

## Claims

- The GCP bill for April and May 2026 showed stale resources that should be cleaned up. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_704b9dd469c1e7597314b22b4b3dbb8a` `source_revision_id=srcrev_d10e6d54449750d1dd6e62a1ba0df0e1` `chunk_id=srcchunk_36e5e43df60a60e137b5fecce77abcb5` `native_locator=slack:C0547N89JUB:1779297888.535289:1779297888.535289` `source_timestamp=2026-05-20T17:24:48Z`
- Mainnet employee-managed validator projects have SSD volumes and snapshots costing approximately $2,000 per month combined. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_704b9dd469c1e7597314b22b4b3dbb8a` `source_revision_id=srcrev_9652917dd94f0303bc628f408613641f` `chunk_id=srcchunk_5d296ca6e7560f439e3f7b5d832bc060` `native_locator=slack:C0547N89JUB:1779297888.535289:1779297949.909969` `source_timestamp=2026-05-20T17:25:49Z`
- Before deleting any resources, verify they are not in the active validator set. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_704b9dd469c1e7597314b22b4b3dbb8a` `source_revision_id=srcrev_fc476854d698c98308d2cb60c11785d5` `chunk_id=srcchunk_a467e1ec1ee09b8b8abe6b6b43cff8d4` `native_locator=slack:C0547N89JUB:1779297888.535289:1779311309.561889` `source_timestamp=2026-05-20T21:08:29Z`
- The nodes have already been moved to AWS; only the large storage volumes and snapshots remain on GCP, causing significant cost. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_704b9dd469c1e7597314b22b4b3dbb8a` `source_revision_id=srcrev_61cbdbd12af99b7b4f8b754a512d7cc4` `chunk_id=srcchunk_cc57292a54e4b818f1df281486a06497` `native_locator=slack:C0547N89JUB:1779297888.535289:1779311432.611929` `source_timestamp=2026-05-20T21:10:32Z`
- Stale GCP resources from employee-managed nodes (not active validators) should be deleted to reduce costs. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_704b9dd469c1e7597314b22b4b3dbb8a` `source_revision_id=srcrev_626499a076b2cc149a8591f63df71fb5` `chunk_id=srcchunk_596f4841e06f3eab2e9da64f22e316b4` `native_locator=slack:C0547N89JUB:1779297888.535289:1779312433.044429` `source_timestamp=2026-05-20T21:27:13Z`
- The cost forecast for a spike date was $10k; some IPs, stopped servers, and snapshots were kept as a precaution but are planned for removal as of 2026-05-20. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_704b9dd469c1e7597314b22b4b3dbb8a` `source_revision_id=srcrev_d281d41c291cb0d07ef1e128971a86ee` `chunk_id=srcchunk_b218a9046af20a815ca29e32b78f043c` `native_locator=slack:C0547N89JUB:1779297888.535289:1779315208.501129` `source_timestamp=2026-05-20T22:13:28Z`

## Sources

- `source_document_id`: `srcdoc_704b9dd469c1e7597314b22b4b3dbb8a`
- `source_revision_id`: `srcrev_d281d41c291cb0d07ef1e128971a86ee`
