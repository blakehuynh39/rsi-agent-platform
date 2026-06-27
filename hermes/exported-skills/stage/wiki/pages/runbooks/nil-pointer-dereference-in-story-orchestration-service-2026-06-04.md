---
title: "Nil Pointer Dereference in Story Orchestration Service (2026-06-04)"
type: "runbook"
slug: "runbooks/nil-pointer-dereference-in-story-orchestration-service-2026-06-04"
freshness: "2026-06-04T17:22:05Z"
tags:
  - "bug"
  - "fix-in-progress"
  - "incident"
owners: []
source_revision_ids:
  - "srcrev_031befdaa44a6d6dd99bb5a631b97e31"
  - "srcrev_7cd9b4be7aa48aaa3ca0433558428118"
  - "srcrev_8cf79c9aa0d3d09f77eaa5d4312946f3"
  - "srcrev_a458a652d00ef32049a3cd6ee8e18b56"
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
  - "srcrev_e4e2e360b745af4678884c7b46537238"
conflict_state: "none"
---

# Nil Pointer Dereference in Story Orchestration Service (2026-06-04)

## Summary

A nil pointer dereference in story-orchestration-service production was diagnosed on June 4, 2026. Root cause: GetJobByWorkflowID returns nil,nil for non‑existent workflow IDs; handler methods missing nil checks. Fix PR #718 committed to staging, not yet in production due to unsigned commit policy.

## Claims

- Error observed with 2,046 events since Feb 12. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Root cause: `GetJobByWorkflowID` returns `nil, nil` for non-existent workflow IDs, and handler methods dereference the nil pointer without a nil check. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Fix PR #718 committed at 07:11 UTC by RSI and deployed to staging. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Fix not yet deployed to production. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Diagnosis was requested from user U0ASDQKU3UL. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_e4e2e360b745af4678884c7b46537238` `chunk_id=srcchunk_906ee42d6119b8a56e21b15b469b6460` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780554692.843629` `source_timestamp=2026-06-04T06:31:32Z`
- RSI session traces available: trace-b4072d25bb744b0ca221339a1e732062 and trace-af2943236bc443bcaf63a9b642236f82 `claim:claim_2_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_031befdaa44a6d6dd99bb5a631b97e31` `chunk_id=srcchunk_f728f0210682df70bd71e8bf06de7435` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593054.313559` `source_timestamp=2026-06-04T17:10:54Z`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_7cd9b4be7aa48aaa3ca0433558428118` `chunk_id=srcchunk_24059ca80678eb27b6908aa42475a8dd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593725.192829` `source_timestamp=2026-06-04T17:22:05Z`
- Timeline and Next Steps attachments provided. `claim:claim_2_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8cf79c9aa0d3d09f77eaa5d4312946f3` `chunk_id=srcchunk_28d9f1df01cc3b1f90301ca6037f2548` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593421.182469` `source_timestamp=2026-06-04T17:17:01Z`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_a458a652d00ef32049a3cd6ee8e18b56` `chunk_id=srcchunk_442b175ab353e1cb89c3ded293a54a48` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593421.972899` `source_timestamp=2026-06-04T17:17:01Z`

## Open Questions

- Fix PR #718 is not in production; deployment pending resolution of commit signing policy.

## Related Pages

- `commit-signing-requirements`
- `rsi-executor-commit-signing-capabilities`
- `story-orchestration-service`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_c324be3df9811270afe5ab85f79a8722`
