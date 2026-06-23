---
title: "Whitelisting IPs for RPC Endpoints"
type: "runbook"
slug: "runbooks/whitelisting-ips-for-rpc-endpoints"
freshness: "2026-01-08T21:13:23Z"
tags:
  - "depin"
  - "infrastructure"
  - "ip-registration"
  - "rpc"
  - "whitelist"
owners:
  - "@U04KTUN5WFQ"
  - "@U07A7AUGL5V"
  - "@U08332YRB7W"
  - "@U09M2SPUTSL"
source_revision_ids:
  - "srcrev_009e3c25a1ef48f9cc209d4da2279f85"
  - "srcrev_0bb0a0e2f8b836be01c7e9e9a94f146f"
  - "srcrev_2611edb877acb2841c3bbaa541cc60c1"
  - "srcrev_2728740a46f96469b5a674253df6cdc9"
  - "srcrev_3f0486431ca0fc66badc95462b5bc9e7"
  - "srcrev_40c679cdbfd7c5ce8750ebfbdaee9ec8"
  - "srcrev_62a1e4d4e655f4fe6ef11c5b69f63f0e"
  - "srcrev_7fd27afb9c382309d7431322fb422eb5"
  - "srcrev_989a32cb7d35365ca7aa454f09566975"
  - "srcrev_9c9415e8cffc52d44242e8f1e395c148"
  - "srcrev_ad10cf0e50e1dd10df0cb86fa78b426c"
  - "srcrev_b1f0b65dba97349b55c2a74d28755631"
  - "srcrev_b2b10cbf6086949c121a998cf4134a49"
  - "srcrev_bf251c861ea343e9f6f53571a1ac9fee"
  - "srcrev_e09bc2ecc992c82e9232de8449fa408a"
  - "srcrev_ed6701507b15036893aea885cd6008e6"
  - "srcrev_f2267ba63434a31584aa83396c908666"
conflict_state: "none"
---

# Whitelisting IPs for RPC Endpoints

## Summary

Procedure for whitelisting IP addresses for restricted internal RPC endpoints, including testnet (aeneid) and mainnet (poseidon). Involves creating a Linear ticket and submitting a PR to piplabs/cloudflare to update whitelists not managed by Pulumi.

## Claims

