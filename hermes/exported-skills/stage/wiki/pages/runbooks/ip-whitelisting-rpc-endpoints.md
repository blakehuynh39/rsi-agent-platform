---
title: "IP Whitelisting for RPC Endpoints"
type: "runbook"
slug: "runbooks/ip-whitelisting-rpc-endpoints"
freshness: "2026-01-08T21:13:23Z"
tags:
  - "cloudflare"
  - "infra"
  - "ip"
  - "rpc"
  - "whitelist"
owners:
  - "U04KTUN5WFQ"
  - "U07A7AUGL5V"
  - "U08332YRB7W"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_0bb0a0e2f8b836be01c7e9e9a94f146f"
  - "srcrev_2d247918ce1c1c368104f6f5309684cf"
  - "srcrev_62a1e4d4e655f4fe6ef11c5b69f63f0e"
  - "srcrev_7fd27afb9c382309d7431322fb422eb5"
  - "srcrev_9c9415e8cffc52d44242e8f1e395c148"
  - "srcrev_bf251c861ea343e9f6f53571a1ac9fee"
  - "srcrev_e09bc2ecc992c82e9232de8449fa408a"
conflict_state: "none"
---

# IP Whitelisting for RPC Endpoints

## Summary

Procedure for requesting and implementing IP whitelist additions for Story Protocol RPC endpoints.

## Claims

- The following IP addresses were requested for whitelisting on the testnet RPC endpoint: 3.224.178.198, 34.225.11.164, 34.206.178.84. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_7fd27afb9c382309d7431322fb422eb5` `chunk_id=srcchunk_92dde4165bcf3d5a5ea9983c38c971ed` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663330.365779` `source_timestamp=2026-01-06T01:35:30Z`
- A Linear ticket SLA-1377 was created to whitelist IP addresses for the staging environment. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_bf251c861ea343e9f6f53571a1ac9fee` `chunk_id=srcchunk_64b0446e9253bd68c54eb523b9a88737` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663836.099919` `source_timestamp=2026-01-06T01:43:56Z`
- The IPs are the NAT gateway IPs of the VPC for the depin prod environment. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_e09bc2ecc992c82e9232de8449fa408a` `chunk_id=srcchunk_638484b91a2a0f4a7e6c015048df20ae` `native_locator=slack:C0547N89JUB:1767663330.365779:1767828739.607109` `source_timestamp=2026-01-07T23:32:32Z`
- Whitelisting for mainnet was completed by adding the Poseidon NAT gateway IPs. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_62a1e4d4e655f4fe6ef11c5b69f63f0e` `chunk_id=srcchunk_b3a7b97636e7555526e18b1075774d01` `native_locator=slack:C0547N89JUB:1767663330.365779:1767906803.617499` `source_timestamp=2026-01-08T21:13:23Z`
- A Cloudflare configuration pull request (https://github.com/piplabs/cloudflare/pull/104) was opened to implement the whitelisting. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_9c9415e8cffc52d44242e8f1e395c148` `chunk_id=srcchunk_145ab1c841e1b300711c087d455f4fe2` `native_locator=slack:C0547N89JUB:1767663330.365779:1767836936.100169` `source_timestamp=2026-01-08T01:48:56Z`
- These IP whitelists are not managed by Pulumi. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_0bb0a0e2f8b836be01c7e9e9a94f146f` `chunk_id=srcchunk_a0f05504e1b3c92ab5b548de30573a56` `native_locator=slack:C0547N89JUB:1767663330.365779:1767667995.409559` `source_timestamp=2026-01-06T02:53:15Z`
- The whitelisting was successfully implemented and tested. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_2d247918ce1c1c368104f6f5309684cf` `chunk_id=srcchunk_ac8c2350eb61c36f79e85d14e18825aa` `native_locator=slack:C0547N89JUB:1767663330.365779:1767667920.630199` `source_timestamp=2026-01-06T02:52:00Z`

## Sources

- `source_document_id`: `srcdoc_f8dd13869d4582b84e34ba2761e3b362`
- `source_revision_id`: `srcrev_bf251c861ea343e9f6f53571a1ac9fee`
