---
title: "Spoken Transcripts Storage Decision"
type: "decision"
slug: "decisions/spoken-transcripts-storage-decision"
freshness: "2026-04-24T20:30:04Z"
tags:
  - "aws-s3"
  - "cloudflare-r2"
  - "decision"
  - "spoken-transcripts"
  - "storage"
owners: []
source_revision_ids:
  - "srcrev_02924860b195c0990ba32819c39dc59e"
  - "srcrev_28e02c00ce34503b66e7c80dc102e718"
  - "srcrev_3d9437b79cfe5d1643b1e1b3c11b3514"
  - "srcrev_924e2760e63c41fc548ee69719d25ce5"
  - "srcrev_981bc9af1e2e3434f973caa5dafae568"
conflict_state: "none"
---

# Spoken Transcripts Storage Decision

## Summary

The team decided to use AWS S3 over Cloudflare R2 for storing spoken transcripts. S3 offers better throughput scaling and more features, while egress costs are currently manageable because users rarely download others' data. Future video storage may require cost reassessment.

## Claims

- The current storage destination for spoken transcripts is AWS S3, chosen over Cloudflare R2. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4cdf8b5e3ca9324953c8a224c3ed9989` `source_revision_id=srcrev_02924860b195c0990ba32819c39dc59e` `chunk_id=srcchunk_c93e02e3c1dbcf379f5b01109425579d` `native_locator=slack:C0AL7EKNHDF:1777058713.062729:1777061635.725019` `source_timestamp=2026-04-24T20:13:55Z`
- S3 was selected because it scales better with throughput, offers more features, and because R2's egress advantage is negligible since users rarely download others' uploaded data. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4cdf8b5e3ca9324953c8a224c3ed9989` `source_revision_id=srcrev_02924860b195c0990ba32819c39dc59e` `chunk_id=srcchunk_c93e02e3c1dbcf379f5b01109425579d` `native_locator=slack:C0AL7EKNHDF:1777058713.062729:1777061635.725019` `source_timestamp=2026-04-24T20:13:55Z`
- Upload of spoken transcripts is synchronous; users must complete the upload to the API before the operation is considered complete. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4cdf8b5e3ca9324953c8a224c3ed9989` `source_revision_id=srcrev_924e2760e63c41fc548ee69719d25ce5` `chunk_id=srcchunk_0decc871cb7d7c860ad58a20f0eda37c` `native_locator=slack:C0AL7EKNHDF:1777058713.062729:1777062561.606359` `source_timestamp=2026-04-24T20:29:21Z`
  - citation: `source_document_id=srcdoc_4cdf8b5e3ca9324953c8a224c3ed9989` `source_revision_id=srcrev_28e02c00ce34503b66e7c80dc102e718` `chunk_id=srcchunk_c34d9f3fa9a765093dab66ef1e8c731f` `native_locator=slack:C0AL7EKNHDF:1777058713.062729:1777062569.298149` `source_timestamp=2026-04-24T20:29:29Z`
- The chosen storage layer must support high concurrency for uploads within short time windows. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4cdf8b5e3ca9324953c8a224c3ed9989` `source_revision_id=srcrev_981bc9af1e2e3434f973caa5dafae568` `chunk_id=srcchunk_29f62e6b5d9262d594fee61dcd8579b7` `native_locator=slack:C0AL7EKNHDF:1777058713.062729:1777062592.723499` `source_timestamp=2026-04-24T20:30:04Z`
- Egress costs from AWS S3 to Cloudflare R2 for the Poseidon validation pipeline are estimated at $0.09 per GB, with the first 100 GB/month free. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4cdf8b5e3ca9324953c8a224c3ed9989` `source_revision_id=srcrev_3d9437b79cfe5d1643b1e1b3c11b3514` `chunk_id=srcchunk_b7cafbf8a9f4165afbc51796c27fdf4e` `native_locator=slack:C0AL7EKNHDF:1777058713.062729:1777062487.614369` `source_timestamp=2026-04-24T20:28:07Z`
- If video transcripts are introduced in the future, S3 egress costs will become a significant concern and may require revisiting the storage decision. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4cdf8b5e3ca9324953c8a224c3ed9989` `source_revision_id=srcrev_3d9437b79cfe5d1643b1e1b3c11b3514` `chunk_id=srcchunk_b7cafbf8a9f4165afbc51796c27fdf4e` `native_locator=slack:C0AL7EKNHDF:1777058713.062729:1777062487.614369` `source_timestamp=2026-04-24T20:28:07Z`

## Open Questions

- Is 100 GB/month free egress enough for the validation pipeline's audio-only workload?
- What will be the actual egress cost impact when video storage is added?
- Will AWS S3 be sufficiently reliable under high-concurrency upload scenarios?

## Sources

- `source_document_id`: `srcdoc_4cdf8b5e3ca9324953c8a224c3ed9989`
- `source_revision_id`: `srcrev_981bc9af1e2e3434f973caa5dafae568`
