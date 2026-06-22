---
title: "Protocol Usage Data Exploration (2026-05-15)"
type: "runbook"
slug: "runbooks/protocol-usage-data-exploration-2026-05-15"
freshness: "2026-05-15T03:33:23Z"
tags:
  - "data-exploration"
  - "events"
  - "ip-graph"
  - "protocol-usage"
  - "royalty"
owners:
  - "@U0772SH7BRA"
source_revision_ids:
  - "srcrev_0407be8724d740cb217f365f67d265ba"
  - "srcrev_09280177eef114923c0e8cc39a8960b4"
  - "srcrev_0bca0e57717273b7b547debb0601386b"
  - "srcrev_17a60977cb6997c3b94aca896dc9d632"
  - "srcrev_1ede3f246ed1b8682db7ce2652ded978"
  - "srcrev_27ccc3db293ff14cc2c47b7e493a8946"
  - "srcrev_3eb9f34e73e2d3a1c2418bf4b0e76074"
  - "srcrev_59209477751beb471c1226a155b2e1b9"
  - "srcrev_67979ea2af05c8712ea32e4f6a997dcd"
  - "srcrev_aa4440658f5649983e84c3d19492d4e9"
  - "srcrev_b62c308782b2e8d15db7cba45ca5a4ce"
  - "srcrev_bab416db7564b1a0a9bec7c85fde1838"
  - "srcrev_c10eff5992b4f454015e3f6a1864228c"
  - "srcrev_df6601fa312d779717eba38ff60f88d2"
  - "srcrev_f7c168547d4afd8e2051a619863029cf"
conflict_state: "none"
---

# Protocol Usage Data Exploration (2026-05-15)

## Summary

On 2026-05-15, an RSI task was initiated to produce high-level usage stats for the Story Protocol over the last 2 years, using data from sos-royalty-graph-prod and story-blockchain-prod. Multiple SQL queries were executed to explore schema and gather event counts. Results were partially captured; actual count values were not included in the chat log.

## Claims

- The analysis was requested to generate a general graph/plot showing protocol usage in the last 2 years using data from sos prod and blockchain prod. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_f7c168547d4afd8e2051a619863029cf` `chunk_id=srcchunk_fd5e8f9cf1133a7b6f59e672f000b695` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813134.988149` `source_timestamp=2026-05-15T02:45:34Z`
- The RSI run was tracked under trace trace-629642f551e84133ab24c3706a6caa4f. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_bab416db7564b1a0a9bec7c85fde1838` `chunk_id=srcchunk_ac53e5357b01272ed2d922a9a25214a8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813139.452969` `source_timestamp=2026-05-15T02:45:39Z`
- The sos-royalty-graph-prod database has 31 tables in its public schema. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_3eb9f34e73e2d3a1c2418bf4b0e76074` `chunk_id=srcchunk_d8e21e7caf9aef76c156ca7b3bf2438d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813262.966189` `source_timestamp=2026-05-15T02:48:23Z`
- The story-blockchain-prod database has 39 tables in its public schema. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_1ede3f246ed1b8682db7ce2652ded978` `chunk_id=srcchunk_9351c645bae65adc6faf6fb6d6b15769` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813335.613029` `source_timestamp=2026-05-15T02:49:36Z`
- A combined count query (nodes, edges, ip_assets, etc.) on sos-royalty-graph-prod timed out. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0bca0e57717273b7b547debb0601386b` `chunk_id=srcchunk_4a8a9eb260afcb83db1bad7a3c9ad1c8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813407.495359` `source_timestamp=2026-05-15T02:53:18Z`
- A total IP registrations count query (`SELECT count(*) FROM ip_registered_events`) on story-blockchain-prod completed successfully. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_27ccc3db293ff14cc2c47b7e493a8946` `chunk_id=srcchunk_70a8a76f95a9ae4fecb2491003e6941d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813629.402109` `source_timestamp=2026-05-15T03:02:26Z`
- A query returned counts for 9 distinct event types on story-blockchain-prod, including license_terms_attached, derivative_registered, royalty_vault_deployed, royalty_paid, metadata_uri_set, dispute_raised, etc. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_09280177eef114923c0e8cc39a8960b4` `chunk_id=srcchunk_d3a67a7f7749d4ce7accdb6e4ea2c5f0` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814168.049689` `source_timestamp=2026-05-15T03:03:19Z`
- The ip_registered_events table in story-blockchain-prod has 12 columns. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_59209477751beb471c1226a155b2e1b9` `chunk_id=srcchunk_27bfbf13e12cbbf2779fcae37f0519d1` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814240.977489` `source_timestamp=2026-05-15T03:04:41Z`
- Monthly IP registration counts were queried from May 2024 onwards, returning 7 rows (months). `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_b62c308782b2e8d15db7cba45ca5a4ce` `chunk_id=srcchunk_cbd46de320ad09ff1600030efe7240da` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814307.883719` `source_timestamp=2026-05-15T03:20:05Z`
- A request to count edges on sos-royalty-graph-prod was ignored because a previous count action was already in a succeeded state. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0407be8724d740cb217f365f67d265ba` `chunk_id=srcchunk_2bb4864083a271ee1396925bcf10328d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815457.800779` `source_timestamp=2026-05-15T03:26:28Z`
- Monthly derivative registration counts from May 2024 yielded 16 months of data. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_c10eff5992b4f454015e3f6a1864228c` `chunk_id=srcchunk_17cf38acac3eb8a2591f4f94087163a5` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815760.785249` `source_timestamp=2026-05-15T03:29:32Z`
- Monthly license terms attached counts from May 2024 also yielded 16 months of data. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_67979ea2af05c8712ea32e4f6a997dcd` `chunk_id=srcchunk_5aaeaf14f3208a8f08ee4a79fac85061` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815793.478939` `source_timestamp=2026-05-15T03:30:29Z`
- Monthly royalty vault deployments from May 2024 resulted in 10 rows. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_17a60977cb6997c3b94aca896dc9d632` `chunk_id=srcchunk_8fc593c79378f2628187849f58776020` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815860.435979` `source_timestamp=2026-05-15T03:31:43Z`
- Monthly metadata URI set events from May 2024 also returned 10 rows. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_aa4440658f5649983e84c3d19492d4e9` `chunk_id=srcchunk_a91493be5a8f765d670cb1248b2bab4a` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815924.763979` `source_timestamp=2026-05-15T03:32:46Z`
- Monthly royalty paid events from May 2024 returned 10 rows. `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_df6601fa312d779717eba38ff60f88d2` `chunk_id=srcchunk_b102718334a0ecf213f72f6b1c99ff98` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815987.878669` `source_timestamp=2026-05-15T03:33:23Z`

## Open Questions

- What were the actual count values for the various metrics? The query results were truncated in the chat and did not include the numeric totals.

## Sources

- `source_document_id`: `srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2`
- `source_revision_id`: `srcrev_2e5822db4394d9419fedde165e1a62a2`
