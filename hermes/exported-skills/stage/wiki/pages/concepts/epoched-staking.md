---
title: "Epoched Staking"
type: "concept"
slug: "concepts/epoched-staking"
freshness: "2026-05-05T06:38:27Z"
tags:
  - "epochs"
  - "evm"
  - "staking"
owners: []
source_revision_ids:
  - "srcrev_583615659d51369d4fe357ef9851044f"
conflict_state: "none"
---

# Epoched Staking

## Summary

Epoched staking mechanism that uses epochs to batch and execute validator set operations.

## Claims

- The epochs module handles epoch information. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`
- EpochInfo contains fields: Identifier, startTime, duration, currentEpoch, currentEpochStartTime, epochCountingStart, currentEpochStartHeight. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`
- Epoch info is updated every BeginBlock. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`
- When an epoch ends, it starts a new epoch with increment of epoch number by 1. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`
- Examples of epochs include hourly, daily, and weekly. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`
- The evmstaking module has params related to epoch for epoched staking: identifier and currentEpoch. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`
- Events related to validator set are queued in keeper: CreateValidator, Delegate, Redelegate, Withdraw, Unjail. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`
- Every EndBlock, check if a new epoch starts by comparing currentEpoch and epoch number of the epoch managed in epochs module. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`
- If epoch ends, the queued messages are executed. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766) `source_document_id=srcdoc_28de48dab6a8e086cfd96c53bc273206` `source_revision_id=srcrev_583615659d51369d4fe357ef9851044f` `chunk_id=srcchunk_c2183b0df68b7daa5e3474ec2176592e` `native_locator=https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766` `source_timestamp=2026-05-05T06:38:27Z`

## Sources

- `source_document_id`: `srcdoc_28de48dab6a8e086cfd96c53bc273206`
- `source_revision_id`: `srcrev_583615659d51369d4fe357ef9851044f`
- `source_url`: [Notion source](https://www.notion.so/Epoched-Staking-8c03539a4bd84d0a98c782139f3dc766)
