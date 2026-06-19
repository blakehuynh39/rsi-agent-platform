---
title: "Rate Limiting Configuration for numolabs.ai"
type: "decision"
slug: "decisions/rate-limiting-config-numolabs"
freshness: "2026-04-28T02:31:08Z"
tags:
  - "cloudflare"
  - "numolabs"
  - "rate-limiting"
owners:
  - "U07TNT9N4JC"
  - "U083MMT1771"
source_revision_ids:
  - "srcrev_11c2af98fc6847799d4e53b59c293505"
  - "srcrev_222f61ded78866666f6f3f42370e652c"
  - "srcrev_50eacba7fab6c754098ccbb9509cf651"
  - "srcrev_6efd94c0e3e3046947c3d0d8ba498159"
  - "srcrev_80e6526adf167818166a179b7b40d6bb"
  - "srcrev_992aa152f9b81fb6511191083451cb75"
  - "srcrev_e10c90d9ece68b7e29146aefe03d49f5"
conflict_state: "none"
---

# Rate Limiting Configuration for numolabs.ai

## Summary

Defines seven rate limiting rules for numolabs.ai Cloudflare zone, covering authentication, reward actions, enumeration-prone reads, user data reads, profile, API global backstop, and app domain backstop. Includes IP allowlist for internal servers and QA testers. Configuration required a plan upgrade to Business.

## Claims

- A GitHub pull request (#192) added rate limiting and WAF rules to the numolabs.ai Cloudflare zone. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1f3f4a64978099adae0f7d29bf6f0d26` `source_revision_id=srcrev_50eacba7fab6c754098ccbb9509cf651` `chunk_id=srcchunk_8e87708ec8be170b8c60c577b4bad2a1` `native_locator=slack:C0AL7EKNHDF:1777327487.626569:1777327487.626569` `source_timestamp=2026-04-27T22:04:47Z`
- The proposed rate limiting rules include: Auth (10 req/min, 2-min lock), Reward actions (15 req/min, 2-min lock), Enumeration-prone reads (20 req/min, 5-min lock), User data reads (60 req/min, 1-min lock), Profile endpoint (120 req/min, 1-min lock), API global backstop (200 req/min, 1-min lock), App domain backstop (600 req/min, 1-min lock). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1f3f4a64978099adae0f7d29bf6f0d26` `source_revision_id=srcrev_11c2af98fc6847799d4e53b59c293505` `chunk_id=srcchunk_925b38b30010d57efc6b66a6c32effbb` `native_locator=slack:C0AL7EKNHDF:1777327487.626569:1777336307.045629` `source_timestamp=2026-04-28T00:31:47Z`
- Internal servers and QA testers are exempt from rate limits via IP allowlists. The app domain backstop exempts QA only, as no internal services call the frontend directly. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1f3f4a64978099adae0f7d29bf6f0d26` `source_revision_id=srcrev_11c2af98fc6847799d4e53b59c293505` `chunk_id=srcchunk_925b38b30010d57efc6b66a6c32effbb` `native_locator=slack:C0AL7EKNHDF:1777327487.626569:1777336307.045629` `source_timestamp=2026-04-28T00:31:47Z`
- The initial configuration was on a free tier, but it was upgraded to a Business plan to better support the rate limiting rules. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1f3f4a64978099adae0f7d29bf6f0d26` `source_revision_id=srcrev_222f61ded78866666f6f3f42370e652c` `chunk_id=srcchunk_e7317dd3714f77640f3832c89faf5a3f` `native_locator=slack:C0AL7EKNHDF:1777327487.626569:1777340356.037039` `source_timestamp=2026-04-28T01:39:16Z`
  - citation: `source_document_id=srcdoc_1f3f4a64978099adae0f7d29bf6f0d26` `source_revision_id=srcrev_80e6526adf167818166a179b7b40d6bb` `chunk_id=srcchunk_f5cd9048d8de43a504316c5e0e9ab8ab` `native_locator=slack:C0AL7EKNHDF:1777327487.626569:1777342476.249919` `source_timestamp=2026-04-28T02:14:36Z`
- Team members approved the pull requests (PR #192 and PR #193) after reviewing the rules. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1f3f4a64978099adae0f7d29bf6f0d26` `source_revision_id=srcrev_992aa152f9b81fb6511191083451cb75` `chunk_id=srcchunk_6edc393726119c6683f3642db0e57d61` `native_locator=slack:C0AL7EKNHDF:1777327487.626569:1777336433.164529` `source_timestamp=2026-04-28T00:33:53Z`
  - citation: `source_document_id=srcdoc_1f3f4a64978099adae0f7d29bf6f0d26` `source_revision_id=srcrev_6efd94c0e3e3046947c3d0d8ba498159` `chunk_id=srcchunk_4b037b1c7c71fabaae921595707c1ee3` `native_locator=slack:C0AL7EKNHDF:1777327487.626569:1777342511.506889` `source_timestamp=2026-04-28T02:15:11Z`
  - citation: `source_document_id=srcdoc_1f3f4a64978099adae0f7d29bf6f0d26` `source_revision_id=srcrev_e10c90d9ece68b7e29146aefe03d49f5` `chunk_id=srcchunk_c64fd4166dc6fda592480124e40cb6b1` `native_locator=slack:C0AL7EKNHDF:1777327487.626569:1777343468.589359` `source_timestamp=2026-04-28T02:31:08Z`

## Open Questions

- Can the Business plan support all seven rate limiting rules given the potential limit of 5 rules and lack of http.request.body.size condition?
- Were any rules dropped or consolidated after merging to fit Business plan limitations?

## Sources

- `source_document_id`: `srcdoc_1f3f4a64978099adae0f7d29bf6f0d26`
- `source_revision_id`: `srcrev_e10c90d9ece68b7e29146aefe03d49f5`
