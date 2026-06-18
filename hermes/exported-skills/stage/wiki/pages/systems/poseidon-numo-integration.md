---
title: "Poseidon ↔ Numo Integration"
type: "system"
slug: "systems/poseidon-numo-integration"
freshness: "2026-06-11T20:38:57Z"
tags:
  - "deepfake"
  - "integration"
  - "payments"
  - "validation"
owners:
  - "New App Team"
  - "Poseidon Team"
source_revision_ids:
  - "srcrev_0b09b90b4e8716ba20fa7b92c005983f"
  - "srcrev_247cf9cf83d1f37502d519c7f6c608c8"
  - "srcrev_48a82b195eaa53cfa654485e122e2092"
  - "srcrev_4c3f0dc86034f897d12d169f0dab0b77"
  - "srcrev_85cde253f61e7e890c5978c07986b392"
  - "srcrev_999623960445b773d7cdec356a08b904"
  - "srcrev_a7d725d2c517e581d70ee297b9af6084"
  - "srcrev_b695b420e6a6ccbf1c490f3bc2266ae2"
  - "srcrev_cbe1d808762afd782e4a5036f446a335"
conflict_state: "none"
---

# Poseidon ↔ Numo Integration

## Summary

Integration between Poseidon (Story Protocol) and Numo for data validation, deepfake scoring, and reward propagation. Covers the Season 1 Multiplier (live) and the Data Validation / NDV pipeline (PR #509).

## Claims

- Deepfake score and verdict will be added to the Numo validation result payload per submission. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_247cf9cf83d1f37502d519c7f6c608c8` `chunk_id=srcchunk_52eb39c016f4de91c73b6818247557d6` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781122593.398499` `source_timestamp=2026-06-10T20:16:33Z`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_4c3f0dc86034f897d12d169f0dab0b77` `chunk_id=srcchunk_6d4161445a32d25d33a24c25c0546b64` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781205528.077089` `source_timestamp=2026-06-11T19:18:48Z`
- There are two distinct Poseidon ↔ Numo integrations: Season 1 Multiplier (live) and Data Validation / NDV pipeline (PR #509, in progress). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_cbe1d808762afd782e4a5036f446a335` `chunk_id=srcchunk_7eaf889fa9ec108a6991cee359f6d6e1` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781209456.774519` `source_timestamp=2026-06-11T20:24:16Z`
- Both integrations automatically propagate to the user's rewards/wallet page. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_cbe1d808762afd782e4a5036f446a335` `chunk_id=srcchunk_7eaf889fa9ec108a6991cee359f6d6e1` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781209456.774519` `source_timestamp=2026-06-11T20:24:16Z`
- Rejected submissions keep the 15% advance but forfeit the 85% remainder. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_a7d725d2c517e581d70ee297b9af6084` `chunk_id=srcchunk_de0bff211b2afc493840693c52367d4e` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781209843.516479` `source_timestamp=2026-06-11T20:30:43Z`
- The backend surfaces lifetime_rejected_count and per-window rejected_count; the wallet page already displays these numbers. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_a7d725d2c517e581d70ee297b9af6084` `chunk_id=srcchunk_de0bff211b2afc493840693c52367d4e` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781209843.516479` `source_timestamp=2026-06-11T20:30:43Z`
- Showing a percentage of accepted/failed out of total submissions is a frontend-only change since all three counts (submitted/verified/rejected) are already in the API response. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_a7d725d2c517e581d70ee297b9af6084` `chunk_id=srcchunk_de0bff211b2afc493840693c52367d4e` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781209843.516479` `source_timestamp=2026-06-11T20:30:43Z`
- When campaigns.auto_apply_validation = FALSE, NDV results are advisory-only; the submission stays in pending_review and admins own the final decision. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_b695b420e6a6ccbf1c490f3bc2266ae2` `chunk_id=srcchunk_bb5b3d3ca7e1042a8ee3990e84fbaace` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781210337.250759` `source_timestamp=2026-06-11T20:38:57Z`
- The admin buffer for NDV is implemented in PR #509, including three admin endpoints (filter by validation_decision, view validation-history per submission, toggle auto_apply per campaign) and admin UI panels (validation column, detail panel, campaign settings toggle). `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_b695b420e6a6ccbf1c490f3bc2266ae2` `chunk_id=srcchunk_bb5b3d3ca7e1042a8ee3990e84fbaace` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781210337.250759` `source_timestamp=2026-06-11T20:38:57Z`
- An architecture diagram for Poseidon ↔ Numo was shared as an SVG file. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_85cde253f61e7e890c5978c07986b392` `chunk_id=srcchunk_4ee5abf5bd8fe60ccb66341506cdabdd` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781209591.674159` `source_timestamp=2026-06-11T20:26:31Z`
- Collaborators for the deepfake payload change included @U04L0DD6B6F, @U083MMT1771, and @U0A2D9U625V. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_247cf9cf83d1f37502d519c7f6c608c8` `chunk_id=srcchunk_52eb39c016f4de91c73b6818247557d6` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781122593.398499` `source_timestamp=2026-06-10T20:16:33Z`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_48a82b195eaa53cfa654485e122e2092` `chunk_id=srcchunk_2eadc1092b7aab5e1b8e91e91a64f5a5` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781133977.704319` `source_timestamp=2026-06-10T23:26:17Z`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_999623960445b773d7cdec356a08b904` `chunk_id=srcchunk_67ca958102feb0c3c8f46aa19664d175` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781144824.888289` `source_timestamp=2026-06-11T02:27:04Z`
- A team member moved to the Lion team for IP‑data work, requesting that the new app team take over the Numo integration. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7` `source_revision_id=srcrev_0b09b90b4e8716ba20fa7b92c005983f` `chunk_id=srcchunk_b730fff0533357dcf3a7231791fb1a65` `native_locator=slack:C0AL7EKNHDF:1781122593.398499:1781200532.469879` `source_timestamp=2026-06-11T19:21:11Z`

## Sources

- `source_document_id`: `srcdoc_cb2da0d63cf01b7cd8599f8f353e72c7`
- `source_revision_id`: `srcrev_b695b420e6a6ccbf1c490f3bc2266ae2`