- Three IP addresses (3.224.178.198, 34.225.11.164, 34.206.178.84) need to be whitelisted for the aeneid testnet internal RPC endpoint (https://internal-full.aeneid.storyrpc.io/) for IP registrations. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_7fd27afb9c382309d7431322fb422eb5` `chunk_id=srcchunk_92dde4165bcf3d5a5ea9983c38c971ed` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663330.365779` `source_timestamp=2026-01-06T01:35:30Z`
- High QPS is expected for the IP registrations, and the public RPC might not handle the load testing. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_2728740a46f96469b5a674253df6cdc9` `chunk_id=srcchunk_720f0e20f92f6f0fd8de2d1f09267b95` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663399.327019` `source_timestamp=2026-01-06T01:36:39Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_2611edb877acb2841c3bbaa541cc60c1` `chunk_id=srcchunk_766faff9e7d6d98a138ecf289a8e4158` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663425.746339` `source_timestamp=2026-01-06T01:37:05Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_ed6701507b15036893aea885cd6008e6` `chunk_id=srcchunk_a0c9e3316ffd3691dfc28581321115bf` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663585.358469` `source_timestamp=2026-01-06T01:39:45Z`
- A Linear ticket SLA-1377 was created for whitelisting IP addresses of IP registration worker for staging environment. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_b2b10cbf6086949c121a998cf4134a49` `chunk_id=srcchunk_2ae1d985f3060b3edd11e6a7dcfb03ac` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663608.982829` `source_timestamp=2026-01-06T01:40:08Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_bf251c861ea343e9f6f53571a1ac9fee` `chunk_id=srcchunk_64b0446e9253bd68c54eb523b9a88737` `native_locator=slack:C0547N89JUB:1767663330.365779:1767663836.099919` `source_timestamp=2026-01-06T01:43:56Z`
- The whitelist is not managed by Pulumi; configuration changes are made via PRs to the piplabs/cloudflare repository. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_009e3c25a1ef48f9cc209d4da2279f85` `chunk_id=srcchunk_1b769099ea4dfff36658d5332eac8249` `native_locator=slack:C0547N89JUB:1767663330.365779:1767667964.756519` `source_timestamp=2026-01-06T02:52:44Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_0bb0a0e2f8b836be01c7e9e9a94f146f` `chunk_id=srcchunk_a0f05504e1b3c92ab5b548de30573a56` `native_locator=slack:C0547N89JUB:1767663330.365779:1767667995.409559` `source_timestamp=2026-01-06T02:53:15Z`
- Same IPs were requested for mainnet RPC endpoint (poseidon), identified as NAT gateway IPs of the VPC for the depin prod environment. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_b1f0b65dba97349b55c2a74d28755631` `chunk_id=srcchunk_fdb03f98e665c7adaee2f259e5caa2c6` `native_locator=slack:C0547N89JUB:1767663330.365779:1767827268.836549` `source_timestamp=2026-01-07T23:07:48Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_ad10cf0e50e1dd10df0cb86fa78b426c` `chunk_id=srcchunk_711ac5ca97719c08b121997d7f40f2cd` `native_locator=slack:C0547N89JUB:1767663330.365779:1767828441.159769` `source_timestamp=2026-01-07T23:27:23Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_e09bc2ecc992c82e9232de8449fa408a` `chunk_id=srcchunk_638484b91a2a0f4a7e6c015048df20ae` `native_locator=slack:C0547N89JUB:1767663330.365779:1767828739.607109` `source_timestamp=2026-01-07T23:32:32Z`
- A PR (https://github.com/piplabs/cloudflare/pull/104) was created to add the IPs; after resolving reviewer access issues and a failed Pulumi apply, the whitelist was updated successfully. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_9c9415e8cffc52d44242e8f1e395c148` `chunk_id=srcchunk_145ab1c841e1b300711c087d455f4fe2` `native_locator=slack:C0547N89JUB:1767663330.365779:1767836936.100169` `source_timestamp=2026-01-08T01:48:56Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_3f0486431ca0fc66badc95462b5bc9e7` `chunk_id=srcchunk_6ec278da864abda56e93b8754b18f784` `native_locator=slack:C0547N89JUB:1767663330.365779:1767899573.151419` `source_timestamp=2026-01-08T19:12:53Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_f2267ba63434a31584aa83396c908666` `chunk_id=srcchunk_c17f1ea2466b9d530060f09b0fc46a2d` `native_locator=slack:C0547N89JUB:1767663330.365779:1767899696.387679` `source_timestamp=2026-01-08T19:14:56Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_40c679cdbfd7c5ce8750ebfbdaee9ec8` `chunk_id=srcchunk_fd2927701cdd7a1cedd4b8e1a00d44bb` `native_locator=slack:C0547N89JUB:1767663330.365779:1767899775.318109` `source_timestamp=2026-01-08T19:16:15Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_989a32cb7d35365ca7aa454f09566975` `chunk_id=srcchunk_340c3682a4e7db9d5b689cfb9c7975c6` `native_locator=slack:C0547N89JUB:1767663330.365779:1767899874.255349` `source_timestamp=2026-01-08T19:17:54Z`
  - citation: `source_document_id=srcdoc_f8dd13869d4582b84e34ba2761e3b362` `source_revision_id=srcrev_62a1e4d4e655f4fe6ef11c5b69f63f0e` `chunk_id=srcchunk_b3a7b97636e7555526e18b1075774d01` `native_locator=slack:C0547N89JUB:1767663330.365779:1767906803.617499` `source_timestamp=2026-01-08T21:13:23Z`

## Sources

- `source_document_id`: `srcdoc_f8dd13869d4582b84e34ba2761e3b362`
- `source_revision_id`: `srcrev_2611edb877acb2841c3bbaa541cc60c1`
