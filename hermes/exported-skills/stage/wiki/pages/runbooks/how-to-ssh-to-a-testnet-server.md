---
title: "HOW-TO SSH to a Testnet Server"
type: "runbook"
slug: "runbooks/how-to-ssh-to-a-testnet-server"
freshness: "2024-07-31T05:45:00Z"
tags:
  - "how-to"
  - "ssh"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_f7af87f66462238e43be19e45c3d768c"
conflict_state: "none"
---

# HOW-TO SSH to a Testnet Server

## Summary

A guide providing SSH commands and key information for accessing partner-testnet servers.

## Claims

- The guide provides SSH commands for partner-testnet bootnodes, explorers, and validators. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-SSH-to-a-Testnet-Server-3bba6ed05b804086a87280fa16d9b878#chunk-1) `source_document_id=srcdoc_76a8108de740e94eb722bd291820a762` `source_revision_id=srcrev_f7af87f66462238e43be19e45c3d768c` `chunk_id=srcchunk_15f37b1fa76baa2a7f57c7f41e278d59` `native_locator=https://www.notion.so/HOW-TO-SSH-to-a-Testnet-Server-3bba6ed05b804086a87280fa16d9b878#chunk-1` `source_timestamp=2024-07-31T05:45:00Z`
- SSH keys for the testnet servers are shared via 1Password. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-SSH-to-a-Testnet-Server-3bba6ed05b804086a87280fa16d9b878#chunk-2) `source_document_id=srcdoc_76a8108de740e94eb722bd291820a762` `source_revision_id=srcrev_f7af87f66462238e43be19e45c3d768c` `chunk_id=srcchunk_29a8b2fe7677004d6c7bcb4917c8d159` `native_locator=https://www.notion.so/HOW-TO-SSH-to-a-Testnet-Server-3bba6ed05b804086a87280fa16d9b878#chunk-2` `source_timestamp=2024-07-31T05:45:00Z`
- An example SSH command for partner-testnet-validator1 is: ssh -i "~/.ssh/partner-testnet-key-us-east-1.pem" ec2-user@ec2-3-89-61-145.compute-1.amazonaws.com `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-SSH-to-a-Testnet-Server-3bba6ed05b804086a87280fa16d9b878#chunk-2) `source_document_id=srcdoc_76a8108de740e94eb722bd291820a762` `source_revision_id=srcrev_f7af87f66462238e43be19e45c3d768c` `chunk_id=srcchunk_29a8b2fe7677004d6c7bcb4917c8d159` `native_locator=https://www.notion.so/HOW-TO-SSH-to-a-Testnet-Server-3bba6ed05b804086a87280fa16d9b878#chunk-2` `source_timestamp=2024-07-31T05:45:00Z`

## Related Pages

- `concepts/how-to-series`

## Sources

- `source_document_id`: `srcdoc_76a8108de740e94eb722bd291820a762`
- `source_revision_id`: `srcrev_f7af87f66462238e43be19e45c3d768c`
- `source_url`: [Notion source](https://www.notion.so/HOW-TO-SSH-to-a-Testnet-Server-3bba6ed05b804086a87280fa16d9b878)
