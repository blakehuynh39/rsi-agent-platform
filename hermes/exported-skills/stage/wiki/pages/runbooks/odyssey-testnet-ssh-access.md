---
title: "SSH Access to Odyssey Testnet Nodes"
type: "runbook"
slug: "runbooks/odyssey-testnet-ssh-access"
freshness: "2026-05-05T06:42:47Z"
tags:
  - "infrastructure"
  - "odyssey"
  - "ssh"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_f1ff80a873657ab70f42b5093b289a35"
conflict_state: "none"
---

# SSH Access to Odyssey Testnet Nodes

## Summary

SSH access details for bootnodes, validators, and RPC nodes in the Odyssey testnet environment.

## Claims

- The Odyssey testnet has 2 bootnodes: bootnode1 at 3.142.16.95 (us-east-2c) and bootnode2 at 44.235.196.208 (us-west-2d). SSH commands use separate .pem keys per region. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SSH-Information-to-Odyssey-Testnet-115051299a5480e4a1a2f5e0cf6696ce) `source_document_id=srcdoc_5dd5153f6d9091d5987766a60393a519` `source_revision_id=srcrev_f1ff80a873657ab70f42b5093b289a35` `chunk_id=srcchunk_49eafb872bad212e4d11069d7718a0e2` `native_locator=https://www.notion.so/SSH-Information-to-Odyssey-Testnet-115051299a5480e4a1a2f5e0cf6696ce` `source_timestamp=2026-05-05T06:42:47Z`
- There are 8 validators distributed across regions: us-east-2 (3.20.133.100, 52.14.39.177), us-west-2 (50.112.252.101, 54.190.123.194), eu-central-1 (35.157.23.109, 52.58.133.0), ap-northeast-1 (35.74.14.151, 52.198.84.213). Each region uses its own .pem key. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SSH-Information-to-Odyssey-Testnet-115051299a5480e4a1a2f5e0cf6696ce) `source_document_id=srcdoc_5dd5153f6d9091d5987766a60393a519` `source_revision_id=srcrev_f1ff80a873657ab70f42b5093b289a35` `chunk_id=srcchunk_49eafb872bad212e4d11069d7718a0e2` `native_locator=https://www.notion.so/SSH-Information-to-Odyssey-Testnet-115051299a5480e4a1a2f5e0cf6696ce` `source_timestamp=2026-05-05T06:42:47Z`
- RPC nodes include 3 internal AWS instances (3.16.175.31, 3.146.164.199, 3.140.224.188, all with archive option on the third) using the us-east-2 key, and 4 external Velia instances (148.72.138.185-188) using a separate velia_key. SSH access uses ec2-user for AWS and ubuntu for Velia. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SSH-Information-to-Odyssey-Testnet-115051299a5480e4a1a2f5e0cf6696ce) `source_document_id=srcdoc_5dd5153f6d9091d5987766a60393a519` `source_revision_id=srcrev_f1ff80a873657ab70f42b5093b289a35` `chunk_id=srcchunk_49eafb872bad212e4d11069d7718a0e2` `native_locator=https://www.notion.so/SSH-Information-to-Odyssey-Testnet-115051299a5480e4a1a2f5e0cf6696ce` `source_timestamp=2026-05-05T06:42:47Z`

## Sources

- `source_document_id`: `srcdoc_5dd5153f6d9091d5987766a60393a519`
- `source_revision_id`: `srcrev_f1ff80a873657ab70f42b5093b289a35`
- `source_url`: [Notion source](https://www.notion.so/SSH-Information-to-Odyssey-Testnet-115051299a5480e4a1a2f5e0cf6696ce)
