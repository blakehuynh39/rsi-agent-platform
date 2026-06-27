---
title: "Runbook: Nil Pointer Dereference in story-orchestration-service"
type: "runbook"
slug: "runbooks/nil-pointer-dereference-story-orchestration-service"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "golang"
  - "incident"
  - "nil-pointer"
  - "story-orchestration-service"
  - "temporal"
owners:
  - "RSI Team"
source_revision_ids:
  - "srcrev_61c6bf7d60075b0ddbc61f685c20567c"
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
  - "srcrev_d226b5407b816c642314daa241f600fd"
conflict_state: "none"
---

# Runbook: Nil Pointer Dereference in story-orchestration-service

## Summary

Diagnosis and fix for a production nil pointer dereference in story-orchestration-service. Root cause: GetJobByWorkflowID returning nil for non-existent workflow IDs, and Temporal workflow handlers dereferencing the nil pointer without a check. The fix (PR #718) is deployed to staging but not yet production. Also covers commit signing issue for the RSI bot that made the fix.

## Claims

- story-orchestration-service experienced a runtime error: invalid memory address or nil pointer dereference. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_61c6bf7d60075b0ddbc61f685c20567c` `chunk_id=srcchunk_bb77c6e2bdd54fb93f73148352c63993` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780554037.550049` `source_timestamp=2026-06-04T06:20:37Z`
- The error has occurred 2,046 times since Feb 12 in production. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Root cause is GetJobByWorkflowID returning nil, nil for non-existent workflow IDs, and three Temporal workflow handler methods dereferencing the nil pointer without a nil check. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- The fix is in PR #718, committed by RSI at 07:11 UTC. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- The fix is deployed to staging but not yet in production. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- The commit (fa1f78e) lacks a verified signature and violates GitHub policy. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_d226b5407b816c642314daa241f600fd` `chunk_id=srcchunk_dffac0c9df4d3ae01ea115f64eb5809d` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593723.167589` `source_timestamp=2026-06-04T17:22:03Z`
- The RSI executor cannot sign commits because it lacks GPG keys, SSH signing keys, and GitHub App tokens do not support signing. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Open Questions

- commit-signing-rsi-executor-bot

## Related Pages

- `commit-signing-rsi-executor-bot`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_7cd9b4be7aa48aaa3ca0433558428118`
