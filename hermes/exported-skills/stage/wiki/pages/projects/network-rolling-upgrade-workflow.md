---
title: "Network Rolling Upgrade Workflow"
type: "project"
slug: "projects/network-rolling-upgrade-workflow"
freshness: "2026-05-05T06:41:00Z"
tags:
  - "automation"
  - "github-actions"
  - "network"
  - "rolling-upgrade"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_1d83a73ffe4db640154f231257b3e5be"
conflict_state: "none"
---

# Network Rolling Upgrade Workflow

## Summary

Design and implementation of a rolling upgrade workflow for the L1 network, enabling node-by-node upgrades while maintaining network health. Emphasizes security via OIDC, dedicated SSH keys per region, and engineer whitelisting; and usability via one-click trigger and real-time Slack notifications.

## Claims

- The rolling upgrade workflow uses short-lived OIDC tokens for AWS authentication and authorization instead of permanent AWS keys. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6) `source_document_id=srcdoc_7e4471078f8787f499000915be46527b` `source_revision_id=srcrev_1d83a73ffe4db640154f231257b3e5be` `chunk_id=srcchunk_4f0e14b4374176081e11d28bf66f7047` `native_locator=https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6` `source_timestamp=2026-05-05T06:41:00Z`
- Dedicated SSH keys are created for each AWS region, with region info encoded in the key name (e.g., ~/.ssh/public-testnet-us-east-1.pem). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6) `source_document_id=srcdoc_7e4471078f8787f499000915be46527b` `source_revision_id=srcrev_1d83a73ffe4db640154f231257b3e5be` `chunk_id=srcchunk_4f0e14b4374176081e11d28bf66f7047` `native_locator=https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6` `source_timestamp=2026-05-05T06:41:00Z`
- Only authorized engineers are allowed to trigger state-changing workflows, enforced via a whitelist. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6) `source_document_id=srcdoc_7e4471078f8787f499000915be46527b` `source_revision_id=srcrev_1d83a73ffe4db640154f231257b3e5be` `chunk_id=srcchunk_4f0e14b4374176081e11d28bf66f7047` `native_locator=https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6` `source_timestamp=2026-05-05T06:41:00Z`
- The workflow can be triggered with one button click using default values, while also allowing customization of geth and story versions or branches. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6) `source_document_id=srcdoc_7e4471078f8787f499000915be46527b` `source_revision_id=srcrev_1d83a73ffe4db640154f231257b3e5be` `chunk_id=srcchunk_4f0e14b4374176081e11d28bf66f7047` `native_locator=https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6` `source_timestamp=2026-05-05T06:41:00Z`
- Real-time version change information is pushed to a Slack channel during rolling upgrades, providing better observability than Grafana's pull-based delay. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6) `source_document_id=srcdoc_7e4471078f8787f499000915be46527b` `source_revision_id=srcrev_1d83a73ffe4db640154f231257b3e5be` `chunk_id=srcchunk_4f0e14b4374176081e11d28bf66f7047` `native_locator=https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6` `source_timestamp=2026-05-05T06:41:00Z`

## Open Questions

- SSH key rotation process needs further discussion with the mentioned engineer (user://e32a65e3-e3e4-431a-8afa-4b2acbc8f408).

## Related Pages

- `network-hard-reset-workflow`

## Sources

- `source_document_id`: `srcdoc_7e4471078f8787f499000915be46527b`
- `source_revision_id`: `srcrev_1d83a73ffe4db640154f231257b3e5be`
- `source_url`: [Notion source](https://www.notion.so/A-Workflow-to-do-a-Network-Rolling-Upgrade-623e9d104e0e4402af7b373457f824a6)
