---
title: "Runbook: Node Storage Low"
type: "runbook"
slug: "runbooks/runbook-node-storage-low"
freshness: "2026-05-05T06:39:53Z"
tags:
  - "aws"
  - "ebs"
  - "runbook"
  - "storage"
  - "validator"
owners: []
source_revision_ids:
  - "srcrev_e1f26abb347b3356793bd442eddf2e71"
conflict_state: "none"
---

# Runbook: Node Storage Low

## Summary

Procedure to increase disk size on validators when a low storage alert (≥ 80%) is received.

## Claims

- If you received an alert (≥ 80%) about not enough disk space available on a validator, chances are we need to increase disk size to all validators. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4) `source_document_id=srcdoc_129d1870f801b7984496e9f54aa7680d` `source_revision_id=srcrev_e1f26abb347b3356793bd442eddf2e71` `chunk_id=srcchunk_4c28579df1e61ef3b2f250fed22d70c4` `native_locator=https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4` `source_timestamp=2026-05-05T06:39:53Z`
- To increase disk size, modify the EBS volume on AWS management console without stopping the EC2 instance. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4) `source_document_id=srcdoc_129d1870f801b7984496e9f54aa7680d` `source_revision_id=srcrev_e1f26abb347b3356793bd442eddf2e71` `chunk_id=srcchunk_4c28579df1e61ef3b2f250fed22d70c4` `native_locator=https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4` `source_timestamp=2026-05-05T06:39:53Z`
- After modifying the EBS volume, SSH to the EC2 instance and execute: sudo growpart /dev/nvme0n1 1, sudo xfs_growfs /, and df -h. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4) `source_document_id=srcdoc_129d1870f801b7984496e9f54aa7680d` `source_revision_id=srcrev_e1f26abb347b3356793bd442eddf2e71` `chunk_id=srcchunk_4c28579df1e61ef3b2f250fed22d70c4` `native_locator=https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4` `source_timestamp=2026-05-05T06:39:53Z`
- Prerequisites for the procedure include valid AWS access, knowing SSH information to the node, and validating before and after changes on Grafana. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4) `source_document_id=srcdoc_129d1870f801b7984496e9f54aa7680d` `source_revision_id=srcrev_e1f26abb347b3356793bd442eddf2e71` `chunk_id=srcchunk_4c28579df1e61ef3b2f250fed22d70c4` `native_locator=https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4` `source_timestamp=2026-05-05T06:39:53Z`

## Sources

- `source_document_id`: `srcdoc_129d1870f801b7984496e9f54aa7680d`
- `source_revision_id`: `srcrev_e1f26abb347b3356793bd442eddf2e71`
- `source_url`: [Notion source](https://www.notion.so/PD-Node-storage-low-115051299a5480fdaf3adeeffc86f4b4)
