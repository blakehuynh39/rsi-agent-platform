---
title: "Numo User Reward Formula"
type: "system"
slug: "systems/numo-user-reward-formula"
freshness: "2026-04-08T21:33:08Z"
tags:
  - "calculation"
  - "formula"
  - "multiplier"
  - "numo"
  - "rewards"
owners:
  - "Numo Team"
source_revision_ids:
  - "srcrev_06d81fb719957f3f2378a8d9b2487856"
  - "srcrev_0851fb719899954133cfaad7e103521c"
  - "srcrev_184710a7640863c66098697c3974fa70"
  - "srcrev_43d72f4fa6e027dcdd67aeec92ecccdb"
  - "srcrev_60aa0731ad6db323830c0faea10b1dfe"
conflict_state: "none"
---

# Numo User Reward Formula

## Summary

Proposed formula for calculating user rewards in Numo: Reward = C * p * m, with pending and settled balances and a $25 minimum withdrawal threshold.

## Claims

- The basic reward formula is: User Reward per task = C * p * m, where C = number of contributions of a particular campaign, p = coefficient for campaign task completion to balance (may differ per campaign), m = user multiplier (e.g., from Poseidon segmentation). `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_43d72f4fa6e027dcdd67aeec92ecccdb` `chunk_id=srcchunk_3c27dc4b62decd1c28fcf4a01fe620a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775676631.557519` `source_timestamp=2026-04-08T19:47:50Z`
- Different campaigns will have different payout coefficients; for example, an English audio task might be $0.05, a video task might be $0.08. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_43d72f4fa6e027dcdd67aeec92ecccdb` `chunk_id=srcchunk_3c27dc4b62decd1c28fcf4a01fe620a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775676631.557519` `source_timestamp=2026-04-08T19:47:50Z`
- The system will display a pending balance (total contributions * p * m) and a settled balance (verified contributions * p * m), with a $25 minimum withdrawal threshold and a progress bar showing remaining amount needed. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_43d72f4fa6e027dcdd67aeec92ecccdb` `chunk_id=srcchunk_3c27dc4b62decd1c28fcf4a01fe620a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775676631.557519` `source_timestamp=2026-04-08T19:47:50Z`
- Example: User John with 1.5x multiplier completes 10 English audio tasks ($0.05 each) and 5 video tasks ($0.08 each); with 5 audio and 3 video verified, pending balance is $1.35 and settled balance is $0.735. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_43d72f4fa6e027dcdd67aeec92ecccdb` `chunk_id=srcchunk_3c27dc4b62decd1c28fcf4a01fe620a2` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775676631.557519` `source_timestamp=2026-04-08T19:47:50Z`
- A person-specific coefficient beyond the Poseidon multiplier may be needed, focusing on reliability or a trust metric; the m multiplier could potentially incorporate this. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_184710a7640863c66098697c3974fa70` `chunk_id=srcchunk_a0ccb4b6edd6383429932df380a5455b` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775683932.081089` `source_timestamp=2026-04-08T21:32:12Z`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_60aa0731ad6db323830c0faea10b1dfe` `chunk_id=srcchunk_0b1889dd611b00aafd6b9c7a4b54abfd` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775683955.432939` `source_timestamp=2026-04-08T21:32:35Z`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_0851fb719899954133cfaad7e103521c` `chunk_id=srcchunk_368de7195a6ac19ed9f2ea77fe5291b0` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775683966.515299` `source_timestamp=2026-04-08T21:32:46Z`
  - citation: `source_document_id=srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f` `source_revision_id=srcrev_06d81fb719957f3f2378a8d9b2487856` `chunk_id=srcchunk_a260b929033fc76f871f04cbc7852771` `native_locator=slack:C0AL7EKNHDF:1775673335.424119:1775683988.558389` `source_timestamp=2026-04-08T21:33:08Z`

## Open Questions

- How will p coefficients be determined and documented for each campaign?
- Should there be a separate person-specific reliability coefficient, or can m be used for both Poseidon legacy and ongoing trust?
- What is the exact definition and calculation of the trust/reliability metric?

## Related Pages

- `poseidon-to-numo-migration-strategy`

## Sources

- `source_document_id`: `srcdoc_e2c39193cfd2b7e4a6fa04a2aec9039f`
- `source_revision_id`: `srcrev_0ff7de8b3737d513b372fae52f3397c1`
