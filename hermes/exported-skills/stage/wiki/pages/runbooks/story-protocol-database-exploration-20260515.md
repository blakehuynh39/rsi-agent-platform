---
title: "Story Protocol Database Exploration (2026-05-15)"
type: "runbook"
slug: "runbooks/story-protocol-database-exploration-20260515"
freshness: "2026-05-15T03:38:41Z"
tags:
  - "analytics"
  - "database"
  - "sql"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_03c0a0c14b843c892de69bfa9a6840a9"
  - "srcrev_0407be8724d740cb217f365f67d265ba"
  - "srcrev_09280177eef114923c0e8cc39a8960b4"
  - "srcrev_0bca0e57717273b7b547debb0601386b"
  - "srcrev_124f9555dc99b6f877ee878f182b4f93"
  - "srcrev_17a60977cb6997c3b94aca896dc9d632"
  - "srcrev_1ede3f246ed1b8682db7ce2652ded978"
  - "srcrev_27ccc3db293ff14cc2c47b7e493a8946"
  - "srcrev_2b355ccd5e2e28f9a69a3b15835c8d66"
  - "srcrev_2e5822db4394d9419fedde165e1a62a2"
  - "srcrev_3eb9f34e73e2d3a1c2418bf4b0e76074"
  - "srcrev_50345dbd1dc0510bd305d01309f9fde2"
  - "srcrev_59209477751beb471c1226a155b2e1b9"
  - "srcrev_67979ea2af05c8712ea32e4f6a997dcd"
  - "srcrev_88557bee07babdd58e2c3e9b10f7eb14"
  - "srcrev_aa4440658f5649983e84c3d19492d4e9"
  - "srcrev_b62c308782b2e8d15db7cba45ca5a4ce"
  - "srcrev_c10eff5992b4f454015e3f6a1864228c"
  - "srcrev_de82fd8293eac32ed1f8352b207999f6"
  - "srcrev_df6601fa312d779717eba38ff60f88d2"
conflict_state: "none"
---

# Story Protocol Database Exploration (2026-05-15)

## Summary

Execution of queries to gather high-level usage statistics for Story Protocol from production databases 'story-blockchain-prod' and 'sos-royalty-graph-prod' on 2026-05-15. The SQL queries and their outcomes are documented.

## Claims

