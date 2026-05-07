---
title: "RDS Backup and Recovery Runbook"
type: "runbook"
slug: "runbooks/rds-backup-and-recovery-runbook"
freshness: "2023-03-06T23:05:00Z"
tags:
  - "aws"
  - "backup"
  - "rds"
  - "recovery"
owners:
  - "Andy Wu"
source_revision_ids:
  - "srcrev_70f7c7ce976ea9c3b56ee1320cdec92b"
conflict_state: "none"
---

# RDS Backup and Recovery Runbook

## Summary

Steps to configure automated daily RDS backups with 7-day retention, using AWS Management Console or CLI.

## Claims

- The example uses the RDS instance in the 'stag' environment in region 'us-east-2'. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f) `source_document_id=srcdoc_5376542b6be2e93b80fcf748a5bf105a` `source_revision_id=srcrev_70f7c7ce976ea9c3b56ee1320cdec92b` `chunk_id=srcchunk_8c4b1ea729f0bfc355ec60a6a6125377` `native_locator=https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f` `source_timestamp=2023-03-06T23:05:00Z`
- A daily RDS snapshot is necessary in case of any bad events. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f) `source_document_id=srcdoc_5376542b6be2e93b80fcf748a5bf105a` `source_revision_id=srcrev_70f7c7ce976ea9c3b56ee1320cdec92b` `chunk_id=srcchunk_8c4b1ea729f0bfc355ec60a6a6125377` `native_locator=https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f` `source_timestamp=2023-03-06T23:05:00Z`
- The backup retention period is set to 7 days. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f) `source_document_id=srcdoc_5376542b6be2e93b80fcf748a5bf105a` `source_revision_id=srcrev_70f7c7ce976ea9c3b56ee1320cdec92b` `chunk_id=srcchunk_8c4b1ea729f0bfc355ec60a6a6125377` `native_locator=https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f` `source_timestamp=2023-03-06T23:05:00Z`
- The configuration can be done via the AWS Management Console or AWS CLI. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f) `source_document_id=srcdoc_5376542b6be2e93b80fcf748a5bf105a` `source_revision_id=srcrev_70f7c7ce976ea9c3b56ee1320cdec92b` `chunk_id=srcchunk_8c4b1ea729f0bfc355ec60a6a6125377` `native_locator=https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f` `source_timestamp=2023-03-06T23:05:00Z`

## Sources

- `source_document_id`: `srcdoc_5376542b6be2e93b80fcf748a5bf105a`
- `source_revision_id`: `srcrev_70f7c7ce976ea9c3b56ee1320cdec92b`
- `source_url`: [Notion source](https://www.notion.so/KB-RDS-Backup-and-Recovery-0530ff1a905447f68da85acf4de04d5f)
