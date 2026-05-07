---
title: "Decision: Dynamic vs. Static Resolving Module Addresses for Function Calls"
type: "decision"
slug: "decisions/decision-dynamic-vs-static-module-address-resolution"
freshness: "2024-03-15T22:40:00Z"
tags:
  - "access-control"
  - "architecture"
  - "gas-optimization"
  - "module-registry"
owners: []
source_revision_ids:
  - "srcrev_d9e965c30c1437bf3a211515b773058d"
conflict_state: "none"
---

# Decision: Dynamic vs. Static Resolving Module Addresses for Function Calls

## Summary

Decision on whether to use dynamic resolving (reading addresses from ModuleRegistry) or static setting (immutable variables) for module addresses in gated function calls. The context involves core modules like LicensingModule and RoyaltyModule that need to permission specific callers.

## Claims

- LicenseRegistry's mintLicense() can only be called by LicensingModule. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- RoyaltyModule's onLicenseMinting() can only be called by LicensingModule. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- Gated functions require a specific modifier to permission callers, e.g., LicenseRegistry checks caller is LicensingModule for mintLicense. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- There are two distinctive options: dynamic resolving vs. static setting of module addresses. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- Dynamic resolving uses keys to read module addresses from ModuleRegistry on gated function calls, either via onlyModule(KEY) for all or module-specific modifiers in core modules. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- A benefit of dynamic resolving is that modules can be redeployed anytime without updating their addresses in dependent modules or registries. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- If core modules are made upgradable, the benefit of dynamic resolving for redeployment is useless because the proxy address remains constant. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- Dynamic resolving incurs additional gas cost: 2100 gas for the first SLOAD, 100 for subsequent in the same transaction, plus a STATICCALL to get the address from another contract. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- The Access Controller should support primitive and common access gating within the protocol. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`
- Stakeholders include Raul and Kingter; relevant participant is noted; decision deadline is 03/14/2024. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012) `source_document_id=srcdoc_eec9b63d323b876c0973b85f2b4030f9` `source_revision_id=srcrev_d9e965c30c1437bf3a211515b773058d` `chunk_id=srcchunk_778597acdf90f58ccbe2ad76c9e84d2b` `native_locator=https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012` `source_timestamp=2024-03-15T22:40:00Z`

## Open Questions

- Final decision between dynamic resolving and static setting not yet documented in this chunk.
- Whether core modules will be made upgradable, which affects the benefit of dynamic resolving.

## Related Pages

- `access-controller`
- `licensing-module`
- `module-registry`
- `royalty-module`

## Sources

- `source_document_id`: `srcdoc_eec9b63d323b876c0973b85f2b4030f9`
- `source_revision_id`: `srcrev_d9e965c30c1437bf3a211515b773058d`
- `source_url`: [Notion source](https://www.notion.so/3-Dynamic-vs-Static-Resolving-Module-Addresses-for-Function-Calls-4422d31dc5e648acbcb377cdb1996012)
