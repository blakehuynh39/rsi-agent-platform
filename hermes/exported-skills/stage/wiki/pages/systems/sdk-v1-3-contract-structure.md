---
title: "SDK V1.3 Contract Structure"
type: "system"
slug: "systems/sdk-v1-3-contract-structure"
freshness: "2026-05-05T06:39:01Z"
tags:
  - "architecture"
  - "contracts"
  - "sdk"
  - "solidity"
owners: []
source_revision_ids:
  - "srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf"
conflict_state: "none"
---

# SDK V1.3 Contract Structure

## Summary

Directory layout and contract organization for the SDK V1.3 smart contracts, including core contracts, workflow modules, libraries, and interfaces.

## Claims

- The SDK V1.3 contracts directory contains SPGNFT.sol and BaseWorkflow.sol at the root level. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296) `source_document_id=srcdoc_a3e7702755f60f2924a99a3f32ba7649` `source_revision_id=srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf` `chunk_id=srcchunk_9041686b9bcbf40c95d668d1e089f445` `native_locator=https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296` `source_timestamp=2026-05-05T06:39:01Z`
- The workflows subdirectory contains GroupingWorkflows.sol, LicenseAttachmentWorkflows.sol, RegistrationWorkflows.sol, and DerivativeWorkflows.sol. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296) `source_document_id=srcdoc_a3e7702755f60f2924a99a3f32ba7649` `source_revision_id=srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf` `chunk_id=srcchunk_9041686b9bcbf40c95d668d1e089f445` `native_locator=https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296` `source_timestamp=2026-05-05T06:39:01Z`
- The lib directory includes PermissionHelper.sol, Errors.sol, LicensingHelper.sol, MetadataHelper.sol, WorkflowStructs.sol, and SPGNFTLib.sol. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296) `source_document_id=srcdoc_a3e7702755f60f2924a99a3f32ba7649` `source_revision_id=srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf` `chunk_id=srcchunk_9041686b9bcbf40c95d668d1e089f445` `native_locator=https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296` `source_timestamp=2026-05-05T06:39:01Z`
- The interfaces directory contains ISPGNFT.sol at the root and workflow interfaces ILicenseAttachmentWorkflows.sol, IRegistrationWorkflows.sol, IDerivativeWorkflows.sol, and IGroupingWorkflows.sol in a workflows subdirectory. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296) `source_document_id=srcdoc_a3e7702755f60f2924a99a3f32ba7649` `source_revision_id=srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf` `chunk_id=srcchunk_9041686b9bcbf40c95d668d1e089f445` `native_locator=https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296` `source_timestamp=2026-05-05T06:39:01Z`

## Sources

- `source_document_id`: `srcdoc_a3e7702755f60f2924a99a3f32ba7649`
- `source_revision_id`: `srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf`
- `source_url`: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296)
