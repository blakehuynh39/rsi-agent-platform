---
title: "IP Registration Auto-Registration Issue"
type: "runbook"
slug: "runbooks/ip-registration-auto-registration-issue"
freshness: "2026-05-05T23:16:42Z"
tags:
  - "bug"
  - "campaign"
  - "ip-registration"
  - "staging"
owners:
  - "U04L0DD6B6F"
  - "U067QP5PD6J"
  - "U08V4SFU7LZ"
source_revision_ids:
  - "srcrev_2fcc8bacf6a88ccf1d84edeeb7d6d734"
  - "srcrev_375ab352cfa4398365debd8482166ce7"
  - "srcrev_5139145ebbc76f2491e87f2652609e15"
  - "srcrev_ad9ea9ef42b2b601ac29c31bb3e555bd"
  - "srcrev_b64747f8d0a50d94bd6b94c429ae6393"
conflict_state: "none"
---

# IP Registration Auto-Registration Issue

## Summary

Inactive campaigns were automatically registering IPs on-chain due to default ip_registration_status=OPEN.

## Claims

- Inactive collections created on May 2 were already registered on-chain, evidenced by a Story Protocol portal link. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_b64747f8d0a50d94bd6b94c429ae6393` `chunk_id=srcchunk_711a865e4407773b50a7861852085b0e` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778018838.984969` `source_timestamp=2026-05-05T22:07:18Z`
- Root cause: all campaigns were created with `ip_registration_status = OPEN`, causing the worker to register them as IPs immediately. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_5139145ebbc76f2491e87f2652609e15` `chunk_id=srcchunk_665302c61840bd7cd40b668dd1704a6d` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778023002.363009` `source_timestamp=2026-05-05T23:16:42Z`
- A fix is being worked on to prevent automatic registration for inactive campaigns. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_375ab352cfa4398365debd8482166ce7` `chunk_id=srcchunk_4c56113da83e9b6df63930be078bfd16` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778019465.093299` `source_timestamp=2026-05-05T22:22:15Z`
- Already registered IPs can be handled by downloading, modifying, and re-uploading the `ip.json` file for each campaign. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_ad9ea9ef42b2b601ac29c31bb3e555bd` `chunk_id=srcchunk_42ff6d5d8f17ca07c6824107afb30eb3` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778022232.515869` `source_timestamp=2026-05-05T23:03:52Z`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_2fcc8bacf6a88ccf1d84edeeb7d6d734` `chunk_id=srcchunk_ee2704de794c0a95c42582fb462218b5` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778022353.406249` `source_timestamp=2026-05-05T23:05:53Z`

## Open Questions

- Potential concern about empty collections on chain revealing upcoming campaigns.
- Should we re-register when activating them or just update images? (Decided: update images)

## Related Pages

- `filipino-language-campaign`
- `staging-test-collection-custom-images`

## Sources

- `source_document_id`: `srcdoc_d15c947836e0a27e3d68a094607272ca`
- `source_revision_id`: `srcrev_eec3d3d009bf46a3831405890d417b81`
