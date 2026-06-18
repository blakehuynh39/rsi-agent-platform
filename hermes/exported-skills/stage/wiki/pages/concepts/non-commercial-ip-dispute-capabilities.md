---
title: "Non-Commercial IP Dispute Capabilities"
type: "concept"
slug: "concepts/non-commercial-ip-dispute-capabilities"
freshness: "2025-09-16T20:12:22Z"
tags:
  - "attestation-service"
  - "ip-disputes"
  - "magma"
  - "portal"
owners: []
source_revision_ids:
  - "srcrev_1b7beb70d022e33197612272f56a1d23"
  - "srcrev_1d8770eb3f95ac02679fd4c5820b1b6a"
  - "srcrev_a76da6e0ee9388cd6e9facb8dd502443"
  - "srcrev_c408a7e23b3e96edb180e22628c07f0f"
  - "srcrev_d629393d2f469ab5c76517c00f7d676c"
conflict_state: "none"
---

# Non-Commercial IP Dispute Capabilities

## Summary

This page captures the current understanding of whether disputes can be raised for non-commercial IP on the Story platform. The Portal seems to have a limitation, and IP infringement checks by the Attestation Service are only conducted for commercial IP. The Magma use case highlights the need to flag or dispute stolen non-commercial artworks.

## Claims

- There is a current limitation on the Portal for disputing non-commercial IP. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2f0512b415b3ad57983da61ff2ef726a` `source_revision_id=srcrev_a76da6e0ee9388cd6e9facb8dd502443` `chunk_id=srcchunk_e00484c5b18c036f7e429455540d69d4` `native_locator=slack:C04T5307FNU:1758045838.529879:1758047524.642079` `source_timestamp=2025-09-16T18:32:04Z`
- IP infringement checks by the Attestation Service are only run for commercial IP. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2f0512b415b3ad57983da61ff2ef726a` `source_revision_id=srcrev_1b7beb70d022e33197612272f56a1d23` `chunk_id=srcchunk_49de8c0ad8be88746b14df9495abd750` `native_locator=slack:C04T5307FNU:1758045838.529879:1758048233.138269` `source_timestamp=2025-09-16T18:43:53Z`
- Magma seeks a way to dispute or hide stolen artwork registrations on Story, especially for non-commercial IP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2f0512b415b3ad57983da61ff2ef726a` `source_revision_id=srcrev_1b7beb70d022e33197612272f56a1d23` `chunk_id=srcchunk_49de8c0ad8be88746b14df9495abd750` `native_locator=slack:C04T5307FNU:1758045838.529879:1758048233.138269` `source_timestamp=2025-09-16T18:43:53Z`
  - citation: `source_document_id=srcdoc_2f0512b415b3ad57983da61ff2ef726a` `source_revision_id=srcrev_d629393d2f469ab5c76517c00f7d676c` `chunk_id=srcchunk_0cbdc5764c1e12e1b551ff52a9d46b51` `native_locator=slack:C04T5307FNU:1758045838.529879:1758049679.467439` `source_timestamp=2025-09-16T19:07:59Z`
- There is a suggestion that Attestation could at least flag infringing works in Portal. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2f0512b415b3ad57983da61ff2ef726a` `source_revision_id=srcrev_c408a7e23b3e96edb180e22628c07f0f` `chunk_id=srcchunk_c55cb6d61a384d743896798ad99aefbb` `native_locator=slack:C04T5307FNU:1758045838.529879:1758051298.229189` `source_timestamp=2025-09-16T19:34:58Z`
- There is a need to clarify what the Attestation Service actually does. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2f0512b415b3ad57983da61ff2ef726a` `source_revision_id=srcrev_1d8770eb3f95ac02679fd4c5820b1b6a` `chunk_id=srcchunk_628a12ca3d0d511b233d07d61e1a8f25` `native_locator=slack:C04T5307FNU:1758045838.529879:1758053542.935649` `source_timestamp=2025-09-16T20:12:22Z`

## Open Questions

- Can disputes be raised for non-commercial IP on Story, either via Portal or other means?
- How can Magma be granted access to flag or hide stolen works in the short term?
- Is there a way to automate IP protections for non-registered works on Magma's platform?
- What is the full capability of the Attestation Service regarding flagging non-commercial IP?
- What would be the end state if a creator or Magma admin reports a registered work as infringing via the Attestation Service?

## Sources

- `source_document_id`: `srcdoc_2f0512b415b3ad57983da61ff2ef726a`
- `source_revision_id`: `srcrev_5369dbca45237f9660fecda992dd2a95`
