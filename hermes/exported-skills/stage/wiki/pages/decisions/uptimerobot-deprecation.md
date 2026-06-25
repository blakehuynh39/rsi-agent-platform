---
title: "UptimeRobot Deprecation"
type: "decision"
slug: "decisions/uptimerobot-deprecation"
freshness: "2026-02-05T18:17:13Z"
tags:
  - "api-monitoring"
  - "betterstack"
  - "deprecation"
  - "monitoring"
  - "uptimerobot"
owners:
  - "Jack"
source_revision_ids:
  - "srcrev_28b7f5e9f53a924cfa38767152f8e713"
  - "srcrev_a973721c0e880f87383b20b507b6322d"
  - "srcrev_bd1842e710f7263db39ead3ae83afe86"
  - "srcrev_c6b02a3be5481ac264db534862467095"
conflict_state: "none"
---

# UptimeRobot Deprecation

## Summary

Decision and actions to decommission UptimeRobot and migrate to Betterstack. Auto-renewal cancelled, user accounts removed, alerts paused.

## Claims

- UptimeRobot was used to monitor API endpoints. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_bd1842e710f7263db39ead3ae83afe86` `chunk_id=srcchunk_3c03d6b5076bb9db915240dcae33e97c` `native_locator=slack:C0547N89JUB:1770149808.313609:1770164971.511409` `source_timestamp=2026-02-04T00:29:31Z`
- Betterstack replaced UptimeRobot as the monitoring tool. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_a973721c0e880f87383b20b507b6322d` `chunk_id=srcchunk_2895f087b1e021f1776d7e7a0c0d390b` `native_locator=slack:C0547N89JUB:1770149808.313609:1770165017.666669` `source_timestamp=2026-02-04T00:30:17Z`
- The decision was made to cancel the UptimeRobot plan after confirming it was no longer needed. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_c6b02a3be5481ac264db534862467095` `chunk_id=srcchunk_ef4d8bdccc5ce3d681d6881d374bded6` `native_locator=slack:C0547N89JUB:1770149808.313609:1770165101.973539` `source_timestamp=2026-02-04T00:31:41Z`
- The UptimeRobot auto-renewal was cancelled. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_28b7f5e9f53a924cfa38767152f8e713` `chunk_id=srcchunk_c1d9d0e75bee8dba7d9d6dd8b69d17d9` `native_locator=slack:C0547N89JUB:1770149808.313609:1770315433.948779` `source_timestamp=2026-02-05T18:17:13Z`
- Andy and Blake's UptimeRobot accounts were removed. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_28b7f5e9f53a924cfa38767152f8e713` `chunk_id=srcchunk_c1d9d0e75bee8dba7d9d6dd8b69d17d9` `native_locator=slack:C0547N89JUB:1770149808.313609:1770315433.948779` `source_timestamp=2026-02-05T18:17:13Z`
- All UptimeRobot alerts were paused. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_28b7f5e9f53a924cfa38767152f8e713` `chunk_id=srcchunk_c1d9d0e75bee8dba7d9d6dd8b69d17d9` `native_locator=slack:C0547N89JUB:1770149808.313609:1770315433.948779` `source_timestamp=2026-02-05T18:17:13Z`

## Sources

- `source_document_id`: `srcdoc_716e10e5fd9b4407d90bd35c7ef67a91`
- `source_revision_id`: `srcrev_e61d90e02f5834af4c654baaed2c457e`
