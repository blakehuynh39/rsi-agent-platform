---
title: "Decision: Remove storyapis.com TLV backend domains from Cloudflare"
type: "decision"
slug: "decisions/decision-remove-storyapis-tlv-backend-domains"
freshness: "2026-05-26T01:20:46Z"
tags:
  - "cloudflare"
  - "domain"
  - "removal"
  - "storyapis"
  - "tlv"
owners: []
source_revision_ids:
  - "srcrev_01d7ed301d3890a5bd557f1894cac94c"
  - "srcrev_58c0f3aa773e27a4f4b4eecc5146c298"
  - "srcrev_61740316b097f6f41947cda85c8c4cdd"
  - "srcrev_8c008c94e6c4bb8e44139fa8516c9863"
  - "srcrev_ec8e7e115b265766cf908df15a9132a0"
conflict_state: "none"
---

# Decision: Remove storyapis.com TLV backend domains from Cloudflare

## Summary

The domains ff-backend.storyapis.com and use1-stage-tlv-backend.storyapis.com, previously owned by Don, were confirmed for removal from Cloudflare configuration via PR #206.

## Claims

- A request was made to remove a domain from Cloudflare configuration via GitHub PR #206. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6e93d673902011a3a98080e14ce45cb7` `source_revision_id=srcrev_8c008c94e6c4bb8e44139fa8516c9863` `chunk_id=srcchunk_5a414f5c71325802a67184da319917c2` `native_locator=slack:C0547N89JUB:1779410458.378569:1779410458.378569` `source_timestamp=2026-05-22T00:40:58Z`
- The TLV backend domains can be confirmed by a specific user. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6e93d673902011a3a98080e14ce45cb7` `source_revision_id=srcrev_61740316b097f6f41947cda85c8c4cdd` `chunk_id=srcchunk_528dffc1b40363acbb9cda76604343e8` `native_locator=slack:C0547N89JUB:1779410458.378569:1779411150.938299` `source_timestamp=2026-05-22T00:52:30Z`
- The domains to be removed are ff-backend.storyapis.com and use1-stage-tlv-backend.storyapis.com, and were previously owned by Don. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6e93d673902011a3a98080e14ce45cb7` `source_revision_id=srcrev_ec8e7e115b265766cf908df15a9132a0` `chunk_id=srcchunk_7d0adba378d058c943331d58833a615f` `native_locator=slack:C0547N89JUB:1779410458.378569:1779419870.127519` `source_timestamp=2026-05-22T03:17:50Z`
- A confirmation was requested again. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6e93d673902011a3a98080e14ce45cb7` `source_revision_id=srcrev_58c0f3aa773e27a4f4b4eecc5146c298` `chunk_id=srcchunk_e84cf6a5d856848cb1a896d185f5681a` `native_locator=slack:C0547N89JUB:1779410458.378569:1779750734.052139` `source_timestamp=2026-05-25T23:12:14Z`
- The domains can be removed; the removal was confirmed with 'yeah it can be removed'. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6e93d673902011a3a98080e14ce45cb7` `source_revision_id=srcrev_01d7ed301d3890a5bd557f1894cac94c` `chunk_id=srcchunk_9eca5df55bcc4d8d4d0977f32cc8f0a2` `native_locator=slack:C0547N89JUB:1779410458.378569:1779758446.424569` `source_timestamp=2026-05-26T01:20:46Z`

## Sources

- `source_document_id`: `srcdoc_6e93d673902011a3a98080e14ce45cb7`
- `source_revision_id`: `srcrev_01d7ed301d3890a5bd557f1894cac94c`
