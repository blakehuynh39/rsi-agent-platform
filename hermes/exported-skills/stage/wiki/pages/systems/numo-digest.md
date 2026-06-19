---
title: "Numo Digest"
type: "system"
slug: "systems/numo-digest"
freshness: "2026-04-30T17:19:40Z"
tags:
  - "digest"
  - "incident"
  - "monitoring"
  - "numo"
owners: []
source_revision_ids:
  - "srcrev_c22aeaa518af4c12d4f182c318b89963"
  - "srcrev_d0d3cfe5adfbbe6a009e5db92cfa12b4"
conflict_state: "none"
---

# Numo Digest

## Summary

A system that generates a daily digest, which experienced a failure that was fixed. It is scheduled to run at 9AM PST and 4PM PST, and the team can provide feedback on useful metrics.

## Claims

- Numo digest failed on a previous run. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4c08574d791f88eed34dadac39911568` `source_revision_id=srcrev_d0d3cfe5adfbbe6a009e5db92cfa12b4` `chunk_id=srcchunk_f7bf32588bcf0b2e1f4653a65f3cf1f3` `native_locator=slack:C0AL7EKNHDF:1777566566.938799:1777566566.938799` `source_timestamp=2026-04-30T16:29:26Z`
- The Numo digest failure was fixed. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4c08574d791f88eed34dadac39911568` `source_revision_id=srcrev_c22aeaa518af4c12d4f182c318b89963` `chunk_id=srcchunk_af23c7e494c8c70c715878f459ededa7` `native_locator=slack:C0AL7EKNHDF:1777566566.938799:1777569580.193119` `source_timestamp=2026-04-30T17:19:40Z`
- The Numo digest is scheduled to run daily at 9AM PST and 4PM PST. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4c08574d791f88eed34dadac39911568` `source_revision_id=srcrev_c22aeaa518af4c12d4f182c318b89963` `chunk_id=srcchunk_af23c7e494c8c70c715878f459ededa7` `native_locator=slack:C0AL7EKNHDF:1777566566.938799:1777569580.193119` `source_timestamp=2026-04-30T17:19:40Z`
- Team members are encouraged to suggest metrics to be surfaced in the digest. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_4c08574d791f88eed34dadac39911568` `source_revision_id=srcrev_c22aeaa518af4c12d4f182c318b89963` `chunk_id=srcchunk_af23c7e494c8c70c715878f459ededa7` `native_locator=slack:C0AL7EKNHDF:1777566566.938799:1777569580.193119` `source_timestamp=2026-04-30T17:19:40Z`

## Open Questions

- What metrics are currently being surfaced?
- What was the root cause of the digest failure?
- Which team members own the digest system?

## Sources

- `source_document_id`: `srcdoc_4c08574d791f88eed34dadac39911568`
- `source_revision_id`: `srcrev_c22aeaa518af4c12d4f182c318b89963`
