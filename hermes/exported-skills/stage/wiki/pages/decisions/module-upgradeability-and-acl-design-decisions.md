---
title: "Module Upgradeability and IPAccount ACL Design Decisions"
type: "decision"
slug: "decisions/module-upgradeability-and-acl-design-decisions"
freshness: "2024-03-12T19:57:00Z"
tags:
  - "acl"
  - "delegation"
  - "ip-account"
  - "meta-transactions"
  - "modules"
  - "security"
  - "upgradeability"
owners:
  - "Leo"
  - "Raul"
source_revision_ids:
  - "srcrev_610e451684bd0621de02f20d76e2a80e"
conflict_state: "none"
---

# Module Upgradeability and IPAccount ACL Design Decisions

## Summary

Design decisions around module upgradeability, IPAccount ACL, delegation, and meta-transactions to unblock the security audit pipeline. Key decisions include adopting a common ModuleUpgradeable contract with a separate admin role, separating delegation from call-through functionality, and implementing meta-transaction patterns (e.g., `commentWithSig`) to reduce reliance on protocol modules for periphery contracts.

## Claims

- Two items must converge to unblock security audit work: upgradeability of modules and untangling the ACL, especially around IPAccount. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-1) `source_document_id=srcdoc_6444547ae72f5957e8ceed98eb305def` `source_revision_id=srcrev_610e451684bd0621de02f20d76e2a80e` `chunk_id=srcchunk_5f8e538d84e64c321bce4f4dd95be66a` `native_locator=https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-1` `source_timestamp=2024-03-12T19:57:00Z`
- The repository is very large for audit; features may need to be launched in sequence to allow auditors to properly review them, as lines of code directly impact audit pricing and protocol complexity. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-1) `source_document_id=srcdoc_6444547ae72f5957e8ceed98eb305def` `source_revision_id=srcrev_610e451684bd0621de02f20d76e2a80e` `chunk_id=srcchunk_5f8e538d84e64c321bce4f4dd95be66a` `native_locator=https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-1` `source_timestamp=2024-03-12T19:57:00Z`
- A common `Module` contract is needed to create a `ModuleUpgradeable`. An admin role for upgrading must be defined, ideally separate from the protocol admin. `Module` will also be `AccessControlled`. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-1) `source_document_id=srcdoc_6444547ae72f5957e8ceed98eb305def` `source_revision_id=srcrev_610e451684bd0621de02f20d76e2a80e` `chunk_id=srcchunk_5f8e538d84e64c321bce4f4dd95be66a` `native_locator=https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-1` `source_timestamp=2024-03-12T19:57:00Z`
- Setting a delegator for an action should not allow that delegator to call through the IPAccount. Delegation and call-through are separate functionalities with different trust assumptions. No current core module or registry needs call-through; periphery contracts can use it to perform actions on behalf of users. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-1) `source_document_id=srcdoc_6444547ae72f5957e8ceed98eb305def` `source_revision_id=srcrev_610e451684bd0621de02f20d76e2a80e` `chunk_id=srcchunk_5f8e538d84e64c321bce4f4dd95be66a` `native_locator=https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-1` `source_timestamp=2024-03-12T19:57:00Z`
- If meta-transactions are embraced at the protocol level, SPG and periphery contracts do not need to be protocol modules; they can use signature passing for authentication. The pattern from Lens (e.g., `comment` and `commentWithSig`) should be adopted, providing two versions of every method: one checking if the caller is the profile owner/delegator, and another expecting a signature. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-2) `source_document_id=srcdoc_6444547ae72f5957e8ceed98eb305def` `source_revision_id=srcrev_610e451684bd0621de02f20d76e2a80e` `chunk_id=srcchunk_11ad6a1a81a26b5340319832fd1094d6` `native_locator=https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-2` `source_timestamp=2024-03-12T19:57:00Z`
- Delegation is not mandatory but unlocks other contracts executing actions on behalf of IPAccounts without call-through (better gas, no generic execution debugging), relayers, and hot wallets/employees managing IP. Without it, mandatory call-through from IPAccount would cause friction and worse gas profile. The feature is minimal to implement and not very risky; it may even reduce risk in IPAccount. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-2) `source_document_id=srcdoc_6444547ae72f5957e8ceed98eb305def` `source_revision_id=srcrev_610e451684bd0621de02f20d76e2a80e` `chunk_id=srcchunk_11ad6a1a81a26b5340319832fd1094d6` `native_locator=https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-2` `source_timestamp=2024-03-12T19:57:00Z`
- Module upgradeability is considered a mandatory feature. Two module types are proposed: upgradeable and non-upgradeable. The IP Registry might be a candidate for non-upgradeable, but this is debatable. Testing must cover the upgrade mechanism, ACL soundness, and storage layout integrity; a team double-check process is also important. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-2) `source_document_id=srcdoc_6444547ae72f5957e8ceed98eb305def` `source_revision_id=srcrev_610e451684bd0621de02f20d76e2a80e` `chunk_id=srcchunk_11ad6a1a81a26b5340319832fd1094d6` `native_locator=https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8#chunk-2` `source_timestamp=2024-03-12T19:57:00Z`

## Open Questions

- Is the IP Registry truly non-upgradeable, or should it also be upgradeable?
- Should users be able to add their own modules?
- What is the exact process for the team to double-check upgrades before execution?

## Sources

- `source_document_id`: `srcdoc_6444547ae72f5957e8ceed98eb305def`
- `source_revision_id`: `srcrev_610e451684bd0621de02f20d76e2a80e`
- `source_url`: [Notion source](https://www.notion.so/Decisions-around-upgradeability-of-modules-and-IPAccount-ACL-eefbba500e9c45c3ac8eb321f701c7b8)
