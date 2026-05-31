---
title: "Periphery Code Freeze Checklist"
type: "runbook"
slug: "runbooks/periphery-code-freeze-checklist"
freshness: "2024-12-24T04:34:00Z"
tags:
  - "backward-compatibility"
  - "code-freeze"
  - "periphery"
  - "v1.2"
owners: []
source_revision_ids:
  - "srcrev_161a1a299545c73d11ec176bd4938ea0"
conflict_state: "none"
---

# Periphery Code Freeze Checklist

## Summary

Checklist for V1.2 periphery contract changes to ensure backward compatibility. Most workflow functions are 100% BC, but a few require updates due to contract size limits or signature changes.

## Claims

- DerivativeWorkflows: No interface change, but need to update the struct from using MakeDerivative to MakeDerivativeDEPR. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- mintAndRegisterIpAndMakeDerivative: 100% BC (nft metadata uniqueness check). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- mintAndRegisterIpAndMakeDerivativeWithLicenseTokens: 100% BC (nft metadata uniqueness check, new maxRts param). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerIpAndMakeDerivativeWithLicenseTokens: 100% BC (new maxRts param, potentially change to use batch permission signature). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerIpAndMakeDerivative: 100% BC (potentially change to use batch permission signature). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- mintAndRegisterIpAndAttachLicenseAndAddToGroup: 100% BC (license terms and config, nft metadata uniqueness check, potentially change to use batch permission). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerIpAndAttachLicenseAndAddToGroup: No interface change, but the signature needs to include permission for set licensing config. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerGroupAndAttachLicense: 100% BC (license terms and config). `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerGroupAndAttachLicenseAndAddIps: 100% BC (license terms and config). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- collectRoyaltiesAndClaimReward: 100% BC (remove groupSnapshotIds param). `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerPILTermsAndAttach: 100% BC (license terms and config, batch permission sig). `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- mintAndRegisterIpAndAttachPILTerms: 100% BC (license terms and config, unique nft metadata uniqueness check). `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerIpAndAttachPILTerms: 100% BC (license terms and config, batch permission sig). `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- mintAndRegisterIp: 100% BC (nft metadata uniqueness check, potentially change to use batch permission signature). `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- mintAndRegisterIpAndAttachPILTermsAndDistributeRoyaltyTokens: 100% BC (nft metadata uniqueness check, license terms and config). `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- mintAndRegisterIpAndMakeDerivativeAndDistributeRoyaltyTokens: Required Change (contract size limit reached). `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerIpAndAttachPILTermsAndDeployRoyaltyVault: 100% BC (license terms and config, batch permission signature). `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`
- registerIpAndMakeDerivativeAndDeployRoyaltyVault: Required Change (contract size limit reached) [truncated in source]. `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e) `source_document_id=srcdoc_8e94206346d5c5d7df51c53c590e130a` `source_revision_id=srcrev_161a1a299545c73d11ec176bd4938ea0` `chunk_id=srcchunk_d6c461317ad7a5c0d54786f1e6152aee` `native_locator=https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e` `source_timestamp=2024-12-24T04:34:00Z`

## Open Questions

- The discussion link for `mintAndRegisterIpAndMakeDerivative` (discussion://16005129-9a54-805c-8330-001cd1e94a75) is not available as a source chunk; its content is unknown.
- The source chunk is truncated; the full details for `registerIpAndMakeDerivativeAndDeployRoyaltyVault` are missing. What is the complete required change?

## Sources

- `source_document_id`: `srcdoc_8e94206346d5c5d7df51c53c590e130a`
- `source_revision_id`: `srcrev_161a1a299545c73d11ec176bd4938ea0`
- `source_url`: [Notion source](https://www.notion.so/Periphery-Code-Freeze-Checklist-15f051299a54803391a7fd9a934ec67e)
