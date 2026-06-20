---
title: "Claude Usage Limits and Upgrade Process"
type: "runbook"
slug: "runbooks/claude-usage-limits-and-upgrade-process"
freshness: "2026-04-21T08:44:33Z"
tags:
  - "billing"
  - "claude"
  - "upgrade"
  - "usage-limits"
owners:
  - "Vinod"
  - "Woojin"
source_revision_ids:
  - "srcrev_04bd521a1bd2e972efb992c421047e92"
  - "srcrev_15969a2b595eea0684fc66a6e192d20e"
  - "srcrev_5b9a052dc97267e7b1a93f3b804803dd"
  - "srcrev_7ccb979a13e45a6e42274f4d345fc4e3"
  - "srcrev_909ea3863068ad2ffab8ccab498ae8dc"
  - "srcrev_95c0ff72913c31ae5409ed0373a91985"
  - "srcrev_9b5d93c713448f2c71e0b013f754e8ca"
  - "srcrev_b3c06be55d0106f83132b579edd1dfb0"
  - "srcrev_d83a926720b8b4095962f87ea69d44e0"
  - "srcrev_ed27342063cf9629c2ef018745ea30a6"
  - "srcrev_f6f7422cf7cb603d80f33786d18e7b72"
conflict_state: "none"
---

# Claude Usage Limits and Upgrade Process

## Summary

Process for handling Claude usage limits (session caps, token quotas) on the Team plan. Admins Woojin and Vinod can adjust credits and plan tiers. Vinod is on vacation until April 15, 2026.

## Claims

- Yao hit the 5-hour session cap on the Claude Code Team plan. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_d83a926720b8b4095962f87ea69d44e0` `chunk_id=srcchunk_6dec9492fdb2765ba1e34234623481fc` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759478.608379` `source_timestamp=2026-04-21T08:17:58Z`
- Woojin confirmed Yao's seat is 'premium'. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_909ea3863068ad2ffab8ccab498ae8dc` `chunk_id=srcchunk_ab92cdc77428850af01aba93be136238` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759607.068049` `source_timestamp=2026-04-21T08:20:07Z`
- Woojin added more credit to the org to increase the usage limit. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_04bd521a1bd2e972efb992c421047e92` `chunk_id=srcchunk_c4e8e5c5b4db78cecfc9ba5876f5242d` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759640.394219` `source_timestamp=2026-04-21T08:20:40Z`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_b3c06be55d0106f83132b579edd1dfb0` `chunk_id=srcchunk_6068396a508bc09bcb0280e22488aa6d` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759706.354769` `source_timestamp=2026-04-21T08:22:32Z`
- Vinod is on vacation through April 15, 2026. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_ed27342063cf9629c2ef018745ea30a6` `chunk_id=srcchunk_32ba820caf397b295e01a74c74f0bb51` `native_locator=slack:C0547N89JUB:1776759478.608379:1776761073.028099` `source_timestamp=2026-04-21T08:44:33Z`
- Another user hit 88% of weekly token quota and needed an upgrade; Woojin will discuss with Vinod. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_9b5d93c713448f2c71e0b013f754e8ca` `chunk_id=srcchunk_8199717321d2b6d38765b943c16041b9` `native_locator=slack:C0547N89JUB:1776759478.608379:1776760665.662889` `source_timestamp=2026-04-21T08:37:45Z`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_5b9a052dc97267e7b1a93f3b804803dd` `chunk_id=srcchunk_f5d0f774da18eade2dcda049cca047a0` `native_locator=slack:C0547N89JUB:1776759478.608379:1776761055.006979` `source_timestamp=2026-04-21T08:44:15Z`
- The IT support bot cannot manage Claude billing or plan details; it can only create tasks for the IT admin. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_95c0ff72913c31ae5409ed0373a91985` `chunk_id=srcchunk_28c6fb52ebca1116da3801102307fbf8` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759487.293099` `source_timestamp=2026-04-21T08:18:07Z`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_15969a2b595eea0684fc66a6e192d20e` `chunk_id=srcchunk_90f77fb97caa0ad90393e6a2fb17b690` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759548.665999` `source_timestamp=2026-04-21T08:19:08Z`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_f6f7422cf7cb603d80f33786d18e7b72` `chunk_id=srcchunk_44244b0b826b57d4f9526dd6434db596` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759613.779839` `source_timestamp=2026-04-21T08:20:13Z`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_7ccb979a13e45a6e42274f4d345fc4e3` `chunk_id=srcchunk_063e6f8489396b75429c14cad19605e3` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759634.433599` `source_timestamp=2026-04-21T08:20:34Z`
- To check current seat level, users should check the Anthropic console or ask Woojin or Vinod. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_15969a2b595eea0684fc66a6e192d20e` `chunk_id=srcchunk_90f77fb97caa0ad90393e6a2fb17b690` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759548.665999` `source_timestamp=2026-04-21T08:19:08Z`
- The IT support bot suggests priority levels P0 (Critical), P1 (High), P2 (Medium/low) for task creation. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_95c0ff72913c31ae5409ed0373a91985` `chunk_id=srcchunk_28c6fb52ebca1116da3801102307fbf8` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759487.293099` `source_timestamp=2026-04-21T08:18:07Z`
- There is a maximum seat/spend that can be increased by adding custom spend. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5a728370e61d3395ebda39217944bd3d` `source_revision_id=srcrev_04bd521a1bd2e972efb992c421047e92` `chunk_id=srcchunk_c4e8e5c5b4db78cecfc9ba5876f5242d` `native_locator=slack:C0547N89JUB:1776759478.608379:1776759640.394219` `source_timestamp=2026-04-21T08:20:40Z`

## Open Questions

- How to enable API billing fallback in Claude workspace settings?
- What are the specific Claude Team plan tiers and their limits?
- What is the process for upgrading user seats to handle token quota limits?

## Related Pages

- `claude-team-plan`
- `it-support-bot`
- `person-vinod`
- `person-woojin`

## Sources

- `source_document_id`: `srcdoc_5a728370e61d3395ebda39217944bd3d`
- `source_revision_id`: `srcrev_ed27342063cf9629c2ef018745ea30a6`
