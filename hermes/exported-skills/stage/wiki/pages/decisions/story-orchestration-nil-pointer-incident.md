---
title: "story-orchestration-service nil pointer dereference incident"
type: "decision"
slug: "decisions/story-orchestration-nil-pointer-incident"
freshness: "2026-06-04T17:17:00Z"
tags:
  - "incident"
  - "nil-pointer"
  - "story-orchestration"
owners: []
source_revision_ids:
  - "srcrev_61c6bf7d60075b0ddbc61f685c20567c"
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
conflict_state: "none"
---

# story-orchestration-service nil pointer dereference incident

## Summary

The story-orchestration-service experienced a nil pointer dereference in production, with 2,046 events since Feb 12. Root cause was GetJobByWorkflowID returning nil for non-existent workflow IDs, leading to nil dereference. A fix (PR #718) was committed and deployed to staging but not yet to production.

## Claims

- The story-orchestration-service experienced a runtime error: invalid memory address or nil pointer dereference. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_61c6bf7d60075b0ddbc61f685c20567c` `chunk_id=srcchunk_bb77c6e2bdd54fb93f73148352c63993` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780554037.550049` `source_timestamp=2026-06-04T06:20:37Z`
- The nil pointer dereference occurred in production, with 2,046 events since February 12. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Root cause is GetJobByWorkflowID returning (nil, nil) for non-existent workflow IDs, and three Temporal workflow handler methods dereferencing the nil pointer without a nil check. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- A fix (PR #718) was committed by RSI today at 07:11 UTC, deployed to staging but not yet in production. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`

## Open Questions

- When will the fix be deployed to production?

## Related Pages

- `rsi-executor-commit-signing`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_8cf79c9aa0d3d09f77eaa5d4312946f3`
