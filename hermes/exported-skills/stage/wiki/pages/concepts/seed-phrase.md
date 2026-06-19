---
title: "Seed Phrase"
type: "concept"
slug: "concepts/seed-phrase"
freshness: "2026-04-28T00:16:47Z"
tags:
  - "numo"
  - "seed-phrase"
  - "voice-verification"
owners: []
source_revision_ids:
  - "srcrev_a947b543513da110ef6ef6d1537aeccb"
  - "srcrev_d030057c90d058a90d111428d264a7bf"
  - "srcrev_d162a20e60ad9a81c98fa4deb89e061d"
  - "srcrev_d682c23490ab50fdb404292171351640"
conflict_state: "none"
---

# Seed Phrase

## Summary

Seed phrase is an audio recording of 10-20 seconds used for voice verification in Numo, required for every account. Numo is responsible for generating seed phrases, with implementation expected in 1-2 weeks.

## Claims

- Every account must have a seed phrase to confirm the speaker is always the same person (voice verification). `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0e93185e465afd76dbb8b41bb680304c` `source_revision_id=srcrev_a947b543513da110ef6ef6d1537aeccb` `chunk_id=srcchunk_6a19edeff3c7fc20991f11750fb42edb` `native_locator=slack:C0AL7EKNHDF:1776806068.899969:1776809132.934049` `source_timestamp=2026-04-21T22:09:37Z`
  - citation: `source_document_id=srcdoc_0e93185e465afd76dbb8b41bb680304c` `source_revision_id=srcrev_d162a20e60ad9a81c98fa4deb89e061d` `chunk_id=srcchunk_766cf5d5c23a0adcdff5d979164e8457` `native_locator=slack:C0AL7EKNHDF:1776806068.899969:1777332164.686549` `source_timestamp=2026-04-27T23:22:44Z`
- Numo will generate the seed phrases; they can be a small set of transcripts reused across users. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0e93185e465afd76dbb8b41bb680304c` `source_revision_id=srcrev_d030057c90d058a90d111428d264a7bf` `chunk_id=srcchunk_17d34e229d861ef953fd620a561f418d` `native_locator=slack:C0AL7EKNHDF:1776806068.899969:1776809137.583899` `source_timestamp=2026-04-21T22:08:19Z`
- Implementation of seed phrase in the app is expected in 1-2 weeks, after a launch focus. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_0e93185e465afd76dbb8b41bb680304c` `source_revision_id=srcrev_d682c23490ab50fdb404292171351640` `chunk_id=srcchunk_931203e220af74ec841525326c443721` `native_locator=slack:C0AL7EKNHDF:1776806068.899969:1777335407.503829` `source_timestamp=2026-04-28T00:16:47Z`

## Open Questions

- How should seed phrases be generated for multilingual support? Options: generate random words directly in each language, or generate in English and translate. Should the phrases be random words or full sentences? (from srcchunk_931203e220af74ec841525326c443721)

## Related Pages

- `poseidon-numo-product-alignment`

## Sources

- `source_document_id`: `srcdoc_0e93185e465afd76dbb8b41bb680304c`
- `source_revision_id`: `srcrev_d682c23490ab50fdb404292171351640`
