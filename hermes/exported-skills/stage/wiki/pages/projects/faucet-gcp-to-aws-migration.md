---
title: "Faucet GCP-to-AWS Migration"
type: "project"
slug: "projects/faucet-gcp-to-aws-migration"
freshness: "2026-02-20T20:09:06Z"
tags:
  - "aws"
  - "faucet"
  - "gcp"
  - "migration"
owners:
  - "Haodi"
  - "U04L0DD6B6F"
  - "U05U5HQAFFH"
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
  - "srcrev_f60985285ce4e96b225afe4e56f53047"
conflict_state: "none"
---

# Faucet GCP-to-AWS Migration

## Summary

Ongoing project to migrate the company faucet service from GCP to AWS due to cost concerns. Includes backend and handler migration, with observed frontend issues.

## Claims

- The migration of the faucet service from GCP to AWS is actively in progress. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_b84f812fe5de6ed9ac19c1f081c2f173` `chunk_id=srcchunk_276d75c40f3eeb79e856a3be4c0510ce` `native_locator=slack:C04T5307FNU:1771556344.494389:1771561440.365729` `source_timestamp=2026-02-20T04:24:00Z`
- The GCP faucet was operational during migration preparations. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_cb6cccae2b00cbc3ca858b2becbec1b8` `chunk_id=srcchunk_be0d68d4b97d465c6b904997a9625e6e` `native_locator=slack:C04T5307FNU:1771556344.494389:1771556462.278939` `source_timestamp=2026-02-20T03:01:02Z`
- A decision was made to discontinue use of the GCP faucet to avoid high fees. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_3b2a66c6a3a876a66b86acf2af4cd928` `chunk_id=srcchunk_87aa13e389b8d98a9d944a72c2901273` `native_locator=slack:C04T5307FNU:1771556344.494389:1771561003.223299` `source_timestamp=2026-02-20T04:16:43Z`
- Backend migration requires support from U0A33S7AM1Q and U0643B9PPC1. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_17dfe4f41bb88f65d270e4af09ac7ba1` `chunk_id=srcchunk_8008204f39a57935d28562e7aa63c37e` `native_locator=slack:C04T5307FNU:1771556344.494389:1771574938.185849` `source_timestamp=2026-02-20T08:08:58Z`
- A handler was migrated but still needs confirmation. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_697e559e04a02b51e7c71503b64dedcf` `chunk_id=srcchunk_aaa597bee525f52a7ace32c4a6288698` `native_locator=slack:C04T5307FNU:1771556344.494389:1771575417.299939` `source_timestamp=2026-02-20T08:19:40Z`
- Instructions for the backend migration were provided by Haodi. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_2488e9f26c19474a1c2a2d1d98a5dbb7` `chunk_id=srcchunk_4b0e3190ca7d323daaf9a9fa9c430c74` `native_locator=slack:C04T5307FNU:1771556344.494389:1771618146.706529` `source_timestamp=2026-02-20T20:09:06Z`
- Frontend errors were reported during the migration process. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_f60985285ce4e96b225afe4e56f53047` `chunk_id=srcchunk_82d8ffa6e1136ffd1ad600a780fed088` `native_locator=slack:C04T5307FNU:1771556344.494389:1771556344.494389` `source_timestamp=2026-02-20T03:02:48Z`
  - citation: `source_document_id=srcdoc_31a4c2cbfbb31aaba80b93569fb7228a` `source_revision_id=srcrev_e3ffac048a482a1d8e299d41c2769832` `chunk_id=srcchunk_41e4eee03e6903cfc9d34fa5d7b47df9` `native_locator=slack:C04T5307FNU:1771556344.494389:1771556613.862529` `source_timestamp=2026-02-20T03:03:33Z`

## Sources

- `source_document_id`: `srcdoc_31a4c2cbfbb31aaba80b93569fb7228a`
- `source_revision_id`: `srcrev_e1fb7bd54e23bde7f6afa33854f20666`
