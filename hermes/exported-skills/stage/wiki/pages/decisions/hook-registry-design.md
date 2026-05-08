---
title: "Hook Registry Design"
type: "decision"
slug: "decisions/hook-registry-design"
freshness: "2023-11-10T01:39:00Z"
tags:
  - "configuration"
  - "hooks"
  - "iporg"
  - "module"
  - "registry"
owners: []
source_revision_ids:
  - "srcrev_b3326b7dccd3bd924bed741cde2db0ff"
conflict_state: "none"
---

# Hook Registry Design

## Summary

Design decisions for the Hook Registry, including proposals for a global registry, per-IPOrg configuration, and context-key-based hook storage.

## Claims

- A global Hook Registry may be needed to provide a whitelist of valid hooks within Story Protocol, supporting the use case where an IPOrg owner needs to know which hooks are available and valid. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291) `source_document_id=srcdoc_8be8361eaca545c46557a4efce3a9994` `source_revision_id=srcrev_b3326b7dccd3bd924bed741cde2db0ff` `chunk_id=srcchunk_328ff14f1dceca38895c0ff3b5fea262` `native_locator=https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291` `source_timestamp=2023-11-10T01:39:00Z`
- A proposal for the global Hook Registry is to reuse the existing ModuleRegistry by adding keys like "Hook_payment", "Hook_nft_gated", and "Hook_license_term". `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291) `source_document_id=srcdoc_8be8361eaca545c46557a4efce3a9994` `source_revision_id=srcrev_b3326b7dccd3bd924bed741cde2db0ff` `chunk_id=srcchunk_328ff14f1dceca38895c0ff3b5fea262` `native_locator=https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291` `source_timestamp=2023-11-10T01:39:00Z`
- In HookRegistry, instead of storing a single array of pre/post action hooks, the design should use a mapping from IPOrg address to hook arrays, allowing per-IPOrg configuration. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291) `source_document_id=srcdoc_8be8361eaca545c46557a4efce3a9994` `source_revision_id=srcrev_b3326b7dccd3bd924bed741cde2db0ff` `chunk_id=srcchunk_328ff14f1dceca38895c0ff3b5fea262` `native_locator=https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291` `source_timestamp=2023-11-10T01:39:00Z`
- An alternative design separates hook configuration by a contextKey, mapping contextKey to hook arrays, where each module can define its own approach to generate the contextKey (e.g., IPA RegistrationModule uses ipOrg, RelationshipModule uses ipOrg and relType). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291) `source_document_id=srcdoc_8be8361eaca545c46557a4efce3a9994` `source_revision_id=srcrev_b3326b7dccd3bd924bed741cde2db0ff` `chunk_id=srcchunk_328ff14f1dceca38895c0ff3b5fea262` `native_locator=https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291` `source_timestamp=2023-11-10T01:39:00Z`
- The contextKey-based HookRegistry contract includes mappings for pre/post action hooks and their configurations, and a _registerHooks function that clears existing hooks before registering new ones. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291) `source_document_id=srcdoc_8be8361eaca545c46557a4efce3a9994` `source_revision_id=srcrev_b3326b7dccd3bd924bed741cde2db0ff` `chunk_id=srcchunk_328ff14f1dceca38895c0ff3b5fea262` `native_locator=https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291` `source_timestamp=2023-11-10T01:39:00Z`
- BaseModule extends HookRegistry and enforces hook execution without calling an external contract. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291) `source_document_id=srcdoc_8be8361eaca545c46557a4efce3a9994` `source_revision_id=srcrev_b3326b7dccd3bd924bed741cde2db0ff` `chunk_id=srcchunk_328ff14f1dceca38895c0ff3b5fea262` `native_locator=https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291` `source_timestamp=2023-11-10T01:39:00Z`

## Open Questions

- Should a global Hook Registry be implemented, or is the per-module local registry sufficient?
- Which approach is preferred: per-IPOrg mapping or contextKey-based mapping for hook configuration?

## Sources

- `source_document_id`: `srcdoc_8be8361eaca545c46557a4efce3a9994`
- `source_revision_id`: `srcrev_b3326b7dccd3bd924bed741cde2db0ff`
- `source_url`: [Notion source](https://www.notion.so/Hook-Open-Design-Questions-8320fc49d37343b6b9bb51df11e16291)
