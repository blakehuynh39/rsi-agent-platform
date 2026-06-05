---
title: "Centralized Exchange Migration Steps"
type: "runbook"
slug: "runbooks/centralized-exchange-data-migration-runbook"
freshness: "2026-06-05T21:39:00Z"
tags:
  - "exchange"
  - "rpc"
  - "ticker"
owners: []
source_revision_ids:
  - "srcrev_387e1c774ce3439b600a5ab109f7d463"
conflict_state: "none"
---

# Centralized Exchange Migration Steps

## Summary

Centralized exchanges should update the token ticker to $DATA, migrate RPC endpoints from storyrpc.io to datanetworkrpc.io, and note that chain ID remains unchanged.

## Claims

- No action regarding chain id is needed; it remains 1514. `claim:claim_3_1` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`
- If you don’t run your own node, the RPC domain will change from mainnet.storyrpc.io and aeneid.storyrpc.io to mainnet.datanetworkrpc.io (and presumably aeneid.datanetworkrpc.io). `claim:claim_3_2` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`
- Exchanges must update ticker to $DATA. `claim:claim_3_3` `confidence:1.00`
  - citation: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893) `source_document_id=srcdoc_595c1fb031be8bea2eb6952240911f03` `source_revision_id=srcrev_387e1c774ce3439b600a5ab109f7d463` `chunk_id=srcchunk_e557c332ffa34b62326d62ed346d958f` `native_locator=https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893` `source_timestamp=2026-06-05T21:39:00Z`

## Sources

- `source_document_id`: `srcdoc_595c1fb031be8bea2eb6952240911f03`
- `source_revision_id`: `srcrev_387e1c774ce3439b600a5ab109f7d463`
- `source_url`: [source](https://app.notion.com/p/DATA-Migration-technical-handbook-external-375051299a54801090a8ca9ab924f893)
