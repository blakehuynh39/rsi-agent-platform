---
title: "Claim Backend Runbook"
type: "runbook"
slug: "runbooks/claim-backend-runbook"
freshness: "2025-02-12T21:21:00Z"
tags:
  - "airdrop"
  - "ansible"
  - "backend"
  - "claim"
  - "deployment"
  - "gcp"
owners: []
source_revision_ids:
  - "srcrev_39f459910b1e87f1169ab782ed4f9055"
conflict_state: "none"
---

# Claim Backend Runbook

## Summary

Runbook for extending and deploying the airdrop claim backend system, including GCP instance creation, configuration, and Ansible deployment steps.

## Claims

- This document describes how to extend the backend of the airdrop claim system. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-1) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_49df4f044d7e4216dec55dc50ecf0618` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-1` `source_timestamp=2025-02-12T21:21:00Z`
- New backend instances must be created in the `us-east1` region of the `mainnet` GCP project. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-1) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_49df4f044d7e4216dec55dc50ecf0618` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-1` `source_timestamp=2025-02-12T21:21:00Z`
- The instance type should be `n2-highcpu-8`, matching existing backends. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-1) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_49df4f044d7e4216dec55dc50ecf0618` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-1` `source_timestamp=2025-02-12T21:21:00Z`
- The default service account must be changed to one with full permission to all Cloud APIs; otherwise the instance cannot access secret keys in GCP KMS. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-2) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_44781e25c1a54d9808c965e5ec49e51a` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-2` `source_timestamp=2025-02-12T21:21:00Z`
- After creation, the instance must be added to the instance group so it receives traffic from the load balancer. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-3) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_eb6b0bf4ad4f8a71be37e1f5a39db62b` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-3` `source_timestamp=2025-02-12T21:21:00Z`
- The backend binary is built with `GOARCH=amd64 go build -ldflags "-s -w" -o claim-backend`. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_81cac867b64df60b47f500c7595d696f` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4` `source_timestamp=2025-02-12T21:21:00Z`
- The bastion host is accessible via SSH at `ubuntu@44.235.52.223` on port `22331`. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_81cac867b64df60b47f500c7595d696f` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4` `source_timestamp=2025-02-12T21:21:00Z`
- Binary and data files must be placed in `/home/ubuntu/ansible/playbook/files/claim-backend/mainnet/` on the bastion host. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_81cac867b64df60b47f500c7595d696f` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4` `source_timestamp=2025-02-12T21:21:00Z`
- Deployment is executed via `ansible-playbook -i mainnet-claim-be.ini playbook/deploy-claim-backend.yml --extra-vars "network=mainnet"` from `/home/ubuntu/ansible/playbook`. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_81cac867b64df60b47f500c7595d696f` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4` `source_timestamp=2025-02-12T21:21:00Z`
- When a new backend instance is added, the `mainnet-claim-be.ini` inventory file must be updated. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_81cac867b64df60b47f500c7595d696f` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4` `source_timestamp=2025-02-12T21:21:00Z`
- The configuration file is a Jinja2 template located at `/home/ubuntu/ansible/playbook/files/claim-backend/mainnet/config.yaml.j2`. Update the config items in this file and re-run the Ansible script to apply changes. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4) `source_document_id=srcdoc_77b27c2fe91c20256d4834c693e96dee` `source_revision_id=srcrev_39f459910b1e87f1169ab782ed4f9055` `chunk_id=srcchunk_81cac867b64df60b47f500c7595d696f` `native_locator=https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e#chunk-4` `source_timestamp=2025-02-12T21:21:00Z`

## Sources

- `source_document_id`: `srcdoc_77b27c2fe91c20256d4834c693e96dee`
- `source_revision_id`: `srcrev_39f459910b1e87f1169ab782ed4f9055`
- `source_url`: [Notion source](https://www.notion.so/Claim-Backend-Runbook-194051299a5480caaa64da4e09caec6e)
