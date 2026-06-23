---
title: "Approve PR #108 for Devnet Domain on AWS"
type: "decision"
slug: "decisions/pr-108-devnet-domain-approval"
freshness: "2026-01-10T01:13:45Z"
tags:
  - "aws"
  - "certificate"
  - "cloudflare"
  - "devnet"
  - "domain"
  - "pr"
owners:
  - "S083BDZ4FTM"
  - "U07TNT9N4JC"
source_revision_ids:
  - "srcrev_22b050154e03ae53746dcaa93e38eec7"
  - "srcrev_5aea57a3d490ab35d7dc7e5e2e50dc8a"
conflict_state: "none"
---

# Approve PR #108 for Devnet Domain on AWS

## Summary

PR #108 in piplabs/cloudflare, intended to configure a domain for devnet on AWS, was approved by U07TNT9N4JC. The approver could not verify the ACM certificate request in AWS ACM.

## Claims

- A PR (pull request) #108 was created in the piplabs/cloudflare repository for domain configuration on AWS for devnet. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41cbb4fa0659a483ae953fc04716638a` `source_revision_id=srcrev_5aea57a3d490ab35d7dc7e5e2e50dc8a` `chunk_id=srcchunk_3c3d86234c66eb2053c713231f2d02cf` `native_locator=slack:C0547N89JUB:1768005433.481429:1768005433.481429` `source_timestamp=2026-01-10T00:37:13Z`
- User U07TNT9N4JC approved PR #108, but indicated inability to see/verify the cert request in AWS ACM. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41cbb4fa0659a483ae953fc04716638a` `source_revision_id=srcrev_22b050154e03ae53746dcaa93e38eec7` `chunk_id=srcchunk_746d87b7d9459c8e4318c02a172a1825` `native_locator=slack:C0547N89JUB:1768005433.481429:1768007625.107379` `source_timestamp=2026-01-10T01:13:45Z`

## Open Questions

- Did the PR ultimately get merged?
- Was the certificate request created in ACM?
- What is the exact domain being configured?

## Sources

- `source_document_id`: `srcdoc_41cbb4fa0659a483ae953fc04716638a`
- `source_revision_id`: `srcrev_22b050154e03ae53746dcaa93e38eec7`
