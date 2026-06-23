---
title: "Mainnet AWS ACM PR #110 Review"
type: "decision"
slug: "decisions/mainnet-aws-acm-pr-110"
freshness: "2026-01-12T05:59:36Z"
tags:
  - "aws-acm"
  - "cloudflare"
  - "github-pr"
  - "mainnet"
owners: []
source_revision_ids:
  - "srcrev_3d091926360ae443240c310d605f3776"
  - "srcrev_8eef39af3b7be19d36097a5229384f14"
  - "srcrev_afab0c9db366e1acface3fd68beb0829"
  - "srcrev_b9aa157cf06895b67e1f1b6ae4abd468"
  - "srcrev_d736feba8db17a6edada3ae2d694f6c2"
  - "srcrev_e8b49445d809cfc01ab8229703c6a2fe"
conflict_state: "none"
---

# Mainnet AWS ACM PR #110 Review

## Summary

Tracks the review and approval of pull request #110 in the piplabs/cloudflare repository, which introduces AWS Certificate Manager changes for the Mainnet environment. The PR was reviewed by the cloudflare team subteam and two individual engineers, and it was confirmed that the change does not affect existing Mainnet RPC usage.

## Claims

- A pull request for Mainnet AWS ACM was created in the piplabs/cloudflare repository at https://github.com/piplabs/cloudflare/pull/110. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d3742e25b3f88fb877417f22d7906176` `source_revision_id=srcrev_e8b49445d809cfc01ab8229703c6a2fe` `chunk_id=srcchunk_62edc47dd64e98c738a76524d577254d` `native_locator=slack:C0547N89JUB:1768190125.582039:1768190125.582039` `source_timestamp=2026-01-12T03:55:25Z`
- The PR review was requested from the subteam <!subteam^S083BDZ4FTM>. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d3742e25b3f88fb877417f22d7906176` `source_revision_id=srcrev_e8b49445d809cfc01ab8229703c6a2fe` `chunk_id=srcchunk_62edc47dd64e98c738a76524d577254d` `native_locator=slack:C0547N89JUB:1768190125.582039:1768190125.582039` `source_timestamp=2026-01-12T03:55:25Z`
- Additionally, <@U08332YRB7W> and <@U080YAW205V> were specifically asked to review the PR. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d3742e25b3f88fb877417f22d7906176` `source_revision_id=srcrev_d736feba8db17a6edada3ae2d694f6c2` `chunk_id=srcchunk_66958edba5b4ee95301ce4f0a70f4374` `native_locator=slack:C0547N89JUB:1768190125.582039:1768192853.838069` `source_timestamp=2026-01-12T04:40:53Z`
- <@U07TNT9N4JC> asked whether the PR affects the usage of current mainnet RPCs. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d3742e25b3f88fb877417f22d7906176` `source_revision_id=srcrev_8eef39af3b7be19d36097a5229384f14` `chunk_id=srcchunk_792082f7ce02c2b6b4b7c29e105e893f` `native_locator=slack:C0547N89JUB:1768190125.582039:1768192931.635629` `source_timestamp=2026-01-12T04:42:18Z`
- It was confirmed that the PR does not affect the usage of current mainnet RPCs. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d3742e25b3f88fb877417f22d7906176` `source_revision_id=srcrev_b9aa157cf06895b67e1f1b6ae4abd468` `chunk_id=srcchunk_af2176806f5509c80c679d48df58a320` `native_locator=slack:C0547N89JUB:1768190125.582039:1768193089.311329` `source_timestamp=2026-01-12T04:44:49Z`
- An approval comment 'I approved from my side' was posted in the thread. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d3742e25b3f88fb877417f22d7906176` `source_revision_id=srcrev_afab0c9db366e1acface3fd68beb0829` `chunk_id=srcchunk_917a1b6087bce394066b91aae33899e4` `native_locator=slack:C0547N89JUB:1768190125.582039:1768193246.256769` `source_timestamp=2026-01-12T04:47:26Z`
- Another approval comment 'I had approved too' was posted in the thread. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d3742e25b3f88fb877417f22d7906176` `source_revision_id=srcrev_3d091926360ae443240c310d605f3776` `chunk_id=srcchunk_298648a6975ce97ddbe364aab8d9ce48` `native_locator=slack:C0547N89JUB:1768190125.582039:1768197576.577249` `source_timestamp=2026-01-12T05:59:36Z`

## Open Questions

- Has the PR been merged?
- Who specifically approved (mapping of user IDs to names)?

## Sources

- `source_document_id`: `srcdoc_d3742e25b3f88fb877417f22d7906176`
- `source_revision_id`: `srcrev_d736feba8db17a6edada3ae2d694f6c2`
