---
title: "New CI Process"
type: "decision"
slug: "decisions/new-ci-process"
freshness: "2024-09-01T01:27:00Z"
tags:
  - "ci-cd"
  - "genesis-state"
  - "network-management"
owners: []
source_revision_ids:
  - "srcrev_262b6fa7751390d752e84fc29b496451"
conflict_state: "none"
---

# New CI Process

## Summary

Proposal for a new CI/CD process to manage genesis state and network deployments for stable and unstable networks, ensuring a single source of truth and automated promotion.

## Claims

- There are four kinds of artifacts: consensus client, execution client, consensus genesis, and execution genesis. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1) `source_document_id=srcdoc_7edd44a2c5edf6d09712f2f5894543d0` `source_revision_id=srcrev_262b6fa7751390d752e84fc29b496451` `chunk_id=srcchunk_2b0bbe933e81b6cb4f518ec6219149ca` `native_locator=https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1` `source_timestamp=2024-09-01T01:27:00Z`
- Networks are classified as stable (Iliad, Partner Testnet) and unstable (mininet, devnet). `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1) `source_document_id=srcdoc_7edd44a2c5edf6d09712f2f5894543d0` `source_revision_id=srcrev_262b6fa7751390d752e84fc29b496451` `chunk_id=srcchunk_2b0bbe933e81b6cb4f518ec6219149ca` `native_locator=https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1` `source_timestamp=2024-09-01T01:27:00Z`
- Current CI/CD issues include no single source of truth for genesis state for unstable networks, manual updates in node-launcher, and no guarantee that developers update genesis files or promote state correctly. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1) `source_document_id=srcdoc_7edd44a2c5edf6d09712f2f5894543d0` `source_revision_id=srcrev_262b6fa7751390d752e84fc29b496451` `chunk_id=srcchunk_2b0bbe933e81b6cb4f518ec6219149ca` `native_locator=https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1` `source_timestamp=2024-09-01T01:27:00Z`
- For unstable networks, clients and genesis state should only be promoted, never manually updated, and should follow a linear promotion flow (e.g., mininet → devnet → staging). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1) `source_document_id=srcdoc_7edd44a2c5edf6d09712f2f5894543d0` `source_revision_id=srcrev_262b6fa7751390d752e84fc29b496451` `chunk_id=srcchunk_2b0bbe933e81b6cb4f518ec6219149ca` `native_locator=https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1` `source_timestamp=2024-09-01T01:27:00Z`
- There should be a single source of truth for execution genesis state and consensus genesis state, and any change affecting genesis state should trigger automated workflows for updating that genesis state. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1) `source_document_id=srcdoc_7edd44a2c5edf6d09712f2f5894543d0` `source_revision_id=srcrev_262b6fa7751390d752e84fc29b496451` `chunk_id=srcchunk_2b0bbe933e81b6cb4f518ec6219149ca` `native_locator=https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-1` `source_timestamp=2024-09-01T01:27:00Z`
- On EL contract changes merged to main, an automated robot should push the corresponding EL genesis state hash to the story repo. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-2) `source_document_id=srcdoc_7edd44a2c5edf6d09712f2f5894543d0` `source_revision_id=srcrev_262b6fa7751390d752e84fc29b496451` `chunk_id=srcchunk_02ac7c5f99fde51073915bcc99e443d2` `native_locator=https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-2` `source_timestamp=2024-09-01T01:27:00Z`
- For network deployment, geth client updates trigger rolling upgrades; story client updates trigger rolling upgrades; consensus genesis changes trigger a push to node-launcher and a network restart promotion process that validates hashes and promotes genesis state across networks. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-2) `source_document_id=srcdoc_7edd44a2c5edf6d09712f2f5894543d0` `source_revision_id=srcrev_262b6fa7751390d752e84fc29b496451` `chunk_id=srcchunk_02ac7c5f99fde51073915bcc99e443d2` `native_locator=https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-2` `source_timestamp=2024-09-01T01:27:00Z`
- Three unstable networks are proposed for testing: dev, staging, prod, with current equivalents being mininet, devnet, and an unspecified third. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-2) `source_document_id=srcdoc_7edd44a2c5edf6d09712f2f5894543d0` `source_revision_id=srcrev_262b6fa7751390d752e84fc29b496451` `chunk_id=srcchunk_02ac7c5f99fde51073915bcc99e443d2` `native_locator=https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb#chunk-2` `source_timestamp=2024-09-01T01:27:00Z`

## Open Questions

- What is the third unstable network (prod) equivalent? (mininet / devnet / ???)

## Sources

- `source_document_id`: `srcdoc_7edd44a2c5edf6d09712f2f5894543d0`
- `source_revision_id`: `srcrev_262b6fa7751390d752e84fc29b496451`
- `source_url`: [Notion source](https://www.notion.so/New-CI-Process-60c16e74501545daad79ff044eded2fb)
