---
title: "Audio Capture Initiative"
type: "project"
slug: "projects/audio-capture-initiative"
freshness: "2026-03-11T16:07:39Z"
tags:
  - "audio"
  - "data-collection"
  - "depin"
owners: []
source_revision_ids:
  - "srcrev_610c26f0766ccac93df4c13ced85632b"
  - "srcrev_6209ff02d6bdef33d036aabb6e9a73bb"
  - "srcrev_c09823bd004315bb6af1022b46e2d41c"
conflict_state: "none"
---

# Audio Capture Initiative

## Summary

Initiative to collect audio data for speech-to-speech model training. A framework for scoping the audio capture was shared on 2026-03-11. The team identified multiple risks during discussion.

## Claims

- A framework for audio capture was shared via an interactive HTML page. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_6209ff02d6bdef33d036aabb6e9a73bb` `chunk_id=srcchunk_f1ad8ab26b9927cc353f0605b05d7e4a` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773242050.279829` `source_timestamp=2026-03-11T15:14:10Z`
- Buyer trust is fragile: giving them bad data can make the team look amateurish, and Season 1 already spent one of those chances. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_c09823bd004315bb6af1022b46e2d41c` `chunk_id=srcchunk_17af7749d6fa44827fb07770e395f0b8` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773244679.324479` `source_timestamp=2026-03-11T15:57:59Z`
- Requirements will change: treating early input as a spec is risky because requirements can shift mid-conversation. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_c09823bd004315bb6af1022b46e2d41c` `chunk_id=srcchunk_17af7749d6fa44827fb07770e395f0b8` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773244679.324479` `source_timestamp=2026-03-11T15:57:59Z`
- Multi-speaker unscripted collection has unknown pitfalls; Sarick couldn't answer directly and said he'd think on it. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_c09823bd004315bb6af1022b46e2d41c` `chunk_id=srcchunk_17af7749d6fa44827fb07770e395f0b8` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773244679.324479` `source_timestamp=2026-03-11T15:57:59Z`
- Scripted data can be gamed or cheated by contributors; unscripted reduces this but introduces its own quality verification problems. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_c09823bd004315bb6af1022b46e2d41c` `chunk_id=srcchunk_17af7749d6fa44827fb07770e395f0b8` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773244679.324479` `source_timestamp=2026-03-11T15:57:59Z`
- AI-as-counterparty (human talks to an AI agent) validity for speech-to-speech model training is unconfirmed: Sarick said it's fine but nobody confirmed if it actually produces acceptable data. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_c09823bd004315bb6af1022b46e2d41c` `chunk_id=srcchunk_17af7749d6fa44827fb07770e395f0b8` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773244679.324479` `source_timestamp=2026-03-11T15:57:59Z`
- There is a chicken-and-egg problem with sales: buyers first ask what data exists, but you can't sell what doesn't exist yet, so collection risk is taken before demand is confirmed. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_c09823bd004315bb6af1022b46e2d41c` `chunk_id=srcchunk_17af7749d6fa44827fb07770e395f0b8` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773244679.324479` `source_timestamp=2026-03-11T15:57:59Z`
- No pre-upload quality gate exists yet: bad submissions might be processed and paid for before rejection; Poseidon may be building this but it's not confirmed. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_c09823bd004315bb6af1022b46e2d41c` `chunk_id=srcchunk_17af7749d6fa44827fb07770e395f0b8` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773244679.324479` `source_timestamp=2026-03-11T15:57:59Z`
- There is a communication/PR risk because DePIN season 1 participants were left hanging; any new data capture initiative needs marketing consultation. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fc15899344232501c257a48462da02af` `source_revision_id=srcrev_610c26f0766ccac93df4c13ced85632b` `chunk_id=srcchunk_f8563c47bbc74304bde4bc8f32b91359` `native_locator=slack:C0AL7EKNHDF:1773242050.279829:1773245259.968919` `source_timestamp=2026-03-11T16:07:39Z`

## Open Questions

- How to address buyer trust given Season 1 failures?
- How to communicate the new initiative to DePIN Season 1 participants?
- How to handle changing requirements without building on stale specs?
- How to prevent gaming in scripted/unscripted data collection?
- How to resolve the chicken-and-egg sales problem?
- How to verify multi-speaker unscripted audio quality?
- Whether a pre-upload quality gate will be implemented by Poseidon?
- Whether AI-as-counterparty produces acceptable training data?

## Sources

- `source_document_id`: `srcdoc_fc15899344232501c257a48462da02af`
- `source_revision_id`: `srcrev_37342dec31b14ebc712d7dba3216f177`
