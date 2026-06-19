---
title: "Numo Contribution Receipt Schema and Licensing Decision"
type: "decision"
slug: "decisions/numo-contribution-receipt-schema"
freshness: "2026-04-21T20:57:05Z"
tags:
  - "data-receipts"
  - "legal"
  - "licensing"
  - "metadata"
  - "numo"
  - "tos"
  - "voice-recording"
owners:
  - "Avi"
  - "Legal"
  - "Romain (author)"
source_revision_ids:
  - "srcrev_092fdb255d9a2cb4fdb7cdd8b3a62c46"
  - "srcrev_69a04fe4d8793b2c9c6aedb1448cc661"
  - "srcrev_69facca3c54a3385659e427b2fb20115"
  - "srcrev_8354839c29cd317c6803708a5346cfc7"
  - "srcrev_d643903300a87722636e3de4b2c2b66d"
  - "srcrev_d670c4191a92dfa2cdee1d03fb30b635"
  - "srcrev_de6366e98d08aae9dd660b6c2da16ab3"
  - "srcrev_fd34ce0e42792427e084aa6e6bae7674"
conflict_state: "none"
---

# Numo Contribution Receipt Schema and Licensing Decision

## Summary

Decided against using Story Protocol's PIL for voice contribution receipts, opting for ToS-backed metadata with consent signatures, IPFS storage for ToS, and S3 for receipt metadata. Jurisdiction field removed. Revocation ledger deferred as non-P0.

## Claims

- PIL (Programmable IP License) is not used for Numo voice contributions. The governing agreement is Terms of Service + Privacy Policy, not a creative license, and forcing PIL would create legal ambiguity without enforcement value. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_fd34ce0e42792427e084aa6e6bae7674` `chunk_id=srcchunk_e648d507f217ee857fe123c21f30f2fc` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776785085.047459` `source_timestamp=2026-04-21T15:24:45Z`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_69facca3c54a3385659e427b2fb20115` `chunk_id=srcchunk_94712ed7c92a8bafc8b450e3517edc77` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776692900.676999` `source_timestamp=2026-04-20T13:48:20Z`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_69a04fe4d8793b2c9c6aedb1448cc661` `chunk_id=srcchunk_5cacec9652337a459c2f7cbc6485a149` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776718656.470279` `source_timestamp=2026-04-20T20:57:36Z`
- Raw Terms of Service documents will be stored on IPFS via Pinata (one-time write per version, re-pinnable). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_fd34ce0e42792427e084aa6e6bae7674` `chunk_id=srcchunk_e648d507f217ee857fe123c21f30f2fc` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776785085.047459` `source_timestamp=2026-04-21T15:24:45Z`
- Per-contribution metadata will be stored on S3. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_fd34ce0e42792427e084aa6e6bae7674` `chunk_id=srcchunk_e648d507f217ee857fe123c21f30f2fc` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776785085.047459` `source_timestamp=2026-04-21T15:24:45Z`
- Per-contribution receipt includes metadata fields: terms_version, terms_hash, terms_cid (IPFS), consent_signature, and a structured usage_grant enum. Jurisdiction (iso code) was initially considered but later removed. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_fd34ce0e42792427e084aa6e6bae7674` `chunk_id=srcchunk_e648d507f217ee857fe123c21f30f2fc` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776785085.047459` `source_timestamp=2026-04-21T15:24:45Z`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_d643903300a87722636e3de4b2c2b66d` `chunk_id=srcchunk_9323af1a142f8ce9a4176d11b2175095` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776804881.723869` `source_timestamp=2026-04-21T20:54:41Z`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_de6366e98d08aae9dd660b6c2da16ab3` `chunk_id=srcchunk_30380499eafd4ec5256d0d7ee2a61cee` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776801503.934829` `source_timestamp=2026-04-21T19:58:23Z`
- A revocation ledger via a separate mutable pointer contract is planned (non-P0) to honor GDPR/CCPA withdrawal rights without breaking consent record immutability. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_fd34ce0e42792427e084aa6e6bae7674` `chunk_id=srcchunk_e648d507f217ee857fe123c21f30f2fc` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776785085.047459` `source_timestamp=2026-04-21T15:24:45Z`
- Consent signature for data grants will be obtained silently in the background for phase 1. For ToS changes, a new click-wrap consent will be required. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_8354839c29cd317c6803708a5346cfc7` `chunk_id=srcchunk_0c5869260310f00e3bb7d294d850a916` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776796326.088439` `source_timestamp=2026-04-21T18:32:06Z`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_092fdb255d9a2cb4fdb7cdd8b3a62c46` `chunk_id=srcchunk_8f154c97be9b12cce0183ca66b41309c` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776799721.606529` `source_timestamp=2026-04-21T19:28:41Z`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_d670c4191a92dfa2cdee1d03fb30b635` `chunk_id=srcchunk_a79f49a465c68177a09c56633ddb3206` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776805025.736639` `source_timestamp=2026-04-21T20:57:05Z`
- The final schema and decision are documented at: https://www.notion.so/storyprotocol/Numo-Contribution-Receipt-Schema-Attribution-Licensing-Layer-no-PIL-349051299a548176a437c292014dc79e `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76863347daf355bd7d746d5bd8fbbc58` `source_revision_id=srcrev_fd34ce0e42792427e084aa6e6bae7674` `chunk_id=srcchunk_e648d507f217ee857fe123c21f30f2fc` `native_locator=slack:C0AL7EKNHDF:1776445105.677419:1776785085.047459` `source_timestamp=2026-04-21T15:24:45Z`

## Open Questions

- Exact design and timeline for the revocation ledger (non-P0).
- How ToS versioning will be handled if changes occur later.

## Sources

- `source_document_id`: `srcdoc_76863347daf355bd7d746d5bd8fbbc58`
- `source_revision_id`: `srcrev_4d9cc5c7b8eae98b5b4225ef7250482b`
