---
title: "IP Whitelisting for Story Protocol RPC Endpoints"
type: "runbook"
slug: "runbooks/ip-whitelisting-story-rpc"
freshness: "2026-01-08T21:13:23Z"
tags:
  - "ip-registration"
  - "mainnet"
  - "networking"
  - "rpc"
  - "testnet"
  - "whitelist"
owners: []
source_revision_ids:
  - "srcrev_0bb0a0e2f8b836be01c7e9e9a94f146f"
  - "srcrev_40c679cdbfd7c5ce8750ebfbdaee9ec8"
  - "srcrev_62a1e4d4e655f4fe6ef11c5b69f63f0e"
  - "srcrev_7fd27afb9c382309d7431322fb422eb5"
  - "srcrev_989a32cb7d35365ca7aa454f09566975"
  - "srcrev_9c9415e8cffc52d44242e8f1e395c148"
  - "srcrev_b1f0b65dba97349b55c2a74d28755631"
  - "srcrev_bf251c861ea343e9f6f53571a1ac9fee"
  - "srcrev_e09bc2ecc992c82e9232de8449fa408a"
  - "srcrev_ed6701507b15036893aea885cd6008e6"
conflict_state: "none"
---

# IP Whitelisting for Story Protocol RPC Endpoints

## Summary

Request to whitelist three IP addresses (3.224.178.198, 34.225.11.164, 34.206.178.84) for Story Protocol's internal RPC endpoints to support IP registration load testing. The IPs correspond to the NAT gateway of the VPC for the depin prod environment. A Linear ticket (SLA-1377) was created, and changes were made via a PR in the piplabs/cloudflare GitHub repository, with the whitelist not being managed by Pulumi.

## Claims

- Whitelist request for three IPs (3.224.178.198, 34.225.11.164, 34.206.178.84) on testnet internal RPC endpoint (https://internal-full.aeneid.storyrpc.io/) for IP registration. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_7fd27afb9c382309d7431322fb422eb5` `chunk_id=srcchunk_92dde4165bcf3d5a5ea9983c38c971ed` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663330.365779` `source_timestamp=2026-01-06T01:35:30Z`
- Public RPC might not handle the load testing, necessitating use of internal RPC. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_ed6701507b15036893aea885cd6008e6` `chunk_id=srcchunk_a0c9e3316ffd3691dfc28581321115bf` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663585.358469` `source_timestamp=2026-01-06T01:39:45Z`
- Linear ticket SLA-1377 created for whitelisting IP addresses for staging environment. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_bf251c861ea343e9f6f53571a1ac9fee` `chunk_id=srcchunk_64b0446e9253bd68c54eb523b9a88737` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663836.099919` `source_timestamp=2026-01-06T01:43:56Z`
- Same three IPs requested to be whitelisted for mainnet RPC endpoint. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_b1f0b65dba97349b55c2a74d28755631` `chunk_id=srcchunk_fdb03f98e665c7adaee2f259e5caa2c6` `native_locator=slack:C0547N89JUB:1767663330.365779:1767827268.836549` `source_timestamp=2026-01-07T23:07:48Z`
- The IP addresses belong to the NAT gateway of the VPC for the depin prod environment. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_e09bc2ecc992c82e9232de8449fa408a` `chunk_id=srcchunk_638484b91a2a0f4a7e6c015048df20ae` `native_locator=slack:C0547N89JUB:1767663330.365779:1767828739.607109` `source_timestamp=2026-01-07T23:32:32Z`
- The whitelist is not managed by Pulumi. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_0bb0a0e2f8b836be01c7e9e9a94f146f` `chunk_id=srcchunk_a0f05504e1b3c92ab5b548de30573a56` `native_locator=slack:C0547N89JUB:1767663330.365779:1767667995.409559` `source_timestamp=2026-01-06T02:53:15Z`
- A PR was created in piplabs/cloudflare (https://github.com/piplabs/cloudflare/pull/104) to effect the whitelist changes. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_9c9415e8cffc52d44242e8f1e395c148` `chunk_id=srcchunk_145ab1c841e1b300711c087d455f4fe2` `native_locator=slack:C0547N89JUB:1767663330.365779:1767836936.100169` `source_timestamp=2026-01-08T01:48:56Z`
- A user was added as admin of the cloudflare repository because it is infra-owned. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_40c679cdbfd7c5ce8750ebfbdaee9ec8` `chunk_id=srcchunk_fd2927701cdd7a1cedd4b8e1a00d44bb` `native_locator=slack:C0547N89JUB:1767663330.365779:1767899775.318109` `source_timestamp=2026-01-08T19:16:15Z`
- The last Pulumi apply failed, but changes were still applied via the PR. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_989a32cb7d35365ca7aa454f09566975` `chunk_id=srcchunk_340c3682a4e7db9d5b689cfb9c7975c6` `native_locator=slack:C0547N89JUB:1767663330.365779:1767899874.255349` `source_timestamp=2026-01-08T19:17:54Z`
- The poseidon NAT gateway IPs were successfully added for RPC. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_62a1e4d4e655f4fe6ef11c5b69f63f0e` `chunk_id=srcchunk_b3a7b97636e7555526e18b1075774d01` `native_locator=slack:C0547N89JUB:1767663330.365779:1767906803.617499` `source_timestamp=2026-01-08T21:13:23Z`

## Open Questions

- Are there any additional IPs that need whitelisting for other environments?
- Is the whitelist change automatically applied via PR merge or does it require manual steps?

## Sources

- `source_document_id`: `srcdoc_f8dd13869d4582b84e34ba2761e3b362`
- `source_revision_id`: `srcrev_b2b10cbf6086949c121a998cf4134a49`
