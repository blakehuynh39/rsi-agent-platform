---
title: "Mainnet Deployment and OpSec"
type: "project"
slug: "projects/mainnet-deployment-and-opsec"
freshness: "2025-04-03T20:36:00Z"
tags:
  - "defender"
  - "deployment"
  - "gnosis"
  - "mainnet"
  - "multisig"
  - "opsec"
owners: []
source_revision_ids:
  - "srcrev_45b2efc2b132b409db598194725ad88c"
conflict_state: "none"
---

# Mainnet Deployment and OpSec

## Summary

Project tracking mainnet deployment tasks and operational security setup, including Gnosis Safe configuration, Defender integration, and security monitoring runbooks.

## Claims

- Gnosis Safe deployment and signer configuration is complete. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07) `source_document_id=srcdoc_fcc22130db96b56c1570dfba6a0509d4` `source_revision_id=srcrev_45b2efc2b132b409db598194725ad88c` `chunk_id=srcchunk_9755ee9069f1789e3f8025faafe555ea` `native_locator=https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07` `source_timestamp=2025-04-03T20:36:00Z`
- Multisig Approval process has been added to Defender. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07) `source_document_id=srcdoc_fcc22130db96b56c1570dfba6a0509d4` `source_revision_id=srcrev_45b2efc2b132b409db598194725ad88c` `chunk_id=srcchunk_9755ee9069f1789e3f8025faafe555ea` `native_locator=https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07` `source_timestamp=2025-04-03T20:36:00Z`
- Contracts have been deployed pointing to the multisig. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07) `source_document_id=srcdoc_fcc22130db96b56c1570dfba6a0509d4` `source_revision_id=srcrev_45b2efc2b132b409db598194725ad88c` `chunk_id=srcchunk_9755ee9069f1789e3f8025faafe555ea` `native_locator=https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07` `source_timestamp=2025-04-03T20:36:00Z`
- Contracts have been manually added to Defender. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07) `source_document_id=srcdoc_fcc22130db96b56c1570dfba6a0509d4` `source_revision_id=srcrev_45b2efc2b132b409db598194725ad88c` `chunk_id=srcchunk_9755ee9069f1789e3f8025faafe555ea` `native_locator=https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07` `source_timestamp=2025-04-03T20:36:00Z`
- Contract state check, specifically access control address, is complete. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07) `source_document_id=srcdoc_fcc22130db96b56c1570dfba6a0509d4` `source_revision_id=srcrev_45b2efc2b132b409db598194725ad88c` `chunk_id=srcchunk_9755ee9069f1789e3f8025faafe555ea` `native_locator=https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07` `source_timestamp=2025-04-03T20:36:00Z`
- Grant role script has been proposed. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07) `source_document_id=srcdoc_fcc22130db96b56c1570dfba6a0509d4` `source_revision_id=srcrev_45b2efc2b132b409db598194725ad88c` `chunk_id=srcchunk_9755ee9069f1789e3f8025faafe555ea` `native_locator=https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07` `source_timestamp=2025-04-03T20:36:00Z`
- Set SPUML script has been proposed. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07) `source_document_id=srcdoc_fcc22130db96b56c1570dfba6a0509d4` `source_revision_id=srcrev_45b2efc2b132b409db598194725ad88c` `chunk_id=srcchunk_9755ee9069f1789e3f8025faafe555ea` `native_locator=https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07` `source_timestamp=2025-04-03T20:36:00Z`
- Adding team and signers to Defender requires a paid plan and is not yet done. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07) `source_document_id=srcdoc_fcc22130db96b56c1570dfba6a0509d4` `source_revision_id=srcrev_45b2efc2b132b409db598194725ad88c` `chunk_id=srcchunk_9755ee9069f1789e3f8025faafe555ea` `native_locator=https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07` `source_timestamp=2025-04-03T20:36:00Z`

## Open Questions

- Should OZ AI code auditor be granted access?
- What dangerous states should be monitored?
- What emergency situations would require the emergency dev multisig to act, and what actions would it take?

## Sources

- `source_document_id`: `srcdoc_fcc22130db96b56c1570dfba6a0509d4`
- `source_revision_id`: `srcrev_45b2efc2b132b409db598194725ad88c`
- `source_url`: [Notion source](https://www.notion.so/Mainnet-deployment-and-opsec-6e09fdddcf474564ac3b51148004fd07)
