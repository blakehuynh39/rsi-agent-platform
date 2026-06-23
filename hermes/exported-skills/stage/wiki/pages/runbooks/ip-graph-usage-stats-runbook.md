---
title: "IP Graph Usage Statistics Runbook"
type: "runbook"
slug: "runbooks/ip-graph-usage-stats-runbook"
freshness: "2026-05-15T03:38:41Z"
tags:
  - "blockchain"
  - "queries"
  - "runbook"
  - "sos"
  - "statistics"
  - "story-protocol"
  - "usage"
owners:
  - "U0772SH7BRA"
  - "U0ASDQKU3UL"
source_revision_ids:
  - "srcrev_03c0a0c14b843c892de69bfa9a6840a9"
  - "srcrev_0407be8724d740cb217f365f67d265ba"
  - "srcrev_09280177eef114923c0e8cc39a8960b4"
  - "srcrev_0bca0e57717273b7b547debb0601386b"
  - "srcrev_17a60977cb6997c3b94aca896dc9d632"
  - "srcrev_1ede3f246ed1b8682db7ce2652ded978"
  - "srcrev_27ccc3db293ff14cc2c47b7e493a8946"
  - "srcrev_2b355ccd5e2e28f9a69a3b15835c8d66"
  - "srcrev_2e5822db4394d9419fedde165e1a62a2"
  - "srcrev_3eb9f34e73e2d3a1c2418bf4b0e76074"
  - "srcrev_50345dbd1dc0510bd305d01309f9fde2"
  - "srcrev_67979ea2af05c8712ea32e4f6a997dcd"
  - "srcrev_88557bee07babdd58e2c3e9b10f7eb14"
  - "srcrev_aa4440658f5649983e84c3d19492d4e9"
  - "srcrev_b62c308782b2e8d15db7cba45ca5a4ce"
  - "srcrev_bab416db7564b1a0a9bec7c85fde1838"
  - "srcrev_c10eff5992b4f454015e3f6a1864228c"
  - "srcrev_de82fd8293eac32ed1f8352b207999f6"
  - "srcrev_df6601fa312d779717eba38ff60f88d2"
conflict_state: "none"
---

# IP Graph Usage Statistics Runbook

## Summary

Runbook for querying usage statistics of the Story Protocol IP graph and blockchain events from production databases sos-royalty-graph-prod and story-blockchain-prod. Covers table inventories, key event counts, and monthly trends for IP registrations, derivatives, license attachments, royalty vaults, metadata, and more. Run tracked at https://staging-rsi-platform.storyprotocol.net/sessions?conversation=conv-54773b52416b45fa931ca5351d64163b&tab=conversations&trace=trace-629642f551e84133ab24c3706a6caa4f

## Claims

