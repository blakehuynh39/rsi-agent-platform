---
title: "Module System"
type: "system"
slug: "systems/module-system"
freshness: "2024-03-12T06:16:00Z"
tags:
  - "execution-flow"
  - "module"
  - "module-types"
  - "trust-levels"
owners: []
source_revision_ids:
  - "srcrev_25fb199b4f55072c6fb0398c899f6e54"
conflict_state: "none"
---

# Module System

## Summary

Design of the module system including execution flow, interface, trust levels, and types.

## Claims

- Modules can be called directly by other modules without going through IPAccount, using a modifier `onlyModule(moduleName)` that checks the caller against the MODULE_REGISTRY. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-1) `source_document_id=srcdoc_be79ae3fb047443be69e41e3e4ca238a` `source_revision_id=srcrev_25fb199b4f55072c6fb0398c899f6e54` `chunk_id=srcchunk_4652ea561ea089a21189053ac3381222` `native_locator=https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-1` `source_timestamp=2024-03-12T06:16:00Z`
- Trusted modules can call external contracts through IPAccount, similar to Safe, ERC6900, and ERC7579 patterns. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-1) `source_document_id=srcdoc_be79ae3fb047443be69e41e3e4ca238a` `source_revision_id=srcrev_25fb199b4f55072c6fb0398c899f6e54` `chunk_id=srcchunk_4652ea561ea089a21189053ac3381222` `native_locator=https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-1` `source_timestamp=2024-03-12T06:16:00Z`
- Modules must implement a `version()` function returning a string version. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-2) `source_document_id=srcdoc_be79ae3fb047443be69e41e3e4ca238a` `source_revision_id=srcrev_25fb199b4f55072c6fb0398c899f6e54` `chunk_id=srcchunk_7cc8f691995270a54a62b7dd86e9babe` `native_locator=https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-2` `source_timestamp=2024-03-12T06:16:00Z`
- Three trust levels: Permissionless Registered Module (can be called by IPAccount, can call another module through IPAccount), Verified Modules (can call external contracts through IPAccount), Core Modules (can be registered as Function/Fallback Handler). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-2) `source_document_id=srcdoc_be79ae3fb047443be69e41e3e4ca238a` `source_revision_id=srcrev_25fb199b4f55072c6fb0398c899f6e54` `chunk_id=srcchunk_7cc8f691995270a54a62b7dd86e9babe` `native_locator=https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-2` `source_timestamp=2024-03-12T06:16:00Z`
- Module types: HookModule renamed to CheckModule, added ViewModule, added Fallback/Function Handler Module. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-2) `source_document_id=srcdoc_be79ae3fb047443be69e41e3e4ca238a` `source_revision_id=srcrev_25fb199b4f55072c6fb0398c899f6e54` `chunk_id=srcchunk_7cc8f691995270a54a62b7dd86e9babe` `native_locator=https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948#chunk-2` `source_timestamp=2024-03-12T06:16:00Z`

## Related Pages

- `ipaccount-design`

## Sources

- `source_document_id`: `srcdoc_be79ae3fb047443be69e41e3e4ca238a`
- `source_revision_id`: `srcrev_25fb199b4f55072c6fb0398c899f6e54`
- `source_url`: [Notion source](https://www.notion.so/Protocol-Post-Beta-13190882a2d9405480bec5bd7fdf4948)
