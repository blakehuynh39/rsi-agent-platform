---
title: "Aeneid Transaction Failure Incident (2026-06-17)"
type: "runbook"
slug: "runbooks/aeneid-tx-failure-incident-2026-06-17"
freshness: "2026-06-16T20:36:49Z"
tags:
  - "aeneid"
  - "incident"
  - "rpc"
  - "testnet"
  - "transactions"
owners:
  - "Aeneid support team"
source_revision_ids:
  - "srcrev_749d2a33a2788f285a187d1fba4d89be"
  - "srcrev_89739c1edcbe207d06795619a25fc35c"
  - "srcrev_8eeeed44ef072b290e72adf1c99fd514"
  - "srcrev_90804202273551e62a6c9fb6ed96223e"
  - "srcrev_9ed082ac464a237429aedcad0b9a58b8"
  - "srcrev_d47e47dd8011d3a05bb3665ab9af13b6"
conflict_state: "none"
---

# Aeneid Transaction Failure Incident (2026-06-17)

## Summary

On 2026-06-17, users reported Aeneid testnet transactions failing despite the chain producing blocks. Investigation showed public RPC responsive, blocks advancing, but 0 txs included and 500 pending. The issue was related to gas fee specification and was being fixed. A transaction succeeded when gas was manually set.

## Claims

- Users reported Aeneid testnet being down and unable to execute transactions. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1c63b2cd8fccb7955787c064755232b1` `source_revision_id=srcrev_d47e47dd8011d3a05bb3665ab9af13b6` `chunk_id=srcchunk_d6d6695872fe8b79e5fbd9c0ad3503a8` `native_locator=slack:C0547N89JUB:1781636003.409529:1781636003.409529` `source_timestamp=2026-06-16T18:53:23Z`
- Public RPC endpoint was responding and blocks were advancing normally. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1c63b2cd8fccb7955787c064755232b1` `source_revision_id=srcrev_749d2a33a2788f285a187d1fba4d89be` `chunk_id=srcchunk_c1b97fa28297b2e11f1315c54b63f07f` `native_locator=slack:C0547N89JUB:1781636003.409529:1781636093.700839` `source_timestamp=2026-06-16T18:54:53Z`
- Transactions were not being included in blocks; 500 pending transactions observed. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1c63b2cd8fccb7955787c064755232b1` `source_revision_id=srcrev_9ed082ac464a237429aedcad0b9a58b8` `chunk_id=srcchunk_272b6df2bede3f3da0133389420b3350` `native_locator=slack:C0547N89JUB:1781636003.409529:1781640715.682809` `source_timestamp=2026-06-16T20:11:55Z`
- Active Aeneid alerts included DKG/partials and a suppressed snapshot-archive peer alert, but these did not indicate a chain halt. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1c63b2cd8fccb7955787c064755232b1` `source_revision_id=srcrev_749d2a33a2788f285a187d1fba4d89be` `chunk_id=srcchunk_c1b97fa28297b2e11f1315c54b63f07f` `native_locator=slack:C0547N89JUB:1781636003.409529:1781636093.700839` `source_timestamp=2026-06-16T18:54:53Z`
- A test transaction succeeded when gas fees were manually specified. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1c63b2cd8fccb7955787c064755232b1` `source_revision_id=srcrev_8eeeed44ef072b290e72adf1c99fd514` `chunk_id=srcchunk_cbda82b852d4e5f3ed9159d4c1a1d985` `native_locator=slack:C0547N89JUB:1781636003.409529:1781642209.447519` `source_timestamp=2026-06-16T20:36:49Z`
- The issue was being addressed and expected to be fixed in several hours. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1c63b2cd8fccb7955787c064755232b1` `source_revision_id=srcrev_90804202273551e62a6c9fb6ed96223e` `chunk_id=srcchunk_e7cb27c49f630d46b557a62c26059ba2` `native_locator=slack:C0547N89JUB:1781636003.409529:1781636107.601459` `source_timestamp=2026-06-16T18:55:46Z`
- The Grafana dashboard for aeneid.storyrpc.io showed no data. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1c63b2cd8fccb7955787c064755232b1` `source_revision_id=srcrev_89739c1edcbe207d06795619a25fc35c` `chunk_id=srcchunk_cf21b13e08683a0499829d2fb759391f` `native_locator=slack:C0547N89JUB:1781636003.409529:1781641787.885799` `source_timestamp=2026-06-16T20:29:47Z`

## Open Questions

- What was the root cause of transaction inclusion failure?
- Why did specifying gas fees manually allow transactions to succeed?
- Why was the aeneid.storyrpc.io Grafana dashboard missing data during this time?

## Sources

- `source_document_id`: `srcdoc_1c63b2cd8fccb7955787c064755232b1`
- `source_revision_id`: `srcrev_3693878dfe3f2127d3c13ccb2b7a561f`
