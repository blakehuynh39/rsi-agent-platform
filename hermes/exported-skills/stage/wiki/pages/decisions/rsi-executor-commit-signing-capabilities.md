---
title: "RSI Executor Commit Signing Capabilities"
type: "decision"
slug: "decisions/rsi-executor-commit-signing-capabilities"
freshness: "2026-06-04T17:24:06Z"
tags:
  - "ci"
  - "github-app"
  - "infrastructure"
  - "rsi"
owners: []
source_revision_ids:
  - "srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944"
conflict_state: "none"
---

# RSI Executor Commit Signing Capabilities

## Summary

The RSI executor, running as rsi-platform-bot via GitHub App token, cannot produce signed commits because it lacks cryptographic keys. Resolution options require either manual operator intervention or infrastructure provisioning.

## Claims

- RSI executor runs as rsi-platform-bot[bot] using a GitHub App token (`GH_TOKEN`). `claim:claim_4_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- The container has no GPG/SSH keys and no signing configuration. `claim:claim_4_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- GitHub App installation tokens cannot cryptographically sign commits. `claim:claim_4_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- Easiest fix: a human operator with push access and a verified GPG/SSH key amends the commit with a signature and force-pushes. `claim:claim_4_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`
- Alternative: infrastructure team provisions a GPG keypair and signing config on the RSI executor for future signed commits. `claim:claim_4_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d0b17829b1c0b813fbc88c526641ccaa` `source_revision_id=srcrev_8ae4b48a4c1bf5fc7dad6bbd8c98a944` `chunk_id=srcchunk_9ea9414c0d39b892968b79554e8cd1fd` `native_locator=slack:C08BWTULNPP:1780554037.550049:1780593846.202839` `source_timestamp=2026-06-04T17:24:06Z`

## Open Questions

- Will the infrastructure team provision signing keys for the RSI executor?

## Related Pages

- `commit-signing-requirements`
- `nil-pointer-dereference-in-story-orchestration-service-2026-06-04`

## Sources

- `source_document_id`: `srcdoc_d0b17829b1c0b813fbc88c526641ccaa`
- `source_revision_id`: `srcrev_c324be3df9811270afe5ab85f79a8722`
