---
title: "Access Control Complexity Analysis"
type: "concept"
slug: "concepts/access-control-complexity-analysis"
freshness: "2024-03-14T18:08:00Z"
tags:
  - "access-control"
  - "ACL"
  - "architecture"
  - "IPAccount"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_a248e46a70eb18e2d292da3b138b32e8"
conflict_state: "none"
---

# Access Control Complexity Analysis

## Summary

Analysis of the complexity and risks in the current internal access control system between IPAccounts and protocol contracts, including proposals for separating signer and delegator permissions.

## Claims

- Every checkPermission() method in the protocol follows the flow from IPAccount to verify permissions. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-1) `source_document_id=srcdoc_c223929e67db4982ab8576a9bf767d6e` `source_revision_id=srcrev_a248e46a70eb18e2d292da3b138b32e8` `chunk_id=srcchunk_5633b4840cb3d9c51876b7de6b434e51` `native_locator=https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-1` `source_timestamp=2024-03-14T18:08:00Z`
- AccessController must handle both internal and external flows, adding extra complexity which is riskier and less efficient. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2) `source_document_id=srcdoc_c223929e67db4982ab8576a9bf767d6e` `source_revision_id=srcrev_a248e46a70eb18e2d292da3b138b32e8` `chunk_id=srcchunk_4f515fd744fd714e3959fd2c4b35318f` `native_locator=https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2` `source_timestamp=2024-03-14T18:08:00Z`
- Global permissions are risky because they can be granted to a module for a legitimate reason, and later the module can steal from the IPAccount. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2) `source_document_id=srcdoc_c223929e67db4982ab8576a9bf767d6e` `source_revision_id=srcrev_a248e46a70eb18e2d292da3b138b32e8` `chunk_id=srcchunk_4f515fd744fd714e3959fd2c4b35318f` `native_locator=https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2` `source_timestamp=2024-03-14T18:08:00Z`
- Having global, module, and per-method level permissions requires differentiating between abstain (not set) and deny (actively removed permissions), whereas a boolean (has permission or not) could reduce complexity. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2) `source_document_id=srcdoc_c223929e67db4982ab8576a9bf767d6e` `source_revision_id=srcrev_a248e46a70eb18e2d292da3b138b32e8` `chunk_id=srcchunk_4f515fd744fd714e3959fd2c4b35318f` `native_locator=https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2` `source_timestamp=2024-03-14T18:08:00Z`
- Proposal to have IPAccount.isValidSigner() use AccessController.checkPermissions for two different sets of addresses: periphery contracts and signers (co-owners of the wallet). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2) `source_document_id=srcdoc_c223929e67db4982ab8576a9bf767d6e` `source_revision_id=srcrev_a248e46a70eb18e2d292da3b138b32e8` `chunk_id=srcchunk_4f515fd744fd714e3959fd2c4b35318f` `native_locator=https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2` `source_timestamp=2024-03-14T18:08:00Z`
- Proposal to not use checkPermissions() from modules for external calls; instead use Delegators, analogous to Lens Delegated Executors, allowing addresses to call methods on modules on behalf of executors without calling through IPAccount. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2) `source_document_id=srcdoc_c223929e67db4982ab8576a9bf767d6e` `source_revision_id=srcrev_a248e46a70eb18e2d292da3b138b32e8` `chunk_id=srcchunk_4f515fd744fd714e3959fd2c4b35318f` `native_locator=https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2` `source_timestamp=2024-03-14T18:08:00Z`
- Currently, all external contracts interacting with SPG or core protocol must have each IPAccount individually grant permission for the external contract to call SPG or core protocol on behalf of IPAccount for specific functions. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2) `source_document_id=srcdoc_c223929e67db4982ab8576a9bf767d6e` `source_revision_id=srcrev_a248e46a70eb18e2d292da3b138b32e8` `chunk_id=srcchunk_4f515fd744fd714e3959fd2c4b35318f` `native_locator=https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2` `source_timestamp=2024-03-14T18:08:00Z`
- For an IP & license marketplace contract, each IPAccount must approve Transfer License: Owner → IPAccount → Marketplace. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2) `source_document_id=srcdoc_c223929e67db4982ab8576a9bf767d6e` `source_revision_id=srcrev_a248e46a70eb18e2d292da3b138b32e8` `chunk_id=srcchunk_4f515fd744fd714e3959fd2c4b35318f` `native_locator=https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b#chunk-2` `source_timestamp=2024-03-14T18:08:00Z`

## Open Questions

- Can the tri-state permission model (abstain/deny/allow) be simplified to a boolean model?
- How to mitigate the risk of global permissions being exploited by previously trusted modules?
- Should the ACL be separated into signer permissions (for IPAccount.execute()) and delegator permissions (for direct module calls)?

## Sources

- `source_document_id`: `srcdoc_c223929e67db4982ab8576a9bf767d6e`
- `source_revision_id`: `srcrev_a248e46a70eb18e2d292da3b138b32e8`
- `source_url`: [Notion source](https://www.notion.so/6-Access-Control-Complexity-a78fc1e3363845b8a90333851f32ef0b)
