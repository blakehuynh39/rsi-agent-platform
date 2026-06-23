---
title: "IP Whitelisting for Internal RPC Endpoints"
type: "runbook"
slug: "runbooks/ip-whitelisting-internal-rpc"
freshness: "2026-01-08T21:13:23Z"
tags:
  - "ip"
  - "mainnet"
  - "rpc"
  - "story-rpc"
  - "testnet"
  - "whitelist"
owners:
  - "U07A7AUGL5V"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_0bb0a0e2f8b836be01c7e9e9a94f146f"
  - "srcrev_2611edb877acb2841c3bbaa541cc60c1"
  - "srcrev_2728740a46f96469b5a674253df6cdc9"
  - "srcrev_40c679cdbfd7c5ce8750ebfbdaee9ec8"
  - "srcrev_62a1e4d4e655f4fe6ef11c5b69f63f0e"
  - "srcrev_7fd27afb9c382309d7431322fb422eb5"
  - "srcrev_b1f0b65dba97349b55c2a74d28755631"
  - "srcrev_bf251c861ea343e9f6f53571a1ac9fee"
  - "srcrev_e09bc2ecc992c82e9232de8449fa408a"
  - "srcrev_ed6701507b15036893aea885cd6008e6"
conflict_state: "none"
---

# IP Whitelisting for Internal RPC Endpoints

## Summary

Procedure for whitelisting IP addresses to access restricted internal RPC endpoints for testnet and mainnet, used for IP registration load testing. Not managed by Pulumi; manual changes via cloudflare repo.

## Claims

- A request was made to whitelist IP addresses 3.224.178.198, 34.225.11.164, and 34.206.178.84 for the testnet restricted internal RPC endpoint (https://internal-full.aeneid.storyrpc.io/). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_7fd27afb9c382309d7431322fb422eb5` `chunk_id=srcchunk_92dde4165bcf3d5a5ea9983c38c971ed` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663330.365779` `source_timestamp=2026-01-06T01:35:30Z`
- Public RPC is insufficient for expected high QPS from IP registration load testing, necessitating the internal endpoint. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_ed6701507b15036893aea885cd6008e6` `chunk_id=srcchunk_a0c9e3316ffd3691dfc28581321115bf` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663585.358469` `source_timestamp=2026-01-06T01:39:45Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_2728740a46f96469b5a674253df6cdc9` `chunk_id=srcchunk_720f0e20f92f6f0fd8de2d1f09267b95` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663399.327019` `source_timestamp=2026-01-06T01:36:39Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_2611edb877acb2841c3bbaa541cc60c1` `chunk_id=srcchunk_766faff9e7d6d98a138ecf289a8e4158` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663425.746339` `source_timestamp=2026-01-06T01:37:05Z`
- A Linear ticket SLA-1377 was created to track the whitelisting request. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_bf251c861ea343e9f6f53571a1ac9fee` `chunk_id=srcchunk_64b0446e9253bd68c54eb523b9a88737` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663836.099919` `source_timestamp=2026-01-06T01:43:56Z`
- The whitelist for the internal RPC endpoints is not managed via Pulumi. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_0bb0a0e2f8b836be01c7e9e9a94f146f` `chunk_id=srcchunk_a0f05504e1b3c92ab5b548de30573a56` `native_locator=slack:C0547N89JUB:1767663330.365779:1767667995.409559` `source_timestamp=2026-01-06T02:53:15Z`
- The same three IPs were later required for mainnet RPC endpoint, identified as NAT gateway IPs of the depin prod VPC. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_b1f0b65dba97349b55c2a74d28755631` `chunk_id=srcchunk_fdb03f98e665c7adaee2f259e5caa2c6` `native_locator=slack:C0547N89JUB:1767663330.365779:1767827268.836549` `source_timestamp=2026-01-07T23:07:48Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_e09bc2ecc992c82e9232de8449fa408a` `chunk_id=srcchunk_638484b91a2a0f4a7e6c015048df20ae` `native_locator=slack:C0547N89JUB:1767663330.365779:1767828739.607109` `source_timestamp=2026-01-07T23:32:32Z`
- The whitelisting for mainnet was performed by U09M2SPUTSL after gaining admin access to the cloudflare repository via PR #104. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_40c679cdbfd7c5ce8750ebfbdaee9ec8` `chunk_id=srcchunk_fd2927701cdd7a1cedd4b8e1a00d44bb` `native_locator=slack:C0547N89JUB:1767663330.365779:1767899775.318109` `source_timestamp=2026-01-08T19:16:15Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_62a1e4d4e655f4fe6ef11c5b69f63f0e` `chunk_id=srcchunk_b3a7b97636e7555526e18b1075774d01` `native_locator=slack:C0547N89JUB:1767663330.365779:1767906803.617499` `source_timestamp=2026-01-08T21:13:23Z`

## Sources

- `source_document_id`: `srcdoc_f8dd13869d4582b84e34ba2761e3b362`
- `source_revision_id`: `srcrev_7fd27afb9c382309d7431322fb422eb5`