- The production database 'story-blockchain-prod' contains 39 tables in the public schema, as shown by a successful query; the 'sos-royalty-graph-prod' database contains 31 tables. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_3eb9f34e73e2d3a1c2418bf4b0e76074` `chunk_id=srcchunk_d8e21e7caf9aef76c156ca7b3bf2438d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813262.966189` `source_timestamp=2026-05-15T02:48:23Z`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_1ede3f246ed1b8682db7ce2652ded978` `chunk_id=srcchunk_9351c645bae65adc6faf6fb6d6b15769` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813335.613029` `source_timestamp=2026-05-15T02:49:36Z`
- The table 'ip_registered_events' in 'story-blockchain-prod' has 12 columns (information_schema reported 12 rows). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_59209477751beb471c1226a155b2e1b9` `chunk_id=srcchunk_27bfbf13e12cbbf2779fcae37f0519d1` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814240.977489` `source_timestamp=2026-05-15T03:04:41Z`
- A total count query on 'ip_registered_events' succeeded, returning 1 row. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_27ccc3db293ff14cc2c47b7e493a8946` `chunk_id=srcchunk_70a8a76f95a9ae4fecb2491003e6941d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813629.402109` `source_timestamp=2026-05-15T03:02:26Z`
- A monthly count of 'ip_registered_events' from 2024-05-01 onward returned 7 rows. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_b62c308782b2e8d15db7cba45ca5a4ce` `chunk_id=srcchunk_cbd46de320ad09ff1600030efe7240da` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814307.883719` `source_timestamp=2026-05-15T03:20:05Z`
- A query to obtain the earliest/latest timestamp and total count of 'ip_registered_events' succeeded, returning 1 row. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_50345dbd1dc0510bd305d01309f9fde2` `chunk_id=srcchunk_e7aa03d52dfd5043c2ac460d031a1dfe` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815231.588419` `source_timestamp=2026-05-15T03:21:11Z`
- A union query over 9 event tables returned 9 rows, covering the following event types: license_terms_attached, license_template_registered, licensing_config_set, derivative_registered, royalty_vault_deployed, royalty_paid, revenue_token_claimed, metadata_uri_set, dispute_raised. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_09280177eef114923c0e8cc39a8960b4` `chunk_id=srcchunk_d3a67a7f7749d4ce7accdb6e4ea2c5f0` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814168.049689` `source_timestamp=2026-05-15T03:03:19Z`
- A query for monthly counts of 'derivative_registered_events' since 2024-05-01 returned 16 rows. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_c10eff5992b4f454015e3f6a1864228c` `chunk_id=srcchunk_17cf38acac3eb8a2591f4f94087163a5` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815760.785249` `source_timestamp=2026-05-15T03:29:32Z`
- A query for monthly counts of 'license_terms_attached_events' since 2024-05-01 returned 16 rows. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_67979ea2af05c8712ea32e4f6a997dcd` `chunk_id=srcchunk_5aaeaf14f3208a8f08ee4a79fac85061` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815793.478939` `source_timestamp=2026-05-15T03:30:29Z`
- A query for monthly counts of 'event_royalty_module_ip_royalty_vault_deployed' since 2024-05-01 returned 10 rows. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_17a60977cb6997c3b94aca896dc9d632` `chunk_id=srcchunk_8fc593c79378f2628187849f58776020` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815860.435979` `source_timestamp=2026-05-15T03:31:43Z`
- A query for monthly counts of 'event_core_metadata_module_metadata_uri_set' since 2024-05-01 returned 10 rows. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_aa4440658f5649983e84c3d19492d4e9` `chunk_id=srcchunk_a91493be5a8f765d670cb1248b2bab4a` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815924.763979` `source_timestamp=2026-05-15T03:32:46Z`
- A query for monthly counts of 'event_royalty_module_royalty_paid' since 2024-05-01 returned 10 rows. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_df6601fa312d779717eba38ff60f88d2` `chunk_id=srcchunk_b102718334a0ecf213f72f6b1c99ff98` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815987.878669` `source_timestamp=2026-05-15T03:33:23Z`
- The table 'event_ip_asset_registry_ip_registered' has 12 columns (information_schema reported 12 rows). `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_124f9555dc99b6f877ee878f182b4f93` `chunk_id=srcchunk_4ec79fd59ce3ae6160c544862fc89522` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816059.769409` `source_timestamp=2026-05-15T03:34:40Z`
- A total count query on 'event_ip_asset_registry_ip_registered' succeeded, returning 1 row. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_2e5822db4394d9419fedde165e1a62a2` `chunk_id=srcchunk_907100be67556c9ea97a805a2f17605f` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816101.701739` `source_timestamp=2026-05-15T03:35:45Z`
- In 'sos-royalty-graph-prod', a count query on 'nodes' succeeded, returning 1 row. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_de82fd8293eac32ed1f8352b207999f6` `chunk_id=srcchunk_9bb66dd041df9d45bb901b8ce2ec75ce` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815411.041569` `source_timestamp=2026-05-15T03:23:56Z`
- In 'sos-royalty-graph-prod', a count query on 'edges' succeeded, returning 1 row (stale action but final state succeeded). `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0407be8724d740cb217f365f67d265ba` `chunk_id=srcchunk_2bb4864083a271ee1396925bcf10328d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815457.800779` `source_timestamp=2026-05-15T03:26:28Z`
- In 'sos-royalty-graph-prod', a count query on 'ip_assets' succeeded, returning 1 row. `claim:claim_1_16` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_03c0a0c14b843c892de69bfa9a6840a9` `chunk_id=srcchunk_3324eef537033559f74c0a05c659c850` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815616.033609` `source_timestamp=2026-05-15T03:27:36Z`
- In 'sos-royalty-graph-prod', a count query on 'collection_aggregates' succeeded, returning 1 row. `claim:claim_1_17` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_2b355ccd5e2e28f9a69a3b15835c8d66` `chunk_id=srcchunk_185f339e34e496dbcc94883747de791c` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816172.244979` `source_timestamp=2026-05-15T03:36:43Z`
- In 'sos-royalty-graph-prod', a union query to count 'ip_licenses', 'ip_transactions', and 'royalty_token_transfers' succeeded, returning 3 rows. `claim:claim_1_18` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_88557bee07babdd58e2c3e9b10f7eb14` `chunk_id=srcchunk_fb3e7116479bf2342a34c7a24e03366b` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816224.376759` `source_timestamp=2026-05-15T03:38:41Z`
- An aggregate counts query on multiple metrics (nodes, edges, ip_assets, ip_licenses, ip_transactions, etc.) from 'sos-royalty-graph-prod' timed out with 'context deadline exceeded'. `claim:claim_1_19` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0bca0e57717273b7b547debb0601386b` `chunk_id=srcchunk_4a8a9eb260afcb83db1bad7a3c9ad1c8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813407.495359` `source_timestamp=2026-05-15T02:53:18Z`

## Open Questions

- What were the actual values returned by the successful count queries (e.g., total ip registrations, monthly breakdowns)?
- Why did the aggregate union query on sos-royalty-graph-prod time out, and can it be optimized?

## Sources

- `source_document_id`: `srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2`
- `source_revision_id`: `srcrev_88557bee07babdd58e2c3e9b10f7eb14`
