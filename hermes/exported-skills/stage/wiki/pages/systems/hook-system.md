---
title: "Hook System Design"
type: "system"
slug: "systems/hook-system"
freshness: "2025-07-22T00:24:00Z"
tags:
  - "extensibility"
  - "hook"
  - "module"
owners: []
source_revision_ids:
  - "srcrev_2f41248749cd362ab2769d8a6ee9c2c4"
conflict_state: "none"
---

# Hook System Design

## Summary

Design of the hook system for extending and customizing module behavior, supporting synchronous and asynchronous operations, with a generic interface, registry, and role-based access.

## Claims

- The hook system defines a generic IHook interface with executeSync and executeAsync functions. `claim:claim_hook_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-1) `source_document_id=srcdoc_18632e830ceb29b74723c900fe1509d7` `source_revision_id=srcrev_2f41248749cd362ab2769d8a6ee9c2c4` `chunk_id=srcchunk_1bf712cf2d83d7ca309520bcd6a595f6` `native_locator=https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-1` `source_timestamp=2025-07-22T00:24:00Z`
- A HookRegistry contract keeps track of all available hooks and the module it is associated with. `claim:claim_hook_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-1) `source_document_id=srcdoc_18632e830ceb29b74723c900fe1509d7` `source_revision_id=srcrev_2f41248749cd362ab2769d8a6ee9c2c4` `chunk_id=srcchunk_1bf712cf2d83d7ca309520bcd6a595f6` `native_locator=https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-1` `source_timestamp=2025-07-22T00:24:00Z`
- Modules have a standardized BaseModule contract with _executeHooks to interact with the hook system. `claim:claim_hook_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-1) `source_document_id=srcdoc_18632e830ceb29b74723c900fe1509d7` `source_revision_id=srcrev_2f41248749cd362ab2769d8a6ee9c2c4` `chunk_id=srcchunk_1bf712cf2d83d7ca309520bcd6a595f6` `native_locator=https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-1` `source_timestamp=2025-07-22T00:24:00Z`
- Synchronous hooks are executed in a loop over all registered hooks, calling executeSync on each. `claim:claim_hook_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-2) `source_document_id=srcdoc_18632e830ceb29b74723c900fe1509d7` `source_revision_id=srcrev_2f41248749cd362ab2769d8a6ee9c2c4` `chunk_id=srcchunk_9c2bfb6bde6d36d56349ee305a86d009` `native_locator=https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-2` `source_timestamp=2025-07-22T00:24:00Z`
- Asynchronous hooks are executed in a non-blocking manner via event-emission and callback mechanism. `claim:claim_hook_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-2) `source_document_id=srcdoc_18632e830ceb29b74723c900fe1509d7` `source_revision_id=srcrev_2f41248749cd362ab2769d8a6ee9c2c4` `chunk_id=srcchunk_9c2bfb6bde6d36d56349ee305a86d009` `native_locator=https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-2` `source_timestamp=2025-07-22T00:24:00Z`
- When a module registers a hook, the HookRegistry contract grants the HOOK_CALLER_ROLE to that module to allow it to execute the hook. `claim:claim_hook_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-3) `source_document_id=srcdoc_18632e830ceb29b74723c900fe1509d7` `source_revision_id=srcrev_2f41248749cd362ab2769d8a6ee9c2c4` `chunk_id=srcchunk_96d78146613354c166650aa4a2bb74d6` `native_locator=https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c#chunk-3` `source_timestamp=2025-07-22T00:24:00Z`

## Open Questions

- There are open design questions for hooks, referenced in the source document.

## Related Pages

- `decisions/access-control-acl`

## Sources

- `source_document_id`: `srcdoc_18632e830ceb29b74723c900fe1509d7`
- `source_revision_id`: `srcrev_2f41248749cd362ab2769d8a6ee9c2c4`
- `source_url`: [Notion source](https://www.notion.so/Hook-bedaf5df6d4a46dfa03b0e1977ac0b2c)
