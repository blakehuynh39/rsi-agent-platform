---
title: "Deprecation of UptimeRobot Monitoring"
type: "decision"
slug: "decisions/uptimerobot-deprecation"
freshness: "2026-02-05T18:17:13Z"
tags:
  - "betterstack"
  - "deprecation"
  - "monitoring"
  - "saas"
  - "uptimerobot"
owners: []
source_revision_ids:
  - "srcrev_28b7f5e9f53a924cfa38767152f8e713"
  - "srcrev_a973721c0e880f87383b20b507b6322d"
  - "srcrev_bd1842e710f7263db39ead3ae83afe86"
conflict_state: "none"
---

# Deprecation of UptimeRobot Monitoring

## Summary

UptimeRobot, previously used to monitor API endpoints, has been superseded by Betterstack. The subscription auto-renewal was cancelled, user accounts removed, and all alerts paused following a Slack discussion.

## Claims

- UptimeRobot was used to monitor API endpoints. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_bd1842e710f7263db39ead3ae83afe86` `chunk_id=srcchunk_3c03d6b5076bb9db915240dcae33e97c` `native_locator=slack:C0547N89JUB:1770149808.313609:1770164971.511409` `source_timestamp=2026-02-04T00:29:31Z`
- Betterstack is believed to have replaced UptimeRobot. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_a973721c0e880f87383b20b507b6322d` `chunk_id=srcchunk_2895f087b1e021f1776d7e7a0c0d390b` `native_locator=slack:C0547N89JUB:1770149808.313609:1770165017.666669` `source_timestamp=2026-02-04T00:30:17Z`
- UptimeRobot auto-renewal was cancelled, Andy and Blake's accounts removed, and all alerts paused. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_716e10e5fd9b4407d90bd35c7ef67a91` `source_revision_id=srcrev_28b7f5e9f53a924cfa38767152f8e713` `chunk_id=srcchunk_c1d9d0e75bee8dba7d9d6dd8b69d17d9` `native_locator=slack:C0547N89JUB:1770149808.313609:1770315433.948779` `source_timestamp=2026-02-05T18:17:13Z`

## Open Questions

- Is Betterstack fully covering all previous UptimeRobot monitors?
- Who was the designated approver for decommissioning? (Mentioned need for confirmation but not explicitly stated.)

## Sources

- `source_document_id`: `srcdoc_716e10e5fd9b4407d90bd35c7ef67a91`
- `source_revision_id`: `srcrev_28b7f5e9f53a924cfa38767152f8e713`
