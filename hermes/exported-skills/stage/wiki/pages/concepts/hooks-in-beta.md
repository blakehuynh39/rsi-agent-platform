---
title: "Hooks in Beta"
type: "concept"
slug: "concepts/hooks-in-beta"
freshness: "2024-02-05T20:14:00Z"
tags:
  - "beta"
  - "hooks"
  - "smart-contracts"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_76e6ae62a1738b8dedf7b9ca2940f71d"
conflict_state: "none"
---

# Hooks in Beta

## Summary

Design notes on hooks for the Beta version, including differences from Alpha, interface, ACL considerations, and proposed integration points in the policy framework.

## Claims

- Hooks in Beta only return true or false via the interface IHook.verifyCondition(address caller, bytes data) returns (bool). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891) `source_document_id=srcdoc_3e915055185c4047142039ee7cba69ba` `source_revision_id=srcrev_76e6ae62a1738b8dedf7b9ca2940f71d` `chunk_id=srcchunk_77dd4c45e625c40a5ab718f384d28470` `native_locator=https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891` `source_timestamp=2024-02-05T20:14:00Z`
- There are no assumptions on flow; asynchronous condition setting is the responsibility of the implementing contract (e.g., LicensorApprovalChecker sets state, hook just verifies). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891) `source_document_id=srcdoc_3e915055185c4047142039ee7cba69ba` `source_revision_id=srcrev_76e6ae62a1738b8dedf7b9ca2940f71d` `chunk_id=srcchunk_77dd4c45e625c40a5ab718f384d28470` `native_locator=https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891` `source_timestamp=2024-02-05T20:14:00Z`
- ACL should not be enforced in the hook interface; each hook contract can choose its own approach, and an immutable allowed caller set on creation is cheaper than ACL to avoid expensive minting/linking/transfer. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891) `source_document_id=srcdoc_3e915055185c4047142039ee7cba69ba` `source_revision_id=srcrev_76e6ae62a1738b8dedf7b9ca2940f71d` `chunk_id=srcchunk_77dd4c45e625c40a5ab718f384d28470` `native_locator=https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891` `source_timestamp=2024-02-05T20:14:00Z`
- Potential hook integration points: verifyMint() and verifyLink() in IPolicyFrameworkManager; in UMLPolicyFrameworkManager, commercializers hook (e.g., token gating), licensor approval hook (converting LicensorApprovalChecker), and update() in LicenseRegistry for transfer checks. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891) `source_document_id=srcdoc_3e915055185c4047142039ee7cba69ba` `source_revision_id=srcrev_76e6ae62a1738b8dedf7b9ca2940f71d` `chunk_id=srcchunk_77dd4c45e625c40a5ab718f384d28470` `native_locator=https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891` `source_timestamp=2024-02-05T20:14:00Z`
- HookRegistry is proposed as a standalone contract; UMLPolicyFramework selects hooks based on policy parameters; licensors configure hooks in UMLPolicy (e.g., commercializer as hook address + bytes data). Policy is copied to derivatives, so hook configuration persists (e.g., all derivatives of Emergence World Bible with ERC721OwnerHook must own the same NFT). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891) `source_document_id=srcdoc_3e915055185c4047142039ee7cba69ba` `source_revision_id=srcrev_76e6ae62a1738b8dedf7b9ca2940f71d` `chunk_id=srcchunk_77dd4c45e625c40a5ab718f384d28470` `native_locator=https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891` `source_timestamp=2024-02-05T20:14:00Z`

## Sources

- `source_document_id`: `srcdoc_3e915055185c4047142039ee7cba69ba`
- `source_revision_id`: `srcrev_76e6ae62a1738b8dedf7b9ca2940f71d`
- `source_url`: [Notion source](https://www.notion.so/Hooks-in-Beta-8e500fa2c8044cbd9c517b7b08e44891)
