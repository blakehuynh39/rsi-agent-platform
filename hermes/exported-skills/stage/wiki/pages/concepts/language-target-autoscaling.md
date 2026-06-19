---
title: "Language Target Auto-Scaling Pay Rates"
type: "concept"
slug: "concepts/language-target-autoscaling"
freshness: "2026-05-07T16:54:52Z"
tags:
  - "autoscaling"
  - "language"
  - "pricing"
owners: []
source_revision_ids:
  - "srcrev_884bc20d3537781a45f29ba768d1c1db"
conflict_state: "none"
---

# Language Target Auto-Scaling Pay Rates

## Summary

Numo dynamically adjusts reward amounts per task for each language based on how close the language is to its target hours goal, using a formula to direct contributor supply where it is most needed.

## Claims

- Each language has a target_hours goal; as current_hours rises, reward_amount_usd scales down via the formula completion = min(current_hours/target_hours,1); gap=1−completion; rate=min_rate+(max_rate−min_rate)×gap. `claim:claim_3_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_884bc20d3537781a45f29ba768d1c1db` `chunk_id=srcchunk_2a02982e0e209e7300bbb1645524da91` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778172892.800209` `source_timestamp=2026-05-07T16:54:52Z`
- Priority languages (hi, ta, te, bn, mr, gu, kn, ml, pa, or, id) use this mechanism with specific min/max rates to direct effort where needed. `claim:claim_3_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_884bc20d3537781a45f29ba768d1c1db` `chunk_id=srcchunk_2a02982e0e209e7300bbb1645524da91` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778172892.800209` `source_timestamp=2026-05-07T16:54:52Z`
- This auto-scaling causes differences in payouts and number of tasks between languages. `claim:claim_3_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_fb33ad0ab045846f1631643dce7e41e5` `source_revision_id=srcrev_884bc20d3537781a45f29ba768d1c1db` `chunk_id=srcchunk_2a02982e0e209e7300bbb1645524da91` `native_locator=slack:C0AL7EKNHDF:1778171216.577589:1778172892.800209` `source_timestamp=2026-05-07T16:54:52Z`

## Related Pages

- `poseidon-season-1-multiplier-system`

## Sources

- `source_document_id`: `srcdoc_fb33ad0ab045846f1631643dce7e41e5`
- `source_revision_id`: `srcrev_fca8765b3b8ea9727d6055c45fb6fd6c`
