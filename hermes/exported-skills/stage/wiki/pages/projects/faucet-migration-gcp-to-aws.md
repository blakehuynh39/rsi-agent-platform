---
title: "Faucet Migration from GCP to AWS"
type: "project"
slug: "projects/faucet-migration-gcp-to-aws"
freshness: "2026-02-20T20:09:06Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners:
  - "U0643B9PPC1"
  - "U07TNT9N4JC"
  - "U0A33S7AM1Q"
source_revision_ids:
  - "srcrev_17dfe4f41bb88f65d270e4af09ac7ba1"
  - "srcrev_2488e9f26c19474a1c2a2d1d98a5dbb7"
  - "srcrev_3b2a66c6a3a876a66b86acf2af4cd928"
  - "srcrev_697e559e04a02b51e7c71503b64dedcf"
  - "srcrev_b84f812fe5de6ed9ac19c1f081c2f173"
  - "srcrev_cb6cccae2b00cbc3ca858b2becbec1b8"
  - "srcrev_e3ffac048a482a1d8e299d41c2769832"
conflict_state: "none"
---

# Faucet Migration from GCP to AWS

## Summary

This project migrates the faucet server from Google Cloud Platform to Amazon Web Services to reduce costs. The migration involves backend and frontend changes. Issues have been encountered with the frontend and the decision was made to stop using GCP faucet to avoid high fees.

## Claims

- The faucet server is being migrated from GCP to AWS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_b84f812fe5de6ed9ac19c1f081c2f173` `chunk_id=srcchunk_276d75c40f3eeb79e856a3be4c0510ce` `native_locator=slack:C04T5307FNU:1771556344.494389:1771561440.365729` `source_timestamp=2026-02-20T04:24:00Z`
- The GCP faucet works, but the AWS endpoint has not been changed yet. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_cb6cccae2b00cbc3ca858b2becbec1b8` `chunk_id=srcchunk_be0d68d4b97d465c6b904997a9625e6e` `native_locator=slack:C04T5307FNU:1771556344.494389:1771556462.278939` `source_timestamp=2026-02-20T03:01:02Z`
- Users are seeing errors on the frontend when using the faucet. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_e3ffac048a482a1d8e299d41c2769832` `chunk_id=srcchunk_41e4eee03e6903cfc9d34fa5d7b47df9` `native_locator=slack:C04T5307FNU:1771556344.494389:1771556613.862529` `source_timestamp=2026-02-20T03:03:33Z`
- Decision made to stop using GCP faucet to avoid high fees. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_3b2a66c6a3a876a66b86acf2af4cd928` `chunk_id=srcchunk_87aa13e389b8d98a9d944a72c2901273` `native_locator=slack:C04T5307FNU:1771556344.494389:1771561003.223299` `source_timestamp=2026-02-20T04:16:43Z`
- Backend migration requires support from U0A33S7AM1Q and U0643B9PPC1. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_17dfe4f41bb88f65d270e4af09ac7ba1` `chunk_id=srcchunk_8008204f39a57935d28562e7aa63c37e` `native_locator=slack:C04T5307FNU:1771556344.494389:1771574938.185849` `source_timestamp=2026-02-20T08:08:58Z`
- Handler migration completed but needs confirmation. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_697e559e04a02b51e7c71503b64dedcf` `chunk_id=srcchunk_aaa597bee525f52a7ace32c4a6288698` `native_locator=slack:C04T5307FNU:1771556344.494389:1771575417.299939` `source_timestamp=2026-02-20T08:19:40Z`
- Instructions for backend migration were provided by Haodi. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_2488e9f26c19474a1c2a2d1d98a5dbb7` `chunk_id=srcchunk_4b0e3190ca7d323daaf9a9fa9c430c74` `native_locator=slack:C04T5307FNU:1771556344.494389:1771618146.706529` `source_timestamp=2026-02-20T20:09:06Z`

## Open Questions

- Has the frontend error been resolved?
- Is the migration to AWS complete?
- What is the status of backend migration support?

## Sources

- `source_document_id`: `srcdoc_31a4c2cbfbb31aaba80b93569fb7228a`
- `source_revision_id`: `srcrev_b84f812fe5de6ed9ac19c1f081c2f173`
