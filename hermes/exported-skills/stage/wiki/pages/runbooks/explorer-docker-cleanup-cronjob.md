---
title: "Explorer Docker Cleanup Cronjob"
type: "runbook"
slug: "runbooks/explorer-docker-cleanup-cronjob"
freshness: "2026-05-05T06:37:09Z"
tags: []
owners: []
source_revision_ids:
  - "srcrev_91b8b87168b92e95a3cec146d29c5d92"
conflict_state: "none"
---

# Explorer Docker Cleanup Cronjob

## Summary

A cronjob that runs daily at midnight to prune unused Docker resources and truncate container log files on explorer instances.

## Claims

- The docker-cleanup.sh script is located at /usr/local/bin/docker-cleanup.sh. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f) `source_document_id=srcdoc_16f99c0e022eb78dfadbd524398740b5` `source_revision_id=srcrev_91b8b87168b92e95a3cec146d29c5d92` `chunk_id=srcchunk_93b717929f03304ef24b98352cdb7fa6` `native_locator=https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f` `source_timestamp=2026-05-05T06:37:09Z`
- The cron job runs the docker-cleanup.sh script daily at midnight. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f) `source_document_id=srcdoc_16f99c0e022eb78dfadbd524398740b5` `source_revision_id=srcrev_91b8b87168b92e95a3cec146d29c5d92` `chunk_id=srcchunk_93b717929f03304ef24b98352cdb7fa6` `native_locator=https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f` `source_timestamp=2026-05-05T06:37:09Z`
- The script removes exited containers, unused images, unused volumes, unused networks, performs a comprehensive system prune, and truncates Docker container log files. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f) `source_document_id=srcdoc_16f99c0e022eb78dfadbd524398740b5` `source_revision_id=srcrev_91b8b87168b92e95a3cec146d29c5d92` `chunk_id=srcchunk_93b717929f03304ef24b98352cdb7fa6` `native_locator=https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f` `source_timestamp=2026-05-05T06:37:09Z`
- The devnet, partner testnet, and mininet explorers all use identical cronjob implementations. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f) `source_document_id=srcdoc_16f99c0e022eb78dfadbd524398740b5` `source_revision_id=srcrev_91b8b87168b92e95a3cec146d29c5d92` `chunk_id=srcchunk_93b717929f03304ef24b98352cdb7fa6` `native_locator=https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f` `source_timestamp=2026-05-05T06:37:09Z`

## Sources

- `source_document_id`: `srcdoc_16f99c0e022eb78dfadbd524398740b5`
- `source_revision_id`: `srcrev_91b8b87168b92e95a3cec146d29c5d92`
- `source_url`: [Notion source](https://www.notion.so/Explorer-Cronjob-Documentation-aa828b8e85c94e328a20872a293efc6f)
