---
title: "story-orchestration-service nil pointer dereference runbook"
type: "runbook"
slug: "runbooks/story-orchestration-service-nil-pointer-dereference-runbook"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "debugging"
  - "incident"
  - "nil-pointer"
  - "story-orchestration-service"
owners:
  - "U0ASDQKU3UL"
source_revision_ids:
  - "srcrev_61c6bf7d60075b0ddbc61f685c20567c"
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
  - "srcrev_c324be3df9811270afe5ab85f79a8722"
conflict_state: "none"
---

# story-orchestration-service nil pointer dereference runbook

## Summary

Runbook for diagnosing and fixing nil pointer dereference errors in story-orchestration-service, based on an incident starting Feb 12, 2026.

## Claims

- story-orchestration-service in production experienced a nil pointer dereference error, with 2,046 events since Feb 12. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_61c6bf7d60075b0ddbc61f685c20567c` `chunk_id=srcchunk_bb77c6e2bdd54fb93f73148352c63993` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780554037.550049` `source_timestamp=2026-06-04T06:20:37Z`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Root cause is GetJobByWorkflowID returning (nil, nil) for non-existent workflow IDs, and three Temporal workflow handler methods dereferencing the nil pointer. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- Fix was committed in PR #718 at 07:11 UTC, deployed to staging but not yet to production. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_c324be3df9811270afe5ab85f79a8722` `chunk_id=srcchunk_b95a10c83aa5bd65c9f39a84ee0bab98` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593420.294409` `source_timestamp=2026-06-04T17:17:00Z`
- RSI executor cannot sign commits due to lack of GPG/SSH keys; external operator required to sign and push the commit. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Open Questions

- When will the fix be deployed to production?

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_031befdaa44a6d6dd99bb5a631b97e31`
