---
title: "Access Control (ACL) Design"
type: "decision"
slug: "decisions/access-control-acl"
freshness: "2024-03-12T00:53:00Z"
tags:
  - "access-control"
  - "architecture"
  - "licensing"
  - "permissions"
  - "roles"
owners: []
source_revision_ids:
  - "srcrev_958caf5fcdc25869779a562df974c938"
conflict_state: "none"
---

# Access Control (ACL) Design

## Summary

Design document for the protocol's Access Control (ACL) contracts and roles, including AccessControlSingleton, ModuleRegistry, and OZ Access Manager, along with use cases for Licensing Module, Disputer delegation, Tagging, and Registration.

## Claims

- The AccessControlSingleton contract maps roleId (bytes32) to an address and supports role admins. `claim:claim_acl_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`
- AccessControlSingleton can check roles for msg.sender using a modifier and for other addresses via `_hasRole(roleId, address)`. `claim:claim_acl_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`
- Defined protocol roles include PROTOCOL_ADMIN_ROLE (bytes32(0)), UPGRADER_ROLE, RELATIONSHIP_MANAGER_ROLE, LICENSING_MANAGER_ROLE, MODULE_REGISTRAR_ROLE, MODULE_EXECUTOR_ROLE, and HOOK_CALLER_ROLE, each derived via keccak256. `claim:claim_acl_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`
- ModuleRegistry maps a module key to an address and authorizes modules/gateways per selector via `mapping(ModuleKey => mapping(IGateway => mapping(bytes4 => bool))) _isAuthorized;`. `claim:claim_acl_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`
- OZ 5 Access Manager provides sequential roleId to address mapping, timelocks for execution and role granting, role admins and guardians, and the ability to pause a target for incident response. `claim:claim_acl_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`
- Licensing Module ACL needs include revoker, derivatives allowed (reciprocal) flag, whether minter must be licensor based on reciprocal flag, and delegation to mint licenses (sub licensing). `claim:claim_acl_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`
- Disputer delegation will be used to flag formal disputes. `claim:claim_acl_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`
- Tagging ACL: Alpha uses protocol relationship types via protocol role; org relationship types are managed by IP Org owner and relationship setting is gated through hooks. Beta is undefined. `claim:claim_acl_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`
- Registration ACL: Alpha IP Org authorizes registration through hooks; Beta licensing enforces access control for minters of derivatives. `claim:claim_acl_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762) `source_document_id=srcdoc_a70a5e58c80f5d28c91475f0fe379c52` `source_revision_id=srcrev_958caf5fcdc25869779a562df974c938` `chunk_id=srcchunk_317582b9657e06f3e1abb789629a1c0a` `native_locator=https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762` `source_timestamp=2024-03-12T00:53:00Z`

## Related Pages

- `decisions/ip-asset-on-chain-metadata-design`

## Sources

- `source_document_id`: `srcdoc_a70a5e58c80f5d28c91475f0fe379c52`
- `source_revision_id`: `srcrev_958caf5fcdc25869779a562df974c938`
- `source_url`: [Notion source](https://www.notion.so/ACL-41c8726cfc344753b9cc786b2a3fb762)
