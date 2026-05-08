---
title: "AWS CLI Setup"
type: "runbook"
slug: "runbooks/aws-cli-setup"
freshness: "2024-01-25T19:53:00Z"
tags:
  - "aws"
  - "cli"
  - "sso"
owners: []
source_revision_ids:
  - "srcrev_4bd2a17d5dbfbb5f119bdea5de9d1da5"
conflict_state: "none"
---

# AWS CLI Setup

## Summary

Instructions for installing and configuring AWS CLI v2 with SSO authentication, including profile setup and login.

## Claims

- Install awscli2 by following the official AWS guide. `claim:claim_aws_cli_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa) `source_document_id=srcdoc_6244d4fdf2e31c7e90c94c83cf6fc243` `source_revision_id=srcrev_4bd2a17d5dbfbb5f119bdea5de9d1da5` `chunk_id=srcchunk_f98c65d3ef38bcee04f375ab78e5ffca` `native_locator=https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa` `source_timestamp=2024-01-25T19:53:00Z`
- Configure SSO by running `aws configure sso` and providing session name, start URL, and region. `claim:claim_aws_cli_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa) `source_document_id=srcdoc_6244d4fdf2e31c7e90c94c83cf6fc243` `source_revision_id=srcrev_4bd2a17d5dbfbb5f119bdea5de9d1da5` `chunk_id=srcchunk_f98c65d3ef38bcee04f375ab78e5ffca` `native_locator=https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa` `source_timestamp=2024-01-25T19:53:00Z`
- Log in to the SSO profile using `aws sso login --profile="${name}"`. `claim:claim_aws_cli_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa) `source_document_id=srcdoc_6244d4fdf2e31c7e90c94c83cf6fc243` `source_revision_id=srcrev_4bd2a17d5dbfbb5f119bdea5de9d1da5` `chunk_id=srcchunk_f98c65d3ef38bcee04f375ab78e5ffca` `native_locator=https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa` `source_timestamp=2024-01-25T19:53:00Z`
- Set the AWS_PROFILE environment variable to avoid specifying --profile in every command. `claim:claim_aws_cli_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa) `source_document_id=srcdoc_6244d4fdf2e31c7e90c94c83cf6fc243` `source_revision_id=srcrev_4bd2a17d5dbfbb5f119bdea5de9d1da5` `chunk_id=srcchunk_f98c65d3ef38bcee04f375ab78e5ffca` `native_locator=https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa` `source_timestamp=2024-01-25T19:53:00Z`
- The `~/.aws/config` file should contain the SSO profile with sso_start_url, sso_region, sso_account_id, sso_role_name, and region. `claim:claim_aws_cli_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa) `source_document_id=srcdoc_6244d4fdf2e31c7e90c94c83cf6fc243` `source_revision_id=srcrev_4bd2a17d5dbfbb5f119bdea5de9d1da5` `chunk_id=srcchunk_f98c65d3ef38bcee04f375ab78e5ffca` `native_locator=https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa` `source_timestamp=2024-01-25T19:53:00Z`

## Related Pages

- `concepts/devex`
- `runbooks/rds-backup-and-recovery-runbook`

## Sources

- `source_document_id`: `srcdoc_6244d4fdf2e31c7e90c94c83cf6fc243`
- `source_revision_id`: `srcrev_4bd2a17d5dbfbb5f119bdea5de9d1da5`
- `source_url`: [Notion source](https://www.notion.so/AWS-CLI-04c9bd60cb144018987991f681148efa)
