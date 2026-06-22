---
title: "Story Protocol Usage Query Runbook"
type: "runbook"
slug: "runbooks/story-protocol-usage-queries"
freshness: "2026-05-15T03:38:41Z"
tags:
  - "database"
  - "monitoring"
  - "query"
  - "sql"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_03c0a0c14b843c892de69bfa9a6840a9"
  - "srcrev_0407be8724d740cb217f365f67d265ba"
  - "srcrev_09280177eef114923c0e8cc39a8960b4"
  - "srcrev_0bca0e57717273b7b547debb0601386b"
  - "srcrev_124f9555dc99b6f877ee878f182b4f93"
  - "srcrev_15ee32dc8a52b1280c4402a07c56df81"
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
  - "srcrev_e5947999833d240dd58fb13540a86cd5"
  - "srcrev_f7c168547d4afd8e2051a619863029cf"
conflict_state: "none"
---

# Story Protocol Usage Query Runbook

## Summary

Catalog of SQL queries executed against production databases (sos-royalty-graph-prod and story-blockchain-prod) to gather high-level protocol activity metrics, as performed in May 2026.

## Claims

- A request was made to analyze IP graph/proof of creativity indexing data from SOS and blockchain prod for the last 2 years to produce general usage stats and plots. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_f7c168547d4afd8e2051a619863029cf` `chunk_id=srcchunk_fd5e8f9cf1133a7b6f59e672f000b695` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813134.988149` `source_timestamp=2026-05-15T02:45:34Z`
- The query `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name` was executed on `sos-royalty-graph-prod` and returned 31 rows. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_3eb9f34e73e2d3a1c2418bf4b0e76074` `chunk_id=srcchunk_d8e21e7caf9aef76c156ca7b3bf2438d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813262.966189` `source_timestamp=2026-05-15T02:48:23Z`
- The query `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name` was executed on `story-blockchain-prod` and returned 39 rows. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_1ede3f246ed1b8682db7ce2652ded978` `chunk_id=srcchunk_9351c645bae65adc6faf6fb6d6b15769` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813335.613029` `source_timestamp=2026-05-15T02:49:36Z`
- A query attempting to union counts from multiple SOS royalty graph tables (`nodes`, `edges`, `ip_assets`, etc.) on `sos-royalty-graph-prod` failed due to timeout. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0bca0e57717273b7b547debb0601386b` `chunk_id=srcchunk_4a8a9eb260afcb83db1bad7a3c9ad1c8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813407.495359` `source_timestamp=2026-05-15T02:53:18Z`
- The query `SELECT count(*) as total_ip_registrations FROM ip_registered_events` was executed on `story-blockchain-prod` and returned 1 row. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_27ccc3db293ff14cc2c47b7e493a8946` `chunk_id=srcchunk_70a8a76f95a9ae4fecb2491003e6941d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813629.402109` `source_timestamp=2026-05-15T03:02:26Z`
- A query unioning counts for 9 event types (license_terms_attached, license_template_registered, licensing_config_set, derivative_registered, royalty_vault_deployed, royalty_paid, revenue_token_claimed, metadata_uri_set, dispute_raised) was executed on `story-blockchain-prod` and returned 9 rows. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_09280177eef114923c0e8cc39a8960b4` `chunk_id=srcchunk_d3a67a7f7749d4ce7accdb6e4ea2c5f0` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814168.049689` `source_timestamp=2026-05-15T03:03:19Z`
- Column metadata for table `ip_registered_events` was retrieved from `story-blockchain-prod`, returning 12 columns. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_59209477751beb471c1226a155b2e1b9` `chunk_id=srcchunk_27bfbf13e12cbbf2779fcae37f0519d1` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814240.977489` `source_timestamp=2026-05-15T03:04:41Z`
- Monthly IP registrations were counted from `ip_registered_events` for dates >= '2024-05-01' on `story-blockchain-prod`, returning 7 monthly rows. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_b62c308782b2e8d15db7cba45ca5a4ce` `chunk_id=srcchunk_cbd46de320ad09ff1600030efe7240da` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814307.883719` `source_timestamp=2026-05-15T03:20:05Z`
- Overall min block_timestamp, max block_timestamp, and total count were retrieved from `ip_registered_events` on `story-blockchain-prod`, returning 1 row. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_50345dbd1dc0510bd305d01309f9fde2` `chunk_id=srcchunk_e7aa03d52dfd5043c2ac460d031a1dfe` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815231.588419` `source_timestamp=2026-05-15T03:21:11Z`
- Column metadata for five tables (`derivative_registered_events`, `license_terms_attached_events`, `event_royalty_module_ip_royalty_vault_deployed`, `event_royalty_module_royalty_paid`, `event_core_metadata_module_metadata_uri_set`) was retrieved from `story-blockchain-prod`, returning 47 rows. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_e5947999833d240dd58fb13540a86cd5` `chunk_id=srcchunk_3e24c51c85f423a9519895793b70598b` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815353.469989` `source_timestamp=2026-05-15T03:22:58Z`
- The query `SELECT count(*) as node_count FROM nodes` was executed on `sos-royalty-graph-prod` and returned 1 row. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_de82fd8293eac32ed1f8352b207999f6` `chunk_id=srcchunk_9bb66dd041df9d45bb901b8ce2ec75ce` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815411.041569` `source_timestamp=2026-05-15T03:23:56Z`
- An attempt to count edges in the SOS royalty graph via `SELECT count(*) as edge_count FROM edges` was marked as stale (action ignored) but still returned 1 row. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0407be8724d740cb217f365f67d265ba` `chunk_id=srcchunk_2bb4864083a271ee1396925bcf10328d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815457.800779` `source_timestamp=2026-05-15T03:26:28Z`
- The query `SELECT count(*) as ip_assets_count FROM ip_assets` was executed on `sos-royalty-graph-prod` and returned 1 row. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_03c0a0c14b843c892de69bfa9a6840a9` `chunk_id=srcchunk_3324eef537033559f74c0a05c659c850` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815616.033609` `source_timestamp=2026-05-15T03:27:36Z`
- Overall min block_timestamp, max block_timestamp, and total count were retrieved from `derivative_registered_events` on `story-blockchain-prod`, returning 1 row. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_15ee32dc8a52b1280c4402a07c56df81` `chunk_id=srcchunk_c4c749b3ea1d66ec06e6703b365ea591` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815683.252819` `source_timestamp=2026-05-15T03:28:44Z`
- Monthly derivative registrations were counted from `derivative_registered_events` for dates >= '2024-05-01' on `story-blockchain-prod`, returning 16 monthly rows. `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_c10eff5992b4f454015e3f6a1864228c` `chunk_id=srcchunk_17cf38acac3eb8a2591f4f94087163a5` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815760.785249` `source_timestamp=2026-05-15T03:29:32Z`
- Monthly license terms attached events were counted from `license_terms_attached_events` for dates >= '2024-05-01' on `story-blockchain-prod`, returning 16 monthly rows. `claim:claim_1_16` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_67979ea2af05c8712ea32e4f6a997dcd` `chunk_id=srcchunk_5aaeaf14f3208a8f08ee4a79fac85061` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815793.478939` `source_timestamp=2026-05-15T03:30:29Z`
- Monthly royalty vault deployments were counted from `event_royalty_module_ip_royalty_vault_deployed` for dates >= '2024-05-01' on `story-blockchain-prod`, returning 10 monthly rows. `claim:claim_1_17` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_17a60977cb6997c3b94aca896dc9d632` `chunk_id=srcchunk_8fc593c79378f2628187849f58776020` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815860.435979` `source_timestamp=2026-05-15T03:31:43Z`
- Monthly metadata URI set events were counted from `event_core_metadata_module_metadata_uri_set` for dates >= '2024-05-01' on `story-blockchain-prod`, returning 10 monthly rows. `claim:claim_1_18` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_aa4440658f5649983e84c3d19492d4e9` `chunk_id=srcchunk_a91493be5a8f765d670cb1248b2bab4a` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815924.763979` `source_timestamp=2026-05-15T03:32:46Z`
- Monthly royalty paid events were counted from `event_royalty_module_royalty_paid` for dates >= '2024-05-01' on `story-blockchain-prod`, returning 10 monthly rows. `claim:claim_1_19` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_df6601fa312d779717eba38ff60f88d2` `chunk_id=srcchunk_b102718334a0ecf213f72f6b1c99ff98` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815987.878669` `source_timestamp=2026-05-15T03:33:23Z`
- Column metadata for table `event_ip_asset_registry_ip_registered` was retrieved from `story-blockchain-prod`, returning 12 columns. `claim:claim_1_20` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_124f9555dc99b6f877ee878f182b4f93` `chunk_id=srcchunk_4ec79fd59ce3ae6160c544862fc89522` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816059.769409` `source_timestamp=2026-05-15T03:34:40Z`
- Overall min created_at, max created_at, and total count were retrieved from `event_ip_asset_registry_ip_registered` on `story-blockchain-prod`, returning 1 row. `claim:claim_1_21` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_2e5822db4394d9419fedde165e1a62a2` `chunk_id=srcchunk_907100be67556c9ea97a805a2f17605f` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816101.701739` `source_timestamp=2026-05-15T03:35:45Z`
- The query `SELECT count(*) as collection_count FROM collection_aggregates` was executed on `sos-royalty-graph-prod` and returned 1 row. `claim:claim_1_22` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_2b355ccd5e2e28f9a69a3b15835c8d66` `chunk_id=srcchunk_185f339e34e496dbcc94883747de791c` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816172.244979` `source_timestamp=2026-05-15T03:36:43Z`
- A query unioning counts for `ip_licenses`, `ip_transactions`, and `royalty_token_transfers` was executed on `sos-royalty-graph-prod` and returned 3 rows. `claim:claim_1_23` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_88557bee07babdd58e2c3e9b10f7eb14` `chunk_id=srcchunk_fb3e7116479bf2342a34c7a24e03366b` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816224.376759` `source_timestamp=2026-05-15T03:38:41Z`

## Sources

- `source_document_id`: `srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2`
- `source_revision_id`: `srcrev_b62c308782b2e8d15db7cba45ca5a4ce`
