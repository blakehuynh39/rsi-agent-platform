---
title: "Poseidon Season 1 Multiplier System"
type: "concept"
slug: "concepts/poseidon-season-1-multiplier-system"
freshness: "2026-05-07T22:43:00Z"
tags:
  - "multiplier"
  - "poseidon"
  - "season-1"
owners: []
source_revision_ids:
  - "srcrev_7bc1173911e7a505517ed1b8e16d6a36"
  - "srcrev_884bc20d3537781a45f29ba768d1c1db"
  - "srcrev_fa76f93ffbc00a7c9feb5b10124fd4b7"
conflict_state: "none"
---

# Poseidon Season 1 Multiplier System

## Summary

The Poseidon Season 1 multiplier system assigns earnings multipliers to contributors based on submission count and quality (tiers) and language weights, with a 1.05–2.00 clamp and a 12-week decay schedule.

## Claims

- Tiers are assigned based on submission count and average score: T4 (≥50 sub + avg ≥0.85 OR ≥5 qualifying languages), base 2.00×; T3 (≥20 sub + avg ≥0.80), base 1.60×; T2 (≥5 sub + avg ≥0.70), base 1.30×; T1 (≥1 sub), base 1.10×. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_884bc20d3537781a45f29ba768d1c1db` `chunk_id=srcchunk_2a02982e0e209e7300bbb1645524da91` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778172892.800209` `source_timestamp=2026-05-07T16:54:52Z`
- Language weight adjustments: +0.30 for priority set (hi, ur, mr, bn, ta, te, gu, pa, kn, ml, id); +0.15 for vi; 0.00 for other non-English; -0.15 for en. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_884bc20d3537781a45f29ba768d1c1db` `chunk_id=srcchunk_2a02982e0e209e7300bbb1645524da91` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778172892.800209` `source_timestamp=2026-05-07T16:54:52Z`
- The final initial multiplier is clamped between 1.05 and 2.00 (clamp(tier_value + lang_value, 1.05, 2.00), rounded to 4 decimals). `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_884bc20d3537781a45f29ba768d1c1db` `chunk_id=srcchunk_2a02982e0e209e7300bbb1645524da91` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778172892.800209` `source_timestamp=2026-05-07T16:54:52Z`
- The multiplier decays: full value for 4 weeks, linear decay to 1.0× over the next 8 weeks (back to baseline at week 12). `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_884bc20d3537781a45f29ba768d1c1db` `chunk_id=srcchunk_2a02982e0e209e7300bbb1645524da91` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778172892.800209` `source_timestamp=2026-05-07T16:54:52Z`
- The overall effective multiplier is further capped at 2.00 (max_effective_multiplier in reward_config) with a floor of 1.0. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_7bc1173911e7a505517ed1b8e16d6a36` `chunk_id=srcchunk_788c646c335346f362e79267cbbc9ff8` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778193604.052329` `source_timestamp=2026-05-07T22:40:04Z`
- The 1.05 floor ensures T1 English contributors (raw 0.95) still receive a 1.05× multiplier. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_fa76f93ffbc00a7c9feb5b10124fd4b7` `chunk_id=srcchunk_77ac3168559a601e002a2d4636956c2f` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778193780.763059` `source_timestamp=2026-05-07T22:43:00Z`
- Published cohort distribution: T1 29,296 users, T2 20,797 users, T3 5,343 users, T4 2,229 users. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_884bc20d3537781a45f29ba768d1c1db` `chunk_id=srcchunk_2a02982e0e209e7300bbb1645524da91` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778172892.800209` `source_timestamp=2026-05-07T16:54:52Z`

## Open Questions

- Could the community push back on the decay mechanism if not communicated upfront?
- Is the decay curve information appropriate for public FAQs, and might the timeline change?
- Should the tiered multiplier system be publicly shared to encourage participation?

## Related Pages

- `language-target-autoscaling`
- `multiplier-tasks-task-bars`

## Sources

- `source_document_id`: `srcdoc_fb33ad0ab045846f1631643dce7e41e5`
- `source_revision_id`: `srcrev_fca8765b3b8ea9727d6055c45fb6fd6c`
