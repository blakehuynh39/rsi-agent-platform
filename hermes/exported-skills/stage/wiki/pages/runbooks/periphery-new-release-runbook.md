---
title: "Periphery New Release Runbook"
type: "runbook"
slug: "runbooks/periphery-new-release-runbook"
freshness: "2024-09-10T17:01:00Z"
tags:
  - "checklist"
  - "periphery"
  - "release"
owners: []
source_revision_ids:
  - "srcrev_028fc6163c9ea0f546921ed54f8a404a"
conflict_state: "none"
---

# Periphery New Release Runbook

## Summary

Checklist of steps to perform a new release of Periphery, including updating package.json, dependencies, changelog, deployment addresses, documentation, creating a PR, creating a GitHub release, and uploading signatures to OpenChain.

## Claims

- Update package.json `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`
- Update dependencies `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`
- Update CHANGELOG.md `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`
- Update deploy-out deployment addresses `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`
- Update deployment addresses in root-level README.md `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`
- Update relevant documentations in /docs `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`
- Create a PR for above changes and merge `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`
- Create a new release on Github `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`
- Upload function/error signatures to OpenChain using forge selectors upload --all `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218) `source_document_id=srcdoc_0da9ddbeaab34493707a2ecf14c17ce5` `source_revision_id=srcrev_028fc6163c9ea0f546921ed54f8a404a` `chunk_id=srcchunk_90359eef6b688e1aac9c944e1f7caed2` `native_locator=https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218` `source_timestamp=2024-09-10T17:01:00Z`

## Sources

- `source_document_id`: `srcdoc_0da9ddbeaab34493707a2ecf14c17ce5`
- `source_revision_id`: `srcrev_028fc6163c9ea0f546921ed54f8a404a`
- `source_url`: [Notion source](https://www.notion.so/Periphery-New-Release-Runbook-bd4a4d9180704f9daef690a38d19f218)
