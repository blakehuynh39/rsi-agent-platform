---
title: "User Certificate Deepfake Screening"
type: "project"
slug: "projects/user-certificate-deepfake-screening"
freshness: "2026-05-29T17:34:10Z"
tags:
  - "c2pa"
  - "certificates"
  - "deepfake"
  - "screening"
  - "synthid"
owners:
  - "U0927FP6HH9"
  - "U0A2D9U625V"
source_revision_ids:
  - "srcrev_2201ac3cd9883c22c9ca35ce3a5afa87"
  - "srcrev_664be2e779da032d1aa54a6aa2567bbc"
  - "srcrev_717ed486e320e5245bf0297f92d547f4"
  - "srcrev_737f4ab90a213b179173e27d6e6e5912"
  - "srcrev_af1ec9d04aef713e4a8d74b33c323670"
  - "srcrev_bf500218d04131be9d51c058b0b57343"
conflict_state: "none"
---

# User Certificate Deepfake Screening

## Summary

Initiative to detect deepfake certificates submitted by users using SynthID, C2PA watermark detection, and LinkedIn profile verification.

## Claims

- Initial user data contains deepfake certificates. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de268702ce75826802f9899256b44395` `source_revision_id=srcrev_bf500218d04131be9d51c058b0b57343` `chunk_id=srcchunk_0fbda6edba25e9eb167b7ca5c5cf8072` `native_locator=slack:C0AL7EKNHDF:1780032757.138519:1780032757.138519` `source_timestamp=2026-05-29T05:36:36Z`
- A plan is needed to validate user-submitted PDFs and images for deepfakes. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de268702ce75826802f9899256b44395` `source_revision_id=srcrev_bf500218d04131be9d51c058b0b57343` `chunk_id=srcchunk_0fbda6edba25e9eb167b7ca5c5cf8072` `native_locator=slack:C0AL7EKNHDF:1780032757.138519:1780032757.138519` `source_timestamp=2026-05-29T05:36:36Z`
- C2PA and watermark detection are proposed as validation methods. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de268702ce75826802f9899256b44395` `source_revision_id=srcrev_737f4ab90a213b179173e27d6e6e5912` `chunk_id=srcchunk_b8408320614c09875ee066e530809a97` `native_locator=slack:C0AL7EKNHDF:1780032757.138519:1780061336.754969` `source_timestamp=2026-05-29T13:28:56Z`
- Google’s SynthID should be explored for deepfake detection. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de268702ce75826802f9899256b44395` `source_revision_id=srcrev_af1ec9d04aef713e4a8d74b33c323670` `chunk_id=srcchunk_5b86c3d913e1162f43c530f0c19d79e8` `native_locator=slack:C0AL7EKNHDF:1780032757.138519:1780072170.201099` `source_timestamp=2026-05-29T16:29:30Z`
- The Poseidon team currently lacks resources for new document processing tasks. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de268702ce75826802f9899256b44395` `source_revision_id=srcrev_717ed486e320e5245bf0297f92d547f4` `chunk_id=srcchunk_f65d9de7431c7d59a695f58dc9ebaf04` `native_locator=slack:C0AL7EKNHDF:1780032757.138519:1780074279.207179` `source_timestamp=2026-05-29T17:04:39Z`
- Collecting LinkedIn profile alongside resume is suggested to verify user legitimacy. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de268702ce75826802f9899256b44395` `source_revision_id=srcrev_2201ac3cd9883c22c9ca35ce3a5afa87` `chunk_id=srcchunk_9c5af7729a47fcdb0f08ed344dcdc7ec` `native_locator=slack:C0AL7EKNHDF:1780032757.138519:1780076003.837949` `source_timestamp=2026-05-29T17:33:23Z`
- SynthID detection should be implemented at application side if possible. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_de268702ce75826802f9899256b44395` `source_revision_id=srcrev_664be2e779da032d1aa54a6aa2567bbc` `chunk_id=srcchunk_84def6b3e0d291edda99e1cb5973e76a` `native_locator=slack:C0AL7EKNHDF:1780032757.138519:1780076050.578509` `source_timestamp=2026-05-29T17:34:10Z`

## Open Questions

- Can SynthID detection be implemented client-side?
- What other document types need validation?
- Who will build the screening process?

## Sources

- `source_document_id`: `srcdoc_de268702ce75826802f9899256b44395`
- `source_revision_id`: `srcrev_664be2e779da032d1aa54a6aa2567bbc`
