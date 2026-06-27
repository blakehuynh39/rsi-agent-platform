---
title: "story-orchestration-service nil pointer dereference incident"
type: "runbook"
slug: "runbooks/story-orchestration-nil-pointer-incident"
freshness: "2026-06-04T17:17:01Z"
tags:
  - "incident"
  - "nil-pointer"
  - "story-orchestration-service"
  - "temporal"
owners: []
source_revision_ids:
  - "srcrev_61c6bf7d60075b0ddbc61f685c20567c"
  - "srcrev_8cf79c9aa0d3d09f77eaa5d4312946f3"
  - "srcrev_a458a652d00ef32049a3cd6ee8e18b56"
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
conflict_state: "none"
---

# story-orchestration-service nil pointer dereference incident

## Summary

Production story-orchestration-service encountered 2,046 nil pointer dereference events since Feb 12 due to missing nil checks in Temporal workflow handlers. Fix in PR #718 deployed to staging but not yet production.

## Claims

- story-orchestration-service experienced runtime error: invalid memory address or nil pointer dereference. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_61c6bf7d60075b0ddbc61f685c20567c` `chunk_id=srcchunk_bb77c6e2bdd54fb93f73148352c63993` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780554037.550049` `source_timestamp=2026-06-04T06:20:37Z`
- Diagnosis: nil pointer dereference in story-orchestration-service production — 2,046 events since Feb 12. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Root cause is `GetJobByWorkflowID` returning `nil, nil` for non-existent workflow IDs, and three Temporal workflow handler methods dereferencing the nil pointer without a nil check. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- The fix (PR #718) was committed by RSI today at 07:11 UTC, deployed to staging but not yet in production. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- A CSV timeline file was attached to the Slack thread. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8cf79c9aa0d3d09f77eaa5d4312946f3` `chunk_id=srcchunk_28d9f1df01cc3b1f90301ca6037f2548` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593421.182469` `source_timestamp=2026-06-04T17:17:01Z`
- A CSV next-steps file was attached to the Slack thread. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_a458a652d00ef32049a3cd6ee8e18b56` `chunk_id=srcchunk_442b175ab353e1cb89c3ded293a54a48` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593421.972899` `source_timestamp=2026-06-04T17:17:01Z`

## Open Questions

- When will PR #718 be deployed to production?

## Related Pages

- `rsi-executor-commit-signing-policy`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_a458a652d00ef32049a3cd6ee8e18b56`
