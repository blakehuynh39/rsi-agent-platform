---
title: "Story Orchestration Service Nil Pointer Dereference Runbook"
type: "runbook"
slug: "runbooks/nil-pointer-dereference-story-orchestration-service-runbook"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "incident"
  - "nil-pointer-dereference"
  - "production"
  - "story-orchestration-service"
owners:
  - "RSI"
source_revision_ids:
  - "srcrev_61c6bf7d60075b0ddbc61f685c20567c"
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
conflict_state: "none"
---

# Story Orchestration Service Nil Pointer Dereference Runbook

## Summary

Runbook for diagnosing and fixing nil pointer dereference errors in story-orchestration-service production.

## Claims

- The story-orchestration-service experienced a nil pointer dereference error: "runtime error: invalid memory address or nil pointer dereference". `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_61c6bf7d60075b0ddbc61f685c20567c` `chunk_id=srcchunk_bb77c6e2bdd54fb93f73148352c63993` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780554037.550049` `source_timestamp=2026-06-04T06:20:37Z`
- The error has occurred 2,046 times since February 12 in production. Root cause: GetJobByWorkflowID returning nil, nil for non-existent workflow IDs, and three Temporal workflow handler methods dereferencing the nil pointer without a nil check. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- A fix was committed (PR #718) by RSI at 07:11 UTC on the day of the conversation, deployed to staging but not yet in production. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- The RSI executor (rsi-platform-bot) cannot sign commits due to missing GPG/SSH keys and GitHub App token limitations. Commit fa1f78e was created unsigned, requiring a human operator to amend and sign it, or infrastructure team to provision signing keys. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Open Questions

- How will the RSI bot commit signing capability be resolved (human operator vs infrastructure provisioning)?
- When will the fix be deployed to production?

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_d226b5407b816c642314daa241f600fd`
