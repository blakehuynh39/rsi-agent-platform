---
title: "RSI executor commit signing issue"
type: "decision"
slug: "decisions/rsi-executor-commit-signing"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "commit-signing"
  - "github-app"
  - "gpg"
  - "rsiexecutor"
owners: []
source_revision_ids:
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
  - "srcrev_d226b5407b816c642314daa241f600fd"
conflict_state: "none"
---

# RSI executor commit signing issue

## Summary

The RSI executor bot (rsi-platform-bot) cannot sign commits because the GitHub App token lacks signing keys and no GPG/SSH keys are installed on the executor. The commit fa1f78e needs to be amended with a signature by a human operator.

## Claims

- The commit fa1f78e lacks a verified signature, breaking GitHub policy. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_d226b5407b816c642314daa241f600fd` `chunk_id=srcchunk_dffac0c9df4d3ae01ea115f64eb5809d` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593723.167589` `source_timestamp=2026-06-04T17:22:03Z`
- The RSI executor bot cannot sign commits because it runs via a GitHub App token (GH_TOKEN) which can't carry signing keys; GPG, SSH keys, and signing configuration are not installed. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- The easiest fix is for a human operator with push access and a verified GPG/SSH key to checkout the commit, amend with signature, and force push to staging-gcp. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Open Questions

- Who will perform the signed commit?

## Related Pages

- `story-orchestration-nil-pointer-incident`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_8cf79c9aa0d3d09f77eaa5d4312946f3`
