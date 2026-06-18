---
title: "Nil pointer dereference in story-orchestration-service"
type: "runbook"
slug: "runbooks/incident-nil-pointer-deref-story-orchestration"
freshness: "2026-06-04T17:22:03Z"
tags:
  - "golang"
  - "incident"
  - "nil-pointer-dereference"
  - "story-orchestration-service"
owners:
  - "Platform Engineering"
source_revision_ids:
  - "srcrev_61c6bf7d60075b0ddbc61f685c20567c"
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
  - "srcrev_d226b5407b816c642314daa241f600fd"
conflict_state: "none"
---

# Nil pointer dereference in story-orchestration-service

## Summary

On Feb 12, story-orchestration-service started experiencing nil pointer dereferences (2,046 events). Root cause: GetJobByWorkflowID returning nil, nil for non-existent workflow IDs, and Temporal workflow handlers lacking nil checks. Fix PR #718 committed to staging, but not yet production. Commit lacks verified signature.

## Claims

- story-orchestration-service encountered a runtime error: 'invalid memory address or nil pointer dereference'. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_61c6bf7d60075b0ddbc61f685c20567c` `chunk_id=srcchunk_bb77c6e2bdd54fb93f73148352c63993` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780554037.550049` `source_timestamp=2026-06-04T06:20:37Z`
- 2,046 nil pointer dereference events have occurred since Feb 12 in production. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Root cause: GetJobByWorkflowID returns nil, nil for non-existent workflow IDs, and the three Temporal workflow handler methods dereference the nil pointer without a nil check. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- RSI committed a fix in PR #718 at 07:11 UTC, which was deployed to staging but not yet to production. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- The commit for the fix (fa1f78e) lacks a verified signature and violates GitHub policy. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_d226b5407b816c642314daa241f600fd` `chunk_id=srcchunk_dffac0c9df4d3ae01ea115f64eb5809d` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593723.167589` `source_timestamp=2026-06-04T17:22:03Z`

## Related Pages

- `rsi-executor-commit-signing`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944`
