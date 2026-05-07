---
title: "Docker ECR Login"
type: "runbook"
slug: "runbooks/docker-ecr-login"
freshness: "2023-12-03T05:53:00Z"
tags:
  - "aws"
  - "cli"
  - "docker"
  - "ecr"
owners: []
source_revision_ids:
  - "srcrev_e8f70ac48e54b16522d123711079dd2e"
conflict_state: "none"
---

# Docker ECR Login

## Summary

Procedure to log into AWS ECR using Docker CLI, as ECR is not publicly accessible.

## Claims

- To log into AWS ECR with Docker CLI, export AWS_PROFILE, AWS_ACCOUNT_ID, AWS_REGION, then run `aws ecr get-login-password --region=$AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com`. The ECR is not publicly accessible, so this login is required. Ensure the region matches the ECR; a token from one region won't work in another. Set AWS_DEFAULT_REGION and AWS_REGION in your AWS config profile to avoid specifying --region each time. `claim:claim_docker_ecr_login` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Docker-CLI-a16c87904b7046b7a909008be970397c) `source_document_id=srcdoc_a089be65568ece5475c1c25c2fb37928` `source_revision_id=srcrev_e8f70ac48e54b16522d123711079dd2e` `chunk_id=srcchunk_2bf08f7aa1a8036bcf3182c7eff21f13` `native_locator=https://www.notion.so/Docker-CLI-a16c87904b7046b7a909008be970397c` `source_timestamp=2023-12-03T05:53:00Z`

## Related Pages

- `concepts/devex`

## Sources

- `source_document_id`: `srcdoc_a089be65568ece5475c1c25c2fb37928`
- `source_revision_id`: `srcrev_e8f70ac48e54b16522d123711079dd2e`
- `source_url`: [Notion source](https://www.notion.so/Docker-CLI-a16c87904b7046b7a909008be970397c)
