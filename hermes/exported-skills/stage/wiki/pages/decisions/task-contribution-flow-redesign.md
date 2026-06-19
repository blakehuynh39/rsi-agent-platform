---
title: "Task Contribution Flow Redesign"
type: "decision"
slug: "decisions/task-contribution-flow-redesign"
freshness: "2026-04-27T11:21:52Z"
tags:
  - "contribution-flow"
  - "numo"
  - "redesign"
  - "ui"
  - "ux"
owners:
  - "U04L0DD6B6F"
  - "U04L0DD71TM"
  - "U05A515NBFC"
  - "U07MLSYUS5R"
  - "U08AGDT08E7"
source_revision_ids:
  - "srcrev_5ef481d47b239cfe8c7b496d29e1f5ba"
  - "srcrev_a26034883fd292062ed4addc440b7779"
conflict_state: "none"
---

# Task Contribution Flow Redesign

## Summary

Redesign of the task contribution flow to improve user experience and incentivize contributions.

## Claims

- Instructions and text block are merged, removing the blurred box and an extra step. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Task card image added to the screen for smoother transition. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Remaining daily contributions count displayed below 'Start Recording' button. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Nav bar removed on task screens (except success screen) to keep user focused. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Visual recording indicator added so users can see microphone is working. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Recording indicator text changes: 'Read for at least 20 seconds to get a reward' initially, then 'Press stop when finished' after 20 seconds. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Review screen elements grouped for visual hierarchy, with warning that low-quality contributions won't get payout. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Final screen: 'Contribute again' button made primary to encourage more contributions, visual emphasis adjusted. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Lex requested visual element suggestion (logo-style image) from U07MLSYUS5R for the final screen. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Design shared via Figma link. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_a26034883fd292062ed4addc440b7779` `chunk_id=srcchunk_91480973cc6fc17b2df8c281154788c9` `native_locator=slack:C0AL7EKNHDF:1777274247.125889:1777274247.125889` `source_timestamp=2026-04-27T07:17:27Z`
- Acknowledged by a team member to review shortly. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_79e891e93fed704f0d198e25156e20d4` `source_revision_id=srcrev_5ef481d47b239cfe8c7b496d29e1f5ba` `chunk_id=srcchunk_d380c24ae9db0b51034b3a9f273b6aae` `native_locator=slack:C0AL7EKNHDF:1777288912.688379` `source_timestamp=2026-04-27T11:21:52Z`

## Sources

- `source_document_id`: `srcdoc_79e891e93fed704f0d198e25156e20d4`
- `source_revision_id`: `srcrev_5ef481d47b239cfe8c7b496d29e1f5ba`
