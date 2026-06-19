---
title: "Poseidon Multiplier Attribution"
type: "decision"
slug: "decisions/poseidon-multiplier-attribution"
freshness: "2026-04-22T16:49:43Z"
tags:
  - "attribution"
  - "multiplier"
  - "onboarding"
  - "poseidon"
owners: []
source_revision_ids:
  - "srcrev_010bf8f67e4ef9c4b1bec5895225e69c"
  - "srcrev_a587a7e7cf7c7208bd702171be3bad37"
  - "srcrev_b9804d4a092477772048c8002965a195"
conflict_state: "none"
---

# Poseidon Multiplier Attribution

## Summary

Decision to automatically attribute Poseidon user multipliers using an email whitelist, with login page reminders.

## Claims

- Poseidon user multipliers will be applied automatically if a user signs up with the same email address that appears in a pre-collected list of Poseidon user emails. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_635c6d6ac5f1d51a55ef312f9b2e33a5` `source_revision_id=srcrev_010bf8f67e4ef9c4b1bec5895225e69c` `chunk_id=srcchunk_14af04bea374b47cdd59428b3da23a47` `native_locator=slack:C0AL7EKNHDF:1776875283.522869:1776876583.772829` `source_timestamp=2026-04-22T16:49:43Z`
- The login page may display instructions to existing Poseidon users to sign up with the same email address they used for Poseidon in order to receive the multiplier. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_635c6d6ac5f1d51a55ef312f9b2e33a5` `source_revision_id=srcrev_010bf8f67e4ef9c4b1bec5895225e69c` `chunk_id=srcchunk_14af04bea374b47cdd59428b3da23a47` `native_locator=slack:C0AL7EKNHDF:1776875283.522869:1776876583.772829` `source_timestamp=2026-04-22T16:49:43Z`
- An allowlist/whitelist-based mechanism was suggested to attribute multipliers to Poseidon users automatically. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_635c6d6ac5f1d51a55ef312f9b2e33a5` `source_revision_id=srcrev_b9804d4a092477772048c8002965a195` `chunk_id=srcchunk_efe68f5a3e6002152e400bdfab000b1c` `native_locator=slack:C0AL7EKNHDF:1776875283.522869:1776875476.234579` `source_timestamp=2026-04-22T16:31:16Z`
- There is a risk of confusion if Poseidon users sign up before the multiplier announcement and subsequently do not receive the multiplier unless attribution logic is in place. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_635c6d6ac5f1d51a55ef312f9b2e33a5` `source_revision_id=srcrev_a587a7e7cf7c7208bd702171be3bad37` `chunk_id=srcchunk_c9262231a58be6db4d30108fc9e72260` `native_locator=slack:C0AL7EKNHDF:1776875283.522869:1776875283.522869` `source_timestamp=2026-04-22T16:28:03Z`

## Open Questions

- How will multipliers be handled for Poseidon users who sign up before the email whitelist is implemented?

## Sources

- `source_document_id`: `srcdoc_635c6d6ac5f1d51a55ef312f9b2e33a5`
- `source_revision_id`: `srcrev_98b5bf48aaf2164acbad225e2401b410`
