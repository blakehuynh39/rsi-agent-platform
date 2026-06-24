---
title: "CloudFlare DNS IaC Update for Staking-API (BitGo)"
type: "decision"
slug: "decisions/cloudflare-dns-iac-staking-api-bitgo"
freshness: "2026-01-24T01:27:22Z"
tags:
  - "bitgo"
  - "cloudflare"
  - "dns"
  - "faucet"
  - "infrastructure-as-code"
  - "staking-api"
owners: []
source_revision_ids:
  - "srcrev_19cb0ae22fbd9d6102b9d4c7060e795c"
  - "srcrev_9b3b622d20ff9f571bd8673906f6b6fb"
  - "srcrev_a0780aea4cdb9cb309087410ccba314a"
conflict_state: "none"
---

# CloudFlare DNS IaC Update for Staking-API (BitGo)

## Summary

Approved Infrastructure-as-Code (IaC) update for CloudFlare DNS to support the private-fork of staking-api for BitGo, including faucet configuration. The PR is not to be shared with BitGo until indexing is complete, and it allows monitoring indexing progress.

## Claims

- A pull request (https://github.com/piplabs/cloudflare/pull/115/files) was created to update CloudFlare DNS for the private-fork of staking-api for BitGo, including faucet configuration to be maintained as Infrastructure-as-Code (IaC) rather than manually. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d` `source_revision_id=srcrev_19cb0ae22fbd9d6102b9d4c7060e795c` `chunk_id=srcchunk_99b8af64876a114fa40eabf3137be474` `native_locator=slack:C0547N89JUB:1769217935.107139:1769217935.107139` `source_timestamp=2026-01-24T01:25:35Z`
- The PR should not be shared with BitGo until indexing is complete, but it allows monitoring the progress of the indexing. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d` `source_revision_id=srcrev_9b3b622d20ff9f571bd8673906f6b6fb` `chunk_id=srcchunk_d13ca4f76a4ac0996697b83cef56b0f1` `native_locator=slack:C0547N89JUB:1769217935.107139:1769217975.824939` `source_timestamp=2026-01-24T01:26:15Z`
- The pull request was reviewed and approved. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d` `source_revision_id=srcrev_a0780aea4cdb9cb309087410ccba314a` `chunk_id=srcchunk_93a30f5223666c3c137eaffbcca253b3` `native_locator=slack:C0547N89JUB:1769217935.107139:1769218042.801209` `source_timestamp=2026-01-24T01:27:22Z`

## Open Questions

- What is the current indexing status for BitGo staking-api? When will indexing be complete and the DNS change be shared with BitGo?

## Sources

- `source_document_id`: `srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d`
- `source_revision_id`: `srcrev_a0780aea4cdb9cb309087410ccba314a`
