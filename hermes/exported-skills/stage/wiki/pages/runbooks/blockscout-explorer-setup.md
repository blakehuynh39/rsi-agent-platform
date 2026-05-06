---
title: "Blockscout Explorer Setup Documentation"
type: "runbook"
slug: "runbooks/blockscout-explorer-setup"
freshness: "2026-05-05T06:41:55Z"
tags:
  - "aws"
  - "blockscout"
  - "docker"
  - "explorer"
  - "nginx"
  - "ssl"
owners: []
source_revision_ids:
  - "srcrev_5a04b3204c002fb35fdaa7514a2b0147"
conflict_state: "none"
---

# Blockscout Explorer Setup Documentation

## Summary

Runbook for setting up Blockscout explorer on AWS EC2 using Docker Compose, including SSH access, Docker installation, configuration of environment files, and Nginx SSL termination.

## Claims

- SSH access to the EC2 instance requires the private SSH key, which may need conversion to .pem format using PuTTYgen. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-1) `source_document_id=srcdoc_566b0bb1347af53f43c4dd3ca713afc7` `source_revision_id=srcrev_5a04b3204c002fb35fdaa7514a2b0147` `chunk_id=srcchunk_78d8f473bac518b47c42f3a82cdd00af` `native_locator=https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-1` `source_timestamp=2026-05-05T06:41:55Z`
- For Linux, connect to the EC2 instance using the command: ssh -i "devnet-aws-stg.pem" ec2-user@ec2-54-183-162-216.us-west-1.compute.amazonaws.com `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-1) `source_document_id=srcdoc_566b0bb1347af53f43c4dd3ca713afc7` `source_revision_id=srcrev_5a04b3204c002fb35fdaa7514a2b0147` `chunk_id=srcchunk_78d8f473bac518b47c42f3a82cdd00af` `native_locator=https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-1` `source_timestamp=2026-05-05T06:41:55Z`
- Docker and docker-compose are installed using dnf, with docker-compose version 1.29.2 downloaded from GitHub. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-1) `source_document_id=srcdoc_566b0bb1347af53f43c4dd3ca713afc7` `source_revision_id=srcrev_5a04b3204c002fb35fdaa7514a2b0147` `chunk_id=srcchunk_78d8f473bac518b47c42f3a82cdd00af` `native_locator=https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-1` `source_timestamp=2026-05-05T06:41:55Z`
- The user-ops-indexer RPC URL is set to https://rpc.partner.testnet.storyprotocol.net/ in services/user-ops-indexer.yml. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-2) `source_document_id=srcdoc_566b0bb1347af53f43c4dd3ca713afc7` `source_revision_id=srcrev_5a04b3204c002fb35fdaa7514a2b0147` `chunk_id=srcchunk_a1e7447bc710d87fd80c888cbd2e7638` `native_locator=https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-2` `source_timestamp=2026-05-05T06:41:55Z`
- Nginx is configured to redirect all HTTP traffic to HTTPS and serve SSL certificates from /etc/nginx/certs/cert.pem and /etc/nginx/certs/key.pem. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-2) `source_document_id=srcdoc_566b0bb1347af53f43c4dd3ca713afc7` `source_revision_id=srcrev_5a04b3204c002fb35fdaa7514a2b0147` `chunk_id=srcchunk_a1e7447bc710d87fd80c888cbd2e7638` `native_locator=https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-2` `source_timestamp=2026-05-05T06:41:55Z`
- The proxy service in docker-compose exposes ports 80, 443, 8080, and 8081. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-2) `source_document_id=srcdoc_566b0bb1347af53f43c4dd3ca713afc7` `source_revision_id=srcrev_5a04b3204c002fb35fdaa7514a2b0147` `chunk_id=srcchunk_a1e7447bc710d87fd80c888cbd2e7638` `native_locator=https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-2` `source_timestamp=2026-05-05T06:41:55Z`
- Docker containers are built and started using 'docker-compose up -d' in the docker-compose directory. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-2) `source_document_id=srcdoc_566b0bb1347af53f43c4dd3ca713afc7` `source_revision_id=srcrev_5a04b3204c002fb35fdaa7514a2b0147` `chunk_id=srcchunk_a1e7447bc710d87fd80c888cbd2e7638` `native_locator=https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45#chunk-2` `source_timestamp=2026-05-05T06:41:55Z`

## Sources

- `source_document_id`: `srcdoc_566b0bb1347af53f43c4dd3ca713afc7`
- `source_revision_id`: `srcrev_5a04b3204c002fb35fdaa7514a2b0147`
- `source_url`: [Notion source](https://www.notion.so/Blockscout-Explorer-Setup-Documentation-f8671192cf18485d8184a3e7a0c16f45)
