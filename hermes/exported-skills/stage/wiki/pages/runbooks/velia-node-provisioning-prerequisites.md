---
title: "Provision RPC Nodes on Velia - Prerequisites"
type: "runbook"
slug: "runbooks/velia-node-provisioning-prerequisites"
freshness: "2024-10-11T22:39:00Z"
tags:
  - "provisioning"
  - "rpc-nodes"
  - "security"
  - "velia"
owners:
  - "user://e2964fe9-744c-4010-abbc-9a8d6033edc6"
source_revision_ids:
  - "srcrev_9f3d720b8fcd04a5aa8f2d1390511c1a"
conflict_state: "none"
---

# Provision RPC Nodes on Velia - Prerequisites

## Summary

Step-by-step guide for initial server setup before provisioning RPC nodes on Velia, including root password change, SSH key creation, regular user with sudo, and Python3 installation.

## Claims

- After machine is ready, change the root password using the passwd command. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383) `source_document_id=srcdoc_f62969b71f25eb1bbbe59751fbde07df` `source_revision_id=srcrev_9f3d720b8fcd04a5aa8f2d1390511c1a` `chunk_id=srcchunk_3fdd596369966ea628f1d0be5b6b63f3` `native_locator=https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383` `source_timestamp=2024-10-11T22:39:00Z`
- Create an SSH key pair using ssh-keygen -t rsa -b 4096 -m PEM -f ~/.ssh/velia_key.pem and copy the public key to the remote nodes using ssh-copy-id. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383) `source_document_id=srcdoc_f62969b71f25eb1bbbe59751fbde07df` `source_revision_id=srcrev_9f3d720b8fcd04a5aa8f2d1390511c1a` `chunk_id=srcchunk_3fdd596369966ea628f1d0be5b6b63f3` `native_locator=https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383` `source_timestamp=2024-10-11T22:39:00Z`
- Create a regular user 'velia-user' with sudo access, copy SSH authorized_keys, and configure passwordless sudo. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383) `source_document_id=srcdoc_f62969b71f25eb1bbbe59751fbde07df` `source_revision_id=srcrev_9f3d720b8fcd04a5aa8f2d1390511c1a` `chunk_id=srcchunk_3fdd596369966ea628f1d0be5b6b63f3` `native_locator=https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383` `source_timestamp=2024-10-11T22:39:00Z`
- Install sudo, vim, and python3 on the server. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383) `source_document_id=srcdoc_f62969b71f25eb1bbbe59751fbde07df` `source_revision_id=srcrev_9f3d720b8fcd04a5aa8f2d1390511c1a` `chunk_id=srcchunk_3fdd596369966ea628f1d0be5b6b63f3` `native_locator=https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383` `source_timestamp=2024-10-11T22:39:00Z`
- Save root username/password and SSH key to 1Password vault named Velia. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383) `source_document_id=srcdoc_f62969b71f25eb1bbbe59751fbde07df` `source_revision_id=srcrev_9f3d720b8fcd04a5aa8f2d1390511c1a` `chunk_id=srcchunk_3fdd596369966ea628f1d0be5b6b63f3` `native_locator=https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383` `source_timestamp=2024-10-11T22:39:00Z`

## Sources

- `source_document_id`: `srcdoc_f62969b71f25eb1bbbe59751fbde07df`
- `source_revision_id`: `srcrev_9f3d720b8fcd04a5aa8f2d1390511c1a`
- `source_url`: [Notion source](https://www.notion.so/Provision-RPC-Nodes-on-Velia-11a051299a5480ec964ee5fe03f30383)
