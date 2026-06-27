---
title: "RSI executor commit signing policy"
type: "decision"
slug: "decisions/rsi-executor-commit-signing-policy"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "git"
  - "github-policy"
  - "rsi-executor"
  - "signing"
owners: []
source_revision_ids:
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
  - "srcrev_d226b5407b816c642314daa241f600fd"
conflict_state: "none"
---

# RSI executor commit signing policy

## Summary

The RSI executor bot made a commit (fa1f78e) without a verified signature, violating GitHub policy. The executor lacks GPG/SSH keys. Options: human operator amend and sign, or infra team provision keys on executor.

## Claims

- Commit fa1f78e does not have a verified signature and breaks GitHub policy. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_d226b5407b816c642314daa241f600fd` `chunk_id=srcchunk_dffac0c9df4d3ae01ea115f64eb5809d` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593723.167589` `source_timestamp=2026-06-04T17:22:03Z`
- RSI executor cannot sign commits; container lacks GPG (`gpg`, `gpg2`), SSH keys, GPG keyring, and `git config user.signingkey`. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- Easiest fix: a human operator with push access and a verified GPG/SSH key does: `git checkout fa1f78e`, `git commit --amend -S --no-edit`, `git push --force-with-lease origin staging-gcp`. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- Alternative: infrastructure team could provision a GPG keypair and signing config on the RSI executor so future commits are signed at creation time. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Open Questions

- Who is the designated human operator with push access to sign the commit?

## Related Pages

- `story-orchestration-nil-pointer-incident`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_a458a652d00ef32049a3cd6ee8e18b56`
