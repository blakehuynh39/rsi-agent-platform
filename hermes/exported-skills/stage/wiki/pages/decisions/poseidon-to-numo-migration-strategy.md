---
title: "Poseidon-to-Numo Migration Strategy"
type: "decision"
slug: "decisions/poseidon-to-numo-migration-strategy"
freshness: "2026-04-08T19:00:01Z"
tags:
  - "migration"
  - "numo"
  - "poseidon"
  - "rewards"
  - "season-1"
owners:
  - "Ben"
  - "Gardo"
  - "Yash"
source_revision_ids:
  - "srcrev_2c97b942be0a1bb62582d7404b08f565"
conflict_state: "none"
---

# Poseidon-to-Numo Migration Strategy

## Summary

Strategy for migrating 83k Poseidon Season 1 users to Numo using a hybrid approach of multipliers based on contribution quality, with geographic and language constraints.

## Claims

- Three options were discussed for migrating Poseidon users: Option 1: Transfer points directly to numo system; Option 2: Convert contributions to multiplier-based rewards tied to new activity; Option 3: Convert points to boosts/multipliers without direct cash equivalents. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_2c97b942be0a1bb62582d7404b08f565` `chunk_id=srcchunk_2ce39177113791595f85a871fe9520a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775674801.622609` `source_timestamp=2026-04-08T19:00:01Z`
- Team preference is a hybrid Option 2/3 approach with quality-based tiering using similarity scores (threshold 65), multiplier system ranging 1.2x-5x, rewards tied to new contributions, and a minimum payout of $25 to manage KYC/processing costs. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_2c97b942be0a1bb62582d7404b08f565` `chunk_id=srcchunk_2ce39177113791595f85a871fe9520a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775674801.622609` `source_timestamp=2026-04-08T19:00:01Z`
- Numo will launch in limited countries versus Poseidon's 16-language global reach; Numo supports only 5 languages, creating geographic and language compliance restrictions. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_2c97b942be0a1bb62582d7404b08f565` `chunk_id=srcchunk_2ce39177113791595f85a871fe9520a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775674801.622609` `source_timestamp=2026-04-08T19:00:01Z`
- Quality filtering will use Yash's similarity score data to tier users, targeting ~3,000 quality contributors from the 83,000 total Poseidon users, avoiding payouts to low-quality/spam contributors. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_2c97b942be0a1bb62582d7404b08f565` `chunk_id=srcchunk_2ce39177113791595f85a871fe9520a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775674801.622609` `source_timestamp=2026-04-08T19:00:01Z`
- Numo will not have a points system; only cash estimates. Leaderboard will show contribution count, not points. Credit page must display two numbers: total possible vs validated earnings. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_2c97b942be0a1bb62582d7404b08f565` `chunk_id=srcchunk_2ce39177113791595f85a871fe9520a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775674801.622609` `source_timestamp=2026-04-08T19:00:01Z`
- Next steps: Gardo to discuss jurisdiction expansion with legal (tomorrow noon Pacific), team to model budget scenarios with tiers/multipliers, get Ben/leadership approval, Yash to provide histogram of user quality scores, and finalize strategy by Friday. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_2c97b942be0a1bb62582d7404b08f565` `chunk_id=srcchunk_2ce39177113791595f85a871fe9520a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775674801.622609` `source_timestamp=2026-04-08T19:00:01Z`

## Open Questions

- How to handle users not initially eligible due to geographic/language constraints?
- What is the exact budget and how will multipliers be assigned?

## Related Pages

- `numo-user-reward-formula`

## Sources

- `source_document_id`: `srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f`
- `source_revision_id`: `srcrev_0ff7de8b3737d513b372fae52f3397c1`
