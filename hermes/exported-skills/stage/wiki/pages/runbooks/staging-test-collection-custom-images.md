---
title: "Staging Test Collection \u0026 Custom Images"
type: "runbook"
slug: "runbooks/staging-test-collection-custom-images"
freshness: "2026-05-05T19:39:36Z"
tags:
  - "custom-images"
  - "ip-registration"
  - "staging"
  - "testing"
owners:
  - "U04L0DD6B6F"
  - "U08951K4SRY"
  - "U08V4SFU7LZ"
source_revision_ids:
  - "srcrev_2648e5cf5aff72b72b0801b0b25fcb77"
  - "srcrev_3070d11619a6a0e70af5438396d71943"
  - "srcrev_848c512f9b55d436c24f6ec3573850d0"
  - "srcrev_a5b399799c910d7ce1d6bf3950b6a45a"
  - "srcrev_abe61d81cea7026e3fa1ebe88f6a8405"
conflict_state: "none"
---

# Staging Test Collection & Custom Images

## Summary

Process for testing IP registration and custom images on staging environment.

## Claims

- A test collection named "test collection" should be created on staging and 2-3 IPs registered to verify the workflow without confusing actual campaigns. `claim:claim_3_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_848c512f9b55d436c24f6ec3573850d0` `chunk_id=srcchunk_9c32b0f3442b0a485087a68ca9ae4fd5` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1777997016.653969` `source_timestamp=2026-05-05T16:04:14Z`
- Custom images are only wired back on production, not on staging. `claim:claim_3_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_abe61d81cea7026e3fa1ebe88f6a8405` `chunk_id=srcchunk_ad7285bca784b9bc1d41090d429fe9dc` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778009394.642689` `source_timestamp=2026-05-05T19:30:15Z`
- Two approaches proposed for sanity check: (a) create a test campaign on prod with a purpose-generated image and delete afterwards, or (b) wire back custom images on staging by removing environment check and merge to prod. `claim:claim_3_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_abe61d81cea7026e3fa1ebe88f6a8405` `chunk_id=srcchunk_ad7285bca784b9bc1d41090d429fe9dc` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778009394.642689` `source_timestamp=2026-05-05T19:30:15Z`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_3070d11619a6a0e70af5438396d71943` `chunk_id=srcchunk_56f938932aebf70ea985151ee9a1dc2a` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778009893.841379` `source_timestamp=2026-05-05T19:38:13Z`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_2648e5cf5aff72b72b0801b0b25fcb77` `chunk_id=srcchunk_bef21a1b9683abca8e8d3a721d0b49cc` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778009976.685059` `source_timestamp=2026-05-05T19:39:36Z`
- A test image for the campaign can be downloaded from the admin dashboard. `claim:claim_3_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_a5b399799c910d7ce1d6bf3950b6a45a` `chunk_id=srcchunk_6d7650671c9ffc5fdd764138faf30a2b` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778007584.797539` `source_timestamp=2026-05-05T18:59:44Z`

## Open Questions

- Comfort level with staging-to-prod merge for image wiring.
- Final decision on which approach to use for custom image sanity check.

## Related Pages

- `filipino-language-campaign`
- `ip-registration-auto-registration-issue`

## Sources

- `source_document_id`: `srcdoc_d15c947836e0a27e3d68a094607272ca`
- `source_revision_id`: `srcrev_eec3d3d009bf46a3831405890d417b81`
