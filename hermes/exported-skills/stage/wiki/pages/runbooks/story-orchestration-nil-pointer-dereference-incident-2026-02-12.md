---
title: "Story Orchestration Service Nil Pointer Dereference Incident (2026-02-12)"
type: "runbook"
slug: "runbooks/story-orchestration-nil-pointer-dereference-incident-2026-02-12"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "incident"
  - "nil-pointer"
  - "story-orchestration-service"
  - "temporal"
owners:
  - "U0ASDQKU3UL"
source_revision_ids:
  - "srcrev_61c6bf7d60075b0ddbc61f685c20567c"
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
  - "srcrev_d226b5407b816c642314daa241f600fd"
conflict_state: "none"
---

# Story Orchestration Service Nil Pointer Dereference Incident (2026-02-12)

## Summary

A nil pointer dereference in story-orchestration-service caused ~2,046 errors in production since Feb 12, 2026. Root cause: GetJobByWorkflowID returns nil for missing workflow IDs, and handlers lacked nil checks. Fix PR #718 committed by RSI bot but lacks verified signature, blocking production deployment.

## Claims

- The story-orchestration-service encountered a nil pointer dereference error in production. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_61c6bf7d60075b0ddbc61f685c20567c` `chunk_id=srcchunk_bb77c6e2bdd54fb93f73148352c63993` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780554037.550049` `source_timestamp=2026-06-04T06:20:37Z`
- The error has occurred 2,046 times since February 12, 2026. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Root cause: GetJobByWorkflowID returns (nil, nil) for non-existent workflow IDs, and three Temporal workflow handler methods dereference the nil pointer without a nil check. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Fix PR #718 was committed by RSI at 07:11 UTC on the day of diagnosis. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- The fix has been deployed to staging but is not yet in production. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- The commit (fa1f78e) lacks a verified GPG/SSH signature and violates GitHub branch protection policy. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_d226b5407b816c642314daa241f600fd` `chunk_id=srcchunk_dffac0c9df4d3ae01ea115f64eb5809d` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593723.167589` `source_timestamp=2026-06-04T17:22:03Z`
- The RSI executor runs as a bot (rsi-platform-bot) via GitHub App token, with no GPG, SSH keys, or signing configuration, and cannot sign commits. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- Resolution options: a human operator amends and force-pushes the signed commit, or infrastructure provides GPG/SSH signing keys to the RSI executor. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Open Questions

- When will the signed commit be pushed and production deployment proceed?

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_e4e2e360b745af4678884c7b46537238`