- The `sos-royalty-graph-prod` database contains 31 tables in the public schema, including `nodes`, `edges`, `ip_assets`, `ip_licenses`, `ip_transactions`, `ip_ancestor_descendant_pairs`, `collection_aggregates`, and `royalty_token_account_transfers`. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_3eb9f34e73e2d3a1c2418bf4b0e76074` `chunk_id=srcchunk_d8e21e7caf9aef76c156ca7b3bf2438d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813262.966189` `source_timestamp=2026-05-15T02:48:23Z`
- The `story-blockchain-prod` database contains 39 tables in the public schema, including event tables such as `ip_registered_events`, `derivative_registered_events`, `license_terms_attached_events`, `licensing_config_set_for_license_events`, `event_royalty_module_ip_royalty_vault_deployed`, `event_royalty_module_royalty_paid`, `event_revenue_token_claimed`, `event_core_metadata_module_metadata_uri_set`, and `event_dispute_module_dispute_raised`. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_1ede3f246ed1b8682db7ce2652ded978` `chunk_id=srcchunk_9351c645bae65adc6faf6fb6d6b15769` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813335.613029` `source_timestamp=2026-05-15T02:49:36Z`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_09280177eef114923c0e8cc39a8960b4` `chunk_id=srcchunk_d3a67a7f7749d4ce7accdb6e4ea2c5f0` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814168.049689` `source_timestamp=2026-05-15T03:03:19Z`
- A query to get counts for `nodes`, `edges`, `ip_assets`, `ip_licenses`, `ip_transactions`, `ip_ancestor_descendant_pairs`, `collection_aggregates`, and `royalty_token_account_transfers` from `sos-royalty-graph-prod` timed out. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0bca0e57717273b7b547debb0601386b` `chunk_id=srcchunk_4a8a9eb260afcb83db1bad7a3c9ad1c8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813407.495359` `source_timestamp=2026-05-15T02:53:18Z`
- A query for total IP registrations count in `ip_registered_events` succeeded but the result row is not displayed. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_27ccc3db293ff14cc2c47b7e493a8946` `chunk_id=srcchunk_70a8a76f95a9ae4fecb2491003e6941d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813629.402109` `source_timestamp=2026-05-15T03:02:26Z`
- Monthly IP registration counts from `ip_registered_events` since May 2024 returned 7 rows, indicating data for 7 months. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_b62c308782b2e8d15db7cba45ca5a4ce` `chunk_id=srcchunk_cbd46de320ad09ff1600030efe7240da` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814307.883719` `source_timestamp=2026-05-15T03:20:05Z`
- A query for the earliest and latest timestamp and total count of `ip_registered_events` succeeded. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_50345dbd1dc0510bd305d01309f9fde2` `chunk_id=srcchunk_e7aa03d52dfd5043c2ac460d031a1dfe` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815231.588419` `source_timestamp=2026-05-15T03:21:11Z`
- Node count query on `sos-royalty-graph-prod` succeeded. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_de82fd8293eac32ed1f8352b207999f6` `chunk_id=srcchunk_9bb66dd041df9d45bb901b8ce2ec75ce` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815411.041569` `source_timestamp=2026-05-15T03:23:56Z`
- Edge count query on `sos-royalty-graph-prod` encountered a stale action and was ignored. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0407be8724d740cb217f365f67d265ba` `chunk_id=srcchunk_2bb4864083a271ee1396925bcf10328d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815457.800779` `source_timestamp=2026-05-15T03:26:28Z`
- IP assets count query on `sos-royalty-graph-prod` succeeded. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_03c0a0c14b843c892de69bfa9a6840a9` `chunk_id=srcchunk_3324eef537033559f74c0a05c659c850` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815616.033609` `source_timestamp=2026-05-15T03:27:36Z`
- Collection aggregates count query on `sos-royalty-graph-prod` succeeded. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_2b355ccd5e2e28f9a69a3b15835c8d66` `chunk_id=srcchunk_185f339e34e496dbcc94883747de791c` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816172.244979` `source_timestamp=2026-05-15T03:36:43Z`
- Query for license, transaction, and royalty token transfer counts from `sos-royalty-graph-prod` succeeded with 3 result rows. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_88557bee07babdd58e2c3e9b10f7eb14` `chunk_id=srcchunk_fb3e7116479bf2342a34c7a24e03366b` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816224.376759` `source_timestamp=2026-05-15T03:38:41Z`
- Monthly derivative registrations since May 2024 returned 16 rows. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_c10eff5992b4f454015e3f6a1864228c` `chunk_id=srcchunk_17cf38acac3eb8a2591f4f94087163a5` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815760.785249` `source_timestamp=2026-05-15T03:29:32Z`
- Monthly license terms attached counts since May 2024 returned 16 rows. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_67979ea2af05c8712ea32e4f6a997dcd` `chunk_id=srcchunk_5aaeaf14f3208a8f08ee4a79fac85061` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815793.478939` `source_timestamp=2026-05-15T03:30:29Z`
- Monthly royalty vault deployments since May 2024 returned 10 rows. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_17a60977cb6997c3b94aca896dc9d632` `chunk_id=srcchunk_8fc593c79378f2628187849f58776020` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815860.435979` `source_timestamp=2026-05-15T03:31:43Z`
- Monthly metadata URI set events since May 2024 returned 10 rows. `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_aa4440658f5649983e84c3d19492d4e9` `chunk_id=srcchunk_a91493be5a8f765d670cb1248b2bab4a` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815924.763979` `source_timestamp=2026-05-15T03:32:46Z`
- Monthly royalty paid events since May 2024 returned 10 rows. `claim:claim_1_16` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_df6601fa312d779717eba38ff60f88d2` `chunk_id=srcchunk_b102718334a0ecf213f72f6b1c99ff98` `native_locator=slack:C04T5307FNU:1778813134.988149:1778815987.878669` `source_timestamp=2026-05-15T03:33:23Z`
- The `story-blockchain-prod` also contains an `event_ip_asset_registry_ip_registered` table with 12 columns. `claim:claim_1_17` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_2e5822db4394d9419fedde165e1a62a2` `chunk_id=srcchunk_907100be67556c9ea97a805a2f17605f` `native_locator=slack:C04T5307FNU:1778813134.988149:1778816101.701739` `source_timestamp=2026-05-15T03:35:45Z`
- A trace for this RSI run is available at the staging RSI platform URL. `claim:claim_1_18` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_bab416db7564b1a0a9bec7c85fde1838` `chunk_id=srcchunk_ac53e5357b01272ed2d922a9a25214a8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813139.452969` `source_timestamp=2026-05-15T02:45:39Z`

## Open Questions

- Actual count values for IP registrations, nodes, edges, IP assets, and other metrics are not captured in the logs; the queries succeeded but the results were not stored. How can these values be retrieved?
- Are there other event tables or data sources that should be included in a comprehensive usage report?
- The query for multiple counts in sos-royalty-graph-prod timed out. What optimization is needed?

## Sources

- `source_document_id`: `srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2`
- `source_revision_id`: `srcrev_0146faa362493022817a790d705f3103`
