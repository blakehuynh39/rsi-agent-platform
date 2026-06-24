---
title: "CloudFlare DNS Update for BitGo Staking-API"
type: "decision"
slug: "decisions/bitgo-staking-api-cloudflare-dns-update"
freshness: "2026-01-24T01:28:12Z"
tags:
  - "bitgo"
  - "cloudflare"
  - "dns"
  - "faucet"
  - "iac"
  - "staking-api"
owners:
  - "U079ZJ48D62"
  - "U08332YRB7W"
source_revision_ids:
  - "srcrev_19cb0ae22fbd9d6102b9d4c7060e795c"
  - "srcrev_2c1a4582cebfe9d88bc322982b558c48"
  - "srcrev_9b3b622d20ff9f571bd8673906f6b6fb"
  - "srcrev_a0780aea4cdb9cb309087410ccba314a"
conflict_state: "none"
---

# CloudFlare DNS Update for BitGo Staking-API

## Summary

PR #115 on piplabs/cloudflare updates DNS records for the private fork of staking-api for BitGo and includes faucet configuration maintained as Infrastructure as Code (IaC). The PR was reviewed and approved. The changes are not to be shared with BitGo until indexing is complete.

## Claims

- PR #115 updates CloudFlare DNS for the private-fork of staking-api for BitGo. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d` `source_revision_id=srcrev_19cb0ae22fbd9d6102b9d4c7060e795c` `chunk_id=srcchunk_99b8af64876a114fa40eabf3137be474` `native_locator=slack:C0547N89JUB:1769217935.107139:1769217935.107139` `source_timestamp=2026-01-24T01:25:35Z`
- The PR also includes faucet information to maintain it in IaC instead of manual changes. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d` `source_revision_id=srcrev_19cb0ae22fbd9d6102b9d4c7060e795c` `chunk_id=srcchunk_99b8af64876a114fa40eabf3137be474` `native_locator=slack:C0547N89JUB:1769217935.107139:1769217935.107139` `source_timestamp=2026-01-24T01:25:35Z`
- The PR changes will not be shared with BitGo until indexing is done. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d` `source_revision_id=srcrev_9b3b622d20ff9f571bd8673906f6b6fb` `chunk_id=srcchunk_d13ca4f76a4ac0996697b83cef56b0f1` `native_locator=slack:C0547N89JUB:1769217935.107139:1769217975.824939` `source_timestamp=2026-01-24T01:26:15Z`
- The PR was checked and approved. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d` `source_revision_id=srcrev_a0780aea4cdb9cb309087410ccba314a` `chunk_id=srcchunk_93a30f5223666c3c137eaffbcca253b3` `native_locator=slack:C0547N89JUB:1769217935.107139:1769218042.801209` `source_timestamp=2026-01-24T01:27:22Z`
- The reviewer thanked the PR author. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d` `source_revision_id=srcrev_2c1a4582cebfe9d88bc322982b558c48` `chunk_id=srcchunk_a146206abed2bd33783e126a35e7774f` `native_locator=slack:C0547N89JUB:1769217935.107139:1769218092.184969` `source_timestamp=2026-01-24T01:28:12Z`

## Sources

- `source_document_id`: `srcdoc_3e9c84a7dd8e09f8e126de182ea61c5d`
- `source_revision_id`: `srcrev_9b3b622d20ff9f571bd8673906f6b6fb`
