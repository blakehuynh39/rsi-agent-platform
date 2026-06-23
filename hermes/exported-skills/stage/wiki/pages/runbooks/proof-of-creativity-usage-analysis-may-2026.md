---
title: "Proof of Creativity Usage Analysis (May 2026)"
type: "runbook"
slug: "runbooks/proof-of-creativity-usage-analysis-may-2026"
freshness: "2026-05-15T03:38:41Z"
tags:
  - "analysis"
  - "data-query"
  - "ip-graph"
  - "proof-of-creativity"
  - "usage-stats"
owners:
  - "U0772SH7BRA"
  - "U0ASDQKU3UL"
source_revision_ids:
  - "srcrev_03c0a0c14b843c892de69bfa9a6840a9"
  - "srcrev_0407be8724d740cb217f365f67d265ba"
  - "srcrev_09280177eef114923c0e8cc39a8960b4"
  - "srcrev_0bca0e57717273b7b547debb0601386b"
  - "srcrev_124f9555dc99b6f877ee878f182b4f93"
  - "srcrev_15ee32dc8a52b1280c4402a07c56df81"
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
  - "srcrev_bab416db7564b1a0a9bec7c85fde1838"
  - "srcrev_c10eff5992b4f454015e3f6a1864228c"
  - "srcrev_de82fd8293eac32ed1f8352b207999f6"
  - "srcrev_df6601fa312d779717eba38ff60f88d2"
  - "srcrev_e5947999833d240dd58fb13540a86cd5"
  - "srcrev_f7c168547d4afd8e2051a619863029cf"
conflict_state: "none"
---

# Proof of Creativity Usage Analysis (May 2026)

## Summary

A data analysis task to obtain high-level usage statistics for the Story Protocol IP graph/Proof of Creativity features over the last 2 years, utilizing sos-royalty-graph-prod and story-blockchain-prod databases.

## Claims

