---
title: "Adaptive Trust Sampling"
type: "system"
slug: "systems/adaptive-trust-sampling"
freshness: "2026-04-08T15:53:00Z"
tags:
  - "bayesian"
  - "contributor-scoring"
  - "cost-efficiency"
  - "numo"
  - "quality"
owners: []
source_revision_ids:
  - "srcrev_367dd868f8187614f467db57dbb74659"
conflict_state: "none"
---

# Adaptive Trust Sampling

## Summary

Adaptive Trust Sampling (ATS) is a contributor-quality evaluation system that uses Bayesian updating to dynamically adjust review levels per user, minimizing review cost while identifying reliable contributors. It classifies users into trusted, uncertain, and low-confidence lanes with corresponding audit rates.

## Claims

- Adaptive Trust Sampling is a contributor-quality evaluation system designed to identify which users consistently submit high-quality data while minimizing review cost by dynamically adjusting review levels based on evidence. `claim:claim_ats_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-1) `source_document_id=srcdoc_d694d4c3868a7ea422d75a2863487a8b` `source_revision_id=srcrev_367dd868f8187614f467db57dbb74659` `chunk_id=srcchunk_f9600e8b4f7a048744b0cd4c3e097b85` `native_locator=https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-1` `source_timestamp=2026-04-08T15:53:00Z`
- The system classifies contributors into three routing lanes: Trusted (posterior mean ≥ 0.90, 90% lower bound ≥ 0.80), Uncertain (posterior mean 0.70–0.90 or insufficient confidence), and Low confidence (posterior mean < 0.70 or repeated integrity issues). `claim:claim_ats_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-4) `source_document_id=srcdoc_d694d4c3868a7ea422d75a2863487a8b` `source_revision_id=srcrev_367dd868f8187614f467db57dbb74659` `chunk_id=srcchunk_5c349c457d865b7442e1b595ce6b37d5` `native_locator=https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-4` `source_timestamp=2026-04-08T15:53:00Z`
- Ongoing audit rates by lane are: Trusted 5%, Uncertain 25–50%, Low confidence 50–100%. `claim:claim_ats_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-4) `source_document_id=srcdoc_d694d4c3868a7ea422d75a2863487a8b` `source_revision_id=srcrev_367dd868f8187614f467db57dbb74659` `chunk_id=srcchunk_5c349c457d865b7442e1b595ce6b37d5` `native_locator=https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-4` `source_timestamp=2026-04-08T15:53:00Z`
- The system uses a Beta-Bernoulli Bayesian model with a Beta(2,2) prior to estimate contributor quality, producing a posterior mean and credible interval that avoids overconfidence from small samples. `claim:claim_ats_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-2) `source_document_id=srcdoc_d694d4c3868a7ea422d75a2863487a8b` `source_revision_id=srcrev_367dd868f8187614f467db57dbb74659` `chunk_id=srcchunk_d497779cb50e2621eaa8267349648f67` `native_locator=https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-2` `source_timestamp=2026-04-08T15:53:00Z`
- The recommended reliability score for MVP is the posterior mean quality, with a production version adding confidence, integrity, and difficulty normalization factors. `claim:claim_ats_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-2) `source_document_id=srcdoc_d694d4c3868a7ea422d75a2863487a8b` `source_revision_id=srcrev_367dd868f8187614f467db57dbb74659` `chunk_id=srcchunk_d497779cb50e2621eaa8267349648f67` `native_locator=https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-2` `source_timestamp=2026-04-08T15:53:00Z`
- The system supports recency weighting, consistency penalties, gold tasks, difficulty normalization, and fraud/anomaly integration as additional improvements beyond MVP. `claim:claim_ats_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-3) `source_document_id=srcdoc_d694d4c3868a7ea422d75a2863487a8b` `source_revision_id=srcrev_367dd868f8187614f467db57dbb74659` `chunk_id=srcchunk_d814d9e0d3d86a9a091df88e576e82a4` `native_locator=https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-3` `source_timestamp=2026-04-08T15:53:00Z`
- Anti-gaming controls include keeping audit logic private, random spot checks for trusted users, delayed audit windows, gold tasks, and monitoring distribution shifts. `claim:claim_ats_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-3) `source_document_id=srcdoc_d694d4c3868a7ea422d75a2863487a8b` `source_revision_id=srcrev_367dd868f8187614f467db57dbb74659` `chunk_id=srcchunk_d814d9e0d3d86a9a091df88e576e82a4` `native_locator=https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-3` `source_timestamp=2026-04-08T15:53:00Z`
- The system enables contributor tiers (New, Verified, Trusted, Specialist), campaign gating, and differentiated payout confidence based on reliability scores. `claim:claim_ats_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-3) `source_document_id=srcdoc_d694d4c3868a7ea422d75a2863487a8b` `source_revision_id=srcrev_367dd868f8187614f467db57dbb74659` `chunk_id=srcchunk_d814d9e0d3d86a9a091df88e576e82a4` `native_locator=https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265#chunk-3` `source_timestamp=2026-04-08T15:53:00Z`

## Related Pages

- `projects/numo-app-strategy`

## Sources

- `source_document_id`: `srcdoc_d694d4c3868a7ea422d75a2863487a8b`
- `source_revision_id`: `srcrev_367dd868f8187614f467db57dbb74659`
- `source_url`: [Notion source](https://www.notion.so/Adaptive-Trust-Sampling-33c051299a5480789341f45c15460265)
