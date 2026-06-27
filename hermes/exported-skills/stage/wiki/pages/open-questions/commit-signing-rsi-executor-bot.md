---
title: "Open Question: Commit Signing for RSI Executor Bot"
type: "open_question"
slug: "open-questions/commit-signing-rsi-executor-bot"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "commit-signing"
  - "github-app"
  - "gpg"
  - "infrastructure"
  - "rsi-platform-bot"
owners:
  - "Infrastructure Team"
source_revision_ids:
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
conflict_state: "none"
---

# Open Question: Commit Signing for RSI Executor Bot

## Summary

The RSI executor bot (rsi-platform-bot[bot]) runs as a GitHub App and cannot sign commits due to lack of GPG keys or SSH keys. The GitHub App model does not support commit signing. A human operator is needed to sign the current commit, or the infrastructure team could provision signing keys on the executor for future commits.

## Claims

- The RSI executor runs as rsi-platform-bot via GitHub App token, and its container lacks GPG, SSH keys, keyring, and git signingkey configuration. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- GitHub App installation tokens cannot cryptographically sign commits. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- The commit fa1f78e was created with author root@... `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- Easiest fix: a human operator with push access and a verified GPG/SSH key can amend and force-push the commit with signing. `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- Alternative: infrastructure team provisions a GPG keypair and signing config on the RSI executor for future commits. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Related Pages

- `nil-pointer-dereference-story-orchestration-service`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_7cd9b4be7aa48aaa3ca0433558428118`
