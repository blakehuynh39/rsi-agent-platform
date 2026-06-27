---
title: "Commit Signing Requirements"
type: "policy"
slug: "policies/commit-signing-requirements"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "compliance"
  - "github"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
  - "srcrev_d226b5407b816c642314daa241f600fd"
conflict_state: "none"
---

# Commit Signing Requirements

## Summary

All commits pushed to GitHub repositories must carry verified cryptographic signatures, enforced by GitHub branch protection policies.

## Claims

- Commits lacking a verified signature break GitHub policy. `claim:claim_3_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_d226b5407b816c642314daa241f600fd` `chunk_id=srcchunk_dffac0c9df4d3ae01ea115f64eb5809d` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593723.167589` `source_timestamp=2026-06-04T17:22:03Z`
- Commit fa1f78e (staging-gcp) lacked a verified signature. `claim:claim_3_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Related Pages

- `nil-pointer-dereference-in-story-orchestration-service-2026-06-04`
- `rsi-executor-commit-signing-capabilities`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_c324be3df9811270afe5ab85f79a8722`
