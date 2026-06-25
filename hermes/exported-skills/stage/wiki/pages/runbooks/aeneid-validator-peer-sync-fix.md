---
title: "Aeneid Validator Peer Sync Issue Resolution"
type: "runbook"
slug: "runbooks/aeneid-validator-peer-sync-fix"
freshness: "2026-01-29T11:23:20Z"
tags:
  - "aeneid"
  - "peer"
  - "sync"
  - "troubleshooting"
  - "validator"
owners: []
source_revision_ids:
  - "srcrev_a7455a72581ef09e8364b80a5b00178d"
  - "srcrev_f60559c7cd8d8e7fe373593b9bb57525"
  - "srcrev_f7c55148aacfece9f1c814514d37e881"
conflict_state: "none"
---

# Aeneid Validator Peer Sync Issue Resolution

## Summary

Resolve sync errors in Aeneid validator nodes where geth logs 'Number of finalized block is missing' by manually adding peers.

## Claims

- Geth error: 'Number of finalized block is missing' and CL logs show 'execution engine is syncing' warnings. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f60559c7cd8d8e7fe373593b9bb57525` `chunk_id=srcchunk_d211a0c443638f4434efda509e900c54` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678206.437309` `source_timestamp=2026-01-29T09:16:46Z`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_f7c55148aacfece9f1c814514d37e881` `chunk_id=srcchunk_794a1f7289b9b9c443ac9fcff3645248` `native_locator=slack:C0547N89JUB:1769674617.908059:1769678252.258289` `source_timestamp=2026-01-29T09:17:32Z`
- The root cause was a peer issue, and the problem was resolved by manually adding a peer. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d` `source_revision_id=srcrev_a7455a72581ef09e8364b80a5b00178d` `chunk_id=srcchunk_0c93db5007818e1eae81bb9ad758c83d` `native_locator=slack:C0547N89JUB:1769674617.908059:1769685706.820279` `source_timestamp=2026-01-29T11:23:20Z`

## Related Pages

- `aeneid-rpc-validator-migration`

## Sources

- `source_document_id`: `srcdoc_b27338e7e24e5f35c74b2f2a9cfd8e4d`
- `source_revision_id`: `srcrev_faefaf0cae75e550b0ef2d719e9b2d0e`