- A request was made to generate a general usage graph/plot for IP graph/Proof of Creativity using data from sos-royalty-graph-prod and story-blockchain-prod. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_f7c168547d4afd8e2051a619863029cf` `chunk_id=srcchunk_fd5e8f9cf1133a7b6f59e672f000b695` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813134.988149` `source_timestamp=2026-05-15T02:45:34Z`
- The analysis session was tracked in the RSI platform. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_bab416db7564b1a0a9bec7c85fde1838` `chunk_id=srcchunk_ac53e5357b01272ed2d922a9a25214a8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813139.452969` `source_timestamp=2026-05-15T02:45:39Z`
- The sos-royalty-graph-prod database has 31 tables in the public schema. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_3eb9f34e73e2d3a1c2418bf4b0e76074` `chunk_id=srcchunk_d8e21e7caf9aef76c156ca7b3bf2438d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813262.966189` `source_timestamp=2026-05-15T02:48:23Z`
- The story-blockchain-prod database has 39 tables in the public schema. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_1ede3f246ed1b8682db7ce2652ded978` `chunk_id=srcchunk_9351c645bae65adc6faf6fb6d6b15769` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813335.613029` `source_timestamp=2026-05-15T02:49:36Z`
- A query to fetch aggregate counts from multiple tables in sos-royalty-graph-prod timed out. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0bca0e57717273b7b547debb0601386b` `chunk_id=srcchunk_4a8a9eb260afcb83db1bad7a3c9ad1c8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813407.495359` `source_timestamp=2026-05-15T02:53:18Z`
- A query to select the total number of ip_registered_events succeeded (rows=1). `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_27ccc3db293ff14cc2c47b7e493a8946` `chunk_id=srcchunk_70a8a76f95a9ae4fecb2491003e6941d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813629.402109` `source_timestamp=2026-05-15T03:02:26Z`
- A query fetching counts for 9 different event types in story-blockchain-prod succeeded (rows=9). `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_09280177eef114923c0e8cc39a8960b4` `chunk_id=srcchunk_d3a67a7f7749d4ce7accdb6e4ea2c5f0` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814168.049689` `source_timestamp=2026-05-15T03:03:19Z`
- The ip_registered_events table has 12 columns (names and types retrieved). `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_59209477751beb471c1226a155b2e1b9` `chunk_id=srcchunk_27bfbf13e12cbbf2779fcae37f0519d1` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814240.977489` `source_timestamp=2026-05-15T03:04:41Z`
- Monthly counts of ip_registered_events from May 2024 onwards were successfully retrieved (rows=7). `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_b62c308782b2e8d15db7cba45ca5a4ce` `chunk_id=srcchunk_cbd46de320ad09ff1600030efe7240da` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814307.883719` `source_timestamp=2026-05-15T03:20:05Z`
- The earliest and latest ip_registered_events timestamps and total count were successfully queried (rows=1). `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_50345dbd1dc0510bd305d01309f9fde2` `chunk_id=srcchunk_e7aa03d52dfd5043c2ac460d031a1dfe` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815231.588419` `source_timestamp=2026-05-15T03:21:11Z`
- Column metadata for derivative_registered_events, license_terms_attached_events, and other related tables were retrieved (47 rows). `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_e5947999833d240dd58fb13540a86cd5` `chunk_id=srcchunk_3e24c51c85f423a9519895793b70598b` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815353.469989` `source_timestamp=2026-05-15T03:22:58Z`
- The node count from the nodes table in sos-royalty-graph-prod was successfully queried (rows=1). `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_de82fd8293eac32ed1f8352b207999f6` `chunk_id=srcchunk_9bb66dd041df9d45bb901b8ce2ec75ce` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815411.041569` `source_timestamp=2026-05-15T03:23:56Z`
- A query for the edge count from edges table was attempted but resulted in a stale action (ignored). `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0407be8724d740cb217f365f67d265ba` `chunk_id=srcchunk_2bb4864083a271ee1396925bcf10328d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815457.800779` `source_timestamp=2026-05-15T03:26:28Z`
- The count of ip_assets in sos-royalty-graph-prod was successfully queried (rows=1). `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_03c0a0c14b843c892de69bfa9a6840a9` `chunk_id=srcchunk_3324eef537033559f74c0a05c659c850` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815616.033609` `source_timestamp=2026-05-15T03:27:36Z`
- Timestamp range and total count for derivative_registered_events were successfully retrieved (rows=1). `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_15ee32dc8a52b1280c4402a07c56df81` `chunk_id=srcchunk_c4c749b3ea1d66ec06e6703b365ea591` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815683.252819` `source_timestamp=2026-05-15T03:28:44Z`
- Monthly derivative_registered_events counts from May 2024 were successfully retrieved (rows=16). `claim:claim_1_16` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_c10eff5992b4f454015e3f6a1864228c` `chunk_id=srcchunk_17cf38acac3eb8a2591f4f94087163a5` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815760.785249` `source_timestamp=2026-05-15T03:29:32Z`
- Monthly license_terms_attached_events counts from May 2024 were successfully fetched (rows=16). `claim:claim_1_17` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_67979ea2af05c8712ea32e4f6a997dcd` `chunk_id=srcchunk_5aaeaf14f3208a8f08ee4a79fac85061` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815793.478939` `source_timestamp=2026-05-15T03:30:29Z`
- Monthly event_royalty_module_ip_royalty_vault_deployed counts from May 2024 were successfully fetched (rows=10). `claim:claim_1_18` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_aa4440658f5649983e84c3d19492d4e9` `chunk_id=srcchunk_a91493be5a8f765d670cb1248b2bab4a` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815924.763979` `source_timestamp=2026-05-15T03:32:46Z`
- Monthly event_core_metadata_module_metadata_uri_set counts from May 2024 were successfully fetched (rows=10). `claim:claim_1_19` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_aa4440658f5649983e84c3d19492d4e9` `chunk_id=srcchunk_a91493be5a8f765d670cb1248b2bab4a` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815924.763979` `source_timestamp=2026-05-15T03:32:46Z`
- Monthly event_royalty_module_royalty_paid counts from May 2024 were successfully fetched (rows=10). `claim:claim_1_20` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_df6601fa312d779717eba38ff60f88d2` `chunk_id=srcchunk_b102718334a0ecf213f72f6b1c99ff98` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815987.878669` `source_timestamp=2026-05-15T03:33:23Z`
- Column metadata for event_ip_asset_registry_ip_registered was retrieved (12 columns). `claim:claim_1_21` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_124f9555dc99b6f877ee878f182b4f93` `chunk_id=srcchunk_4ec79fd59ce3ae6160c544862fc89522` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816059.769409` `source_timestamp=2026-05-15T03:34:40Z`
- Timestamp range and total count for event_ip_asset_registry_ip_registered were successfully retrieved (rows=1). `claim:claim_1_22` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_2e5822db4394d9419fedde165e1a62a2` `chunk_id=srcchunk_907100be67556c9ea97a805a2f17605f` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816101.701739` `source_timestamp=2026-05-15T03:35:45Z`
- The count of collection_aggregates in sos-royalty-graph-prod was successfully queried (rows=1). `claim:claim_1_23` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_2b355ccd5e2e28f9a69a3b15835c8d66` `chunk_id=srcchunk_185f339e34e496dbcc94883747de791c` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816172.244979` `source_timestamp=2026-05-15T03:36:43Z`
- Counts of ip_licenses, ip_transactions, and royalty_token_transfers from sos-royalty-graph-prod were successfully retrieved (rows=3). `claim:claim_1_24` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_88557bee07babdd58e2c3e9b10f7eb14` `chunk_id=srcchunk_fb3e7116479bf2342a34c7a24e03366b` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816224.376759` `source_timestamp=2026-05-15T03:38:41Z`

## Open Questions

- Actual numeric results of the queries were not captured in these logs; only metadata about query execution (row counts) is available.

## Sources

- `source_document_id`: `srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2`
- `source_revision_id`: `srcrev_373a11538dfa4c9577902f8b902d95ac`
