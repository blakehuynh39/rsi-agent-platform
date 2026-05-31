---
title: "Frontend / Periphery Contracts"
type: "concept"
slug: "concepts/frontend-periphery-contracts"
freshness: "2024-12-03T05:20:00Z"
tags:
  - "gateway"
  - "module-registry"
  - "smart-contracts"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_cd15059203b45a6a2cd6b0b5a301810d"
conflict_state: "none"
---

# Frontend / Periphery Contracts

## Summary

Defines the interface and responsibilities for frontend (gateway) contracts in Story Protocol, including the IGateway and IModuleRegistry interfaces and their role in managing module dependencies.

## Claims

- The IGateway interface requires frontend contracts to declare module dependencies via a ModuleDependencies struct containing keys and function signatures. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6) `source_document_id=srcdoc_9dd6dfee136ec58d0fa5990aa6e763b4` `source_revision_id=srcrev_cd15059203b45a6a2cd6b0b5a301810d` `chunk_id=srcchunk_916a7b54a95663a25edb9f37d46287ec` `native_locator=https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6` `source_timestamp=2024-12-03T05:20:00Z`
- The IGateway interface includes an updateDependencies function that synchronizes downstream dependencies via the module registry and may only be called by the module registry. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6) `source_document_id=srcdoc_9dd6dfee136ec58d0fa5990aa6e763b4` `source_revision_id=srcrev_cd15059203b45a6a2cd6b0b5a301810d` `chunk_id=srcchunk_916a7b54a95663a25edb9f37d46287ec` `native_locator=https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6` `source_timestamp=2024-12-03T05:20:00Z`
- The IGateway interface includes a getDependencies view function to fetch all module dependencies required by the gateway contract. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6) `source_document_id=srcdoc_9dd6dfee136ec58d0fa5990aa6e763b4` `source_revision_id=srcrev_cd15059203b45a6a2cd6b0b5a301810d` `chunk_id=srcchunk_916a7b54a95663a25edb9f37d46287ec` `native_locator=https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6` `source_timestamp=2024-12-03T05:20:00Z`
- The IModuleRegistry (referred to as IStoryProtocolACL in the source) emits ModuleAuthorizationGranted events when a gateway is authorized for a specific module dependency. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6) `source_document_id=srcdoc_9dd6dfee136ec58d0fa5990aa6e763b4` `source_revision_id=srcrev_cd15059203b45a6a2cd6b0b5a301810d` `chunk_id=srcchunk_916a7b54a95663a25edb9f37d46287ec` `native_locator=https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6` `source_timestamp=2024-12-03T05:20:00Z`
- The IModuleRegistry emits ModuleAdded and ModuleRemoved events when modules are enrolled or removed from the protocol. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6) `source_document_id=srcdoc_9dd6dfee136ec58d0fa5990aa6e763b4` `source_revision_id=srcrev_cd15059203b45a6a2cd6b0b5a301810d` `chunk_id=srcchunk_916a7b54a95663a25edb9f37d46287ec` `native_locator=https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6` `source_timestamp=2024-12-03T05:20:00Z`
- The IModuleRegistry includes a registerProtocolModule function to register a new module of a provided type to Story Protocol. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6) `source_document_id=srcdoc_9dd6dfee136ec58d0fa5990aa6e763b4` `source_revision_id=srcrev_cd15059203b45a6a2cd6b0b5a301810d` `chunk_id=srcchunk_916a7b54a95663a25edb9f37d46287ec` `native_locator=https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6` `source_timestamp=2024-12-03T05:20:00Z`
- The team has converged on the roles and responsibilities for frontends, but has not yet converged on other unspecified aspects. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6) `source_document_id=srcdoc_9dd6dfee136ec58d0fa5990aa6e763b4` `source_revision_id=srcrev_cd15059203b45a6a2cd6b0b5a301810d` `chunk_id=srcchunk_916a7b54a95663a25edb9f37d46287ec` `native_locator=https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6` `source_timestamp=2024-12-03T05:20:00Z`

## Open Questions

- What specific aspects of frontend roles and responsibilities have not yet been converged on?

## Sources

- `source_document_id`: `srcdoc_9dd6dfee136ec58d0fa5990aa6e763b4`
- `source_revision_id`: `srcrev_cd15059203b45a6a2cd6b0b5a301810d`
- `source_url`: [Notion source](https://www.notion.so/Frontend-Periphery-Contracts-9505d08fcabb4cf8848244b54f6d08d6)
