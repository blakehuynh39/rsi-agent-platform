---
title: "Depin App Transcript Bucket"
type: "system"
slug: "systems/depin-app-transcript-bucket"
freshness: "2026-04-25T01:21:20Z"
tags:
  - "cloudflare-r2"
  - "data-pipeline"
  - "depin-app"
  - "indic-languages"
  - "transcripts"
owners:
  - "U04L0DD6B6F"
  - "U0A2D9U625V"
source_revision_ids:
  - "srcrev_81db517c736d40c2ab1542a2084a3f37"
  - "srcrev_ca7dea89a06838143348089fd0cf380a"
conflict_state: "none"
---

# Depin App Transcript Bucket

## Summary

Cloudflare R2 bucket containing 350,000 transcripts per language for the depin app, used as a source for multi-language seed phrases.

## Claims

- The bucket contains 350,000 transcripts per language. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_65608ac66a81c24d178af9ae24c0c180` `source_revision_id=srcrev_ca7dea89a06838143348089fd0cf380a` `chunk_id=srcchunk_de7eee17932b89b53df21d336f36ed9a` `native_locator=slack:C0AL7EKNHDF:1777061851.343039:1777061851.343039` `source_timestamp=2026-04-24T20:17:31Z`
- The bucket is located at a specific Cloudflare R2 URL. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_65608ac66a81c24d178af9ae24c0c180` `source_revision_id=srcrev_ca7dea89a06838143348089fd0cf380a` `chunk_id=srcchunk_de7eee17932b89b53df21d336f36ed9a` `native_locator=slack:C0AL7EKNHDF:1777061851.343039:1777061851.343039` `source_timestamp=2026-04-24T20:17:31Z`
- Transcripts were uploaded to production and staging environments. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_65608ac66a81c24d178af9ae24c0c180` `source_revision_id=srcrev_81db517c736d40c2ab1542a2084a3f37` `chunk_id=srcchunk_14f19da5b51d18df7c777ea621b25819` `native_locator=slack:C0AL7EKNHDF:1777061851.343039:1777080080.583449` `source_timestamp=2026-04-25T01:21:20Z`

## Open Questions

- How exactly will transcripts be used as seed phrases?
- What is the handoff plan for seedphrases?

## Related Pages

- `multi-language-seedphrase-reintroduction`

## Sources

- `source_document_id`: `srcdoc_65608ac66a81c24d178af9ae24c0c180`
- `source_revision_id`: `srcrev_81db517c736d40c2ab1542a2084a3f37`
