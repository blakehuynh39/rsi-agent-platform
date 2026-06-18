---
title: "RSI Executor Commit Signing"
type: "runbook"
slug: "runbooks/rsi-executor-commit-signing"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "commit-signing"
  - "github"
  - "gpg"
  - "rsi-executor"
  - "ssh"
owners:
  - "Infrastructure"
  - "Platform Engineering"
source_revision_ids:
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
conflict_state: "none"
---

# RSI Executor Commit Signing

## Summary

The RSI executor (rsi-platform-bot) cannot sign commits due to missing GPG/SSH keys and GitHub App limitations. To satisfy the verified signature policy, either a human operator must sign existing commits or the executor must be provisioned with signing capabilities.

## Claims

- The RSI executor is running as rsi-platform-bot[bot] via a GitHub App token (GH_TOKEN). `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- The executor's container has no GPG (gpg, gpg2) installed, empty ~/.ssh/, no ~/.gnupg/, and no git config user.signingkey. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- GitHub App installation tokens cannot cryptographically sign commits. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- The commit fa1f78e was created via direct git commit on the executor with author root@... `claim:claim_2_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- A human operator with push access and a verified GPG/SSH key can sign the commit by checking out fa1f78e, amending with -S, and force-pushing to staging-gcp. `claim:claim_2_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- Alternatively, the infrastructure team could provision a GPG keypair and signing config on the RSI executor so future commits are signed at creation time. `claim:claim_2_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Related Pages

- `incident-nil-pointer-deref-story-orchestration`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944`
