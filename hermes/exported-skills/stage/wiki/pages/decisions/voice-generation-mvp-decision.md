---
title: "Voice Generation MVP: Eleven Labs Selection"
type: "decision"
slug: "decisions/voice-generation-mvp-decision"
freshness: "2026-03-20T19:15:46Z"
tags:
  - "architecture"
  - "cost-optimization"
  - "eleven-labs"
  - "mvp"
  - "voice-generation"
owners: []
source_revision_ids:
  - "srcrev_027564746930f7c50e69ab7524b3e11d"
  - "srcrev_7577ad5347dc5cc34f0814f6a18a0d6b"
conflict_state: "none"
---

# Voice Generation MVP: Eleven Labs Selection

## Summary

The team decided to use Eleven Labs for the MVP voice generation, with a POC for custom infrastructure as backup, given the April 10 deadline. A free-tier client-side LLM alternative was considered but not chosen.

## Claims

- The possibility of using an LLM with a free tier accessible via client-side browser API calls was raised to avoid costs and throttle limits. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_aac61335760a55e511b7391f578d90e6` `source_revision_id=srcrev_7577ad5347dc5cc34f0814f6a18a0d6b` `chunk_id=srcchunk_51d23e4164c25f47df587c42b286193a` `native_locator=slack:C0AL7EKNHDF:1774024893.497499:1774024893.497499` `source_timestamp=2026-03-20T16:41:33Z`
- The MVP will start using Eleven Labs. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_aac61335760a55e511b7391f578d90e6` `source_revision_id=srcrev_027564746930f7c50e69ab7524b3e11d` `chunk_id=srcchunk_41fb699c08413b91ead1a4c377235f4f` `native_locator=slack:C0AL7EKNHDF:1774024893.497499:1774034146.989219` `source_timestamp=2026-03-20T19:15:46Z`
- A proof of concept (POC) with own infrastructure for scaling has been demonstrated. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_aac61335760a55e511b7391f578d90e6` `source_revision_id=srcrev_027564746930f7c50e69ab7524b3e11d` `chunk_id=srcchunk_41fb699c08413b91ead1a4c377235f4f` `native_locator=slack:C0AL7EKNHDF:1774024893.497499:1774034146.989219` `source_timestamp=2026-03-20T19:15:46Z`
- The project deadline is April 10 (4/10). `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_aac61335760a55e511b7391f578d90e6` `source_revision_id=srcrev_027564746930f7c50e69ab7524b3e11d` `chunk_id=srcchunk_41fb699c08413b91ead1a4c377235f4f` `native_locator=slack:C0AL7EKNHDF:1774024893.497499:1774034146.989219` `source_timestamp=2026-03-20T19:15:46Z`
- The team will proceed step by step given the upcoming deadline. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_aac61335760a55e511b7391f578d90e6` `source_revision_id=srcrev_027564746930f7c50e69ab7524b3e11d` `chunk_id=srcchunk_41fb699c08413b91ead1a4c377235f4f` `native_locator=slack:C0AL7EKNHDF:1774024893.497499:1774034146.989219` `source_timestamp=2026-03-20T19:15:46Z`

## Open Questions

- Will the team reconsider the free-tier client-side LLM approach if Eleven Labs costs become prohibitive at scale?

## Sources

- `source_document_id`: `srcdoc_aac61335760a55e511b7391f578d90e6`
- `source_revision_id`: `srcrev_027564746930f7c50e69ab7524b3e11d`
