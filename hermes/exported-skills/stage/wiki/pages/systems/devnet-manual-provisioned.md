---
title: "Devnet (Manual Provisioned)"
type: "system"
slug: "systems/devnet-manual-provisioned"
freshness: "2024-06-01T21:12:00Z"
tags:
  - "aws"
  - "devnet"
  - "infrastructure"
owners:
  - "user://e2964fe9-744c-4010-abbc-9a8d6033edc6"
source_revision_ids:
  - "srcrev_e37c42fa89e8071d8023156682bbc0f3"
conflict_state: "none"
---

# Devnet (Manual Provisioned)

## Summary

Details of the manually provisioned devnet with 15 validators, 1 bootnode, and an explorer, hosted in AWS us-west-1.

## Claims

- The initial devnet consists of 15 validators and 1 bootnode. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c) `source_document_id=srcdoc_739c49d0208ee3bc84ed0788bc91f5a6` `source_revision_id=srcrev_e37c42fa89e8071d8023156682bbc0f3` `chunk_id=srcchunk_a35e80c454db006f2fb4410cf0be7eff` `native_locator=https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c` `source_timestamp=2024-06-01T21:12:00Z`
- All nodes are located in the us-west-1 region of AWS account 478656756051. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c) `source_document_id=srcdoc_739c49d0208ee3bc84ed0788bc91f5a6` `source_revision_id=srcrev_e37c42fa89e8071d8023156682bbc0f3` `chunk_id=srcchunk_a35e80c454db006f2fb4410cf0be7eff` `native_locator=https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c` `source_timestamp=2024-06-01T21:12:00Z`
- The SSH key for accessing the servers is devnet-aws-stg.pem. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c) `source_document_id=srcdoc_739c49d0208ee3bc84ed0788bc91f5a6` `source_revision_id=srcrev_e37c42fa89e8071d8023156682bbc0f3` `chunk_id=srcchunk_a35e80c454db006f2fb4410cf0be7eff` `native_locator=https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c` `source_timestamp=2024-06-01T21:12:00Z`
- All validators, bootnodes, and explorers use the same EC2 instance type: t3a.xlarge with 16 GB RAM, 4 vCPUs, priced at $0.1504 per hour. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c) `source_document_id=srcdoc_739c49d0208ee3bc84ed0788bc91f5a6` `source_revision_id=srcrev_e37c42fa89e8071d8023156682bbc0f3` `chunk_id=srcchunk_a35e80c454db006f2fb4410cf0be7eff` `native_locator=https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c` `source_timestamp=2024-06-01T21:12:00Z`
- The bootnode is accessible at ec2-54-151-119-54.us-west-1.compute.amazonaws.com. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c) `source_document_id=srcdoc_739c49d0208ee3bc84ed0788bc91f5a6` `source_revision_id=srcrev_e37c42fa89e8071d8023156682bbc0f3` `chunk_id=srcchunk_a35e80c454db006f2fb4410cf0be7eff` `native_locator=https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c` `source_timestamp=2024-06-01T21:12:00Z`
- The 15 validators are accessible via SSH at the following addresses: ec2-54-67-95-71.us-west-1.compute.amazonaws.com (validator 1), ec2-204-236-149-190.us-west-1.compute.amazonaws.com (validator 2), ec2-13-57-205-57.us-west-1.compute.amazonaws.com (validator 3), ec2-52-53-213-187.us-west-1.compute.amazonaws.com (validator 4), ec2-54-153-111-104.us-west-1.compute.amazonaws.com (validator 5), ec2-18-144-166-110.us-west-1.compute.amazonaws.com (validator 6), ec2-18-144-89-70.us-west-1.compute.amazonaws.com (validator 7), ec2-54-151-6-221.us-west-1.compute.amazonaws.com (validator 8), ec2-18-144-23-168.us-west-1.compute.amazonaws.com (validator 9), ec2-54-183-120-252.us-west-1.compute.amazonaws.com (validator 10), ec2-54-215-99-101.us-west-1.compute.amazonaws.com (validator 11), ec2-13-52-214-176.us-west-1.compute.amazonaws.com (validator 12), ec2-54-215-246-207.us-west-1.compute.amazonaws.com (validator 13), ec2-13-57-214-146.us-west-1.compute.amazonaws.com (validator 14), ec2-13-52-239-195.us-west-1.compute.amazonaws.com (validator 15). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c) `source_document_id=srcdoc_739c49d0208ee3bc84ed0788bc91f5a6` `source_revision_id=srcrev_e37c42fa89e8071d8023156682bbc0f3` `chunk_id=srcchunk_a35e80c454db006f2fb4410cf0be7eff` `native_locator=https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c` `source_timestamp=2024-06-01T21:12:00Z`
- The explorer is accessible at http://54.183.162.216/ and has a domain name https://explorer.devnet.storyprotocol.net/ (SSL pending due to Cloudflare rate limit). `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c) `source_document_id=srcdoc_739c49d0208ee3bc84ed0788bc91f5a6` `source_revision_id=srcrev_e37c42fa89e8071d8023156682bbc0f3` `chunk_id=srcchunk_a35e80c454db006f2fb4410cf0be7eff` `native_locator=https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c` `source_timestamp=2024-06-01T21:12:00Z`
- SSH access to the servers requires contacting the user mentioned in the document. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c) `source_document_id=srcdoc_739c49d0208ee3bc84ed0788bc91f5a6` `source_revision_id=srcrev_e37c42fa89e8071d8023156682bbc0f3` `chunk_id=srcchunk_a35e80c454db006f2fb4410cf0be7eff` `native_locator=https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c` `source_timestamp=2024-06-01T21:12:00Z`

## Sources

- `source_document_id`: `srcdoc_739c49d0208ee3bc84ed0788bc91f5a6`
- `source_revision_id`: `srcrev_e37c42fa89e8071d8023156682bbc0f3`
- `source_url`: [Notion source](https://www.notion.so/Devnet-Manual-Provisioned-8ea064d70c884ce19f988e5ae641a46c)
