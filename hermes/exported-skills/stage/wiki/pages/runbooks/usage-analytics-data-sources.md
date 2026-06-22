---
title: "Protocol Usage Analytics Data Sources"
type: "runbook"
slug: "runbooks/usage-analytics-data-sources"
freshness: "2026-05-15T03:20:05Z"
tags:
  - "analytics"
  - "blockchain"
  - "data"
  - "protocol-usage"
owners:
  - "\u003c@U0772SH7BRA\u003e"
  - "\u003c@U0ASDQKU3UL\u003e"
source_revision_ids:
  - "srcrev_09280177eef114923c0e8cc39a8960b4"
  - "srcrev_0bca0e57717273b7b547debb0601386b"
  - "srcrev_27ccc3db293ff14cc2c47b7e493a8946"
  - "srcrev_59209477751beb471c1226a155b2e1b9"
  - "srcrev_b62c308782b2e8d15db7cba45ca5a4ce"
conflict_state: "none"
---

# Protocol Usage Analytics Data Sources

## Summary

Data sources and queries explored for generating protocol usage graphs for the last 2 years, per Yao's request. Covers sos-royalty-graph-prod and story-blockchain-prod databases.

## Claims

- The sos-royalty-graph-prod database contains tables for tracking IP graph data, including nodes, edges, ip_assets, ip_licenses, ip_transactions, ip_ancestor_descendant_pairs, collection_aggregates, and royalty_token_account_transfers. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_0bca0e57717273b7b547debb0601386b` `chunk_id=srcchunk_4a8a9eb260afcb83db1bad7a3c9ad1c8` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813407.495359` `source_timestamp=2026-05-15T02:53:18Z`
- The story-blockchain-prod database tracks multiple event types including ip_registered_events, license_terms_attached_events, license_template_registered_events, licensing_config_set_for_license_events, derivative_registered_events, event_royalty_module_ip_royalty_vault_deployed, event_royalty_module_royalty_paid, event_revenue_token_claimed, event_core_metadata_module_metadata_uri_set, and event_dispute_module_dispute_raised. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_09280177eef114923c0e8cc39a8960b4` `chunk_id=srcchunk_d3a67a7f7749d4ce7accdb6e4ea2c5f0` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814168.049689` `source_timestamp=2026-05-15T03:03:19Z`
- The ip_registered_events table in story-blockchain-prod has 12 columns. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_59209477751beb471c1226a155b2e1b9` `chunk_id=srcchunk_27bfbf13e12cbbf2779fcae37f0519d1` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814240.977489` `source_timestamp=2026-05-15T03:04:41Z`
- Monthly IP registration counts were obtained from May 2024 onwards. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_b62c308782b2e8d15db7cba45ca5a4ce` `chunk_id=srcchunk_cbd46de320ad09ff1600030efe7240da` `native_locator=slack:C04T5307FNU:1778813134.988149:1778814307.883719` `source_timestamp=2026-05-15T03:20:05Z`
- The total number of IP registrations was successfully queried. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2` `source_revision_id=srcrev_27ccc3db293ff14cc2c47b7e493a8946` `chunk_id=srcchunk_70a8a76f95a9ae4fecb2491003e6941d` `native_locator=slack:C04T5307FNU:1778813134.988149:1778813629.402109` `source_timestamp=2026-05-15T03:02:26Z`

## Open Questions

- Exact numeric results of queries were not captured in the conversation.

## Sources

- `source_document_id`: `srcdoc_ef9cf0857b9dda81ae554d4a7219cbc2`
- `source_revision_id`: `srcrev_15ee32dc8a52b1280c4402a07c56df81`
