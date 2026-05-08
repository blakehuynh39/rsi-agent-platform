---
title: "Omni Security Fixes"
type: "project"
slug: "projects/omni-security-fixes"
freshness: "2024-10-16T05:12:00Z"
tags:
  - "audit"
  - "fixes"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_c54a3e7ba7def370d6029a135ebdb302"
conflict_state: "none"
---

# Omni Security Fixes

## Summary

Collection of security-related pull requests and improvements in the Omni project, including Spearbit audit fixes.

## Claims

- PR #2130 avoids need for nil checks with array types, simplifying code. `claim:pr-2130` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #2124 adds support for referencing other repo issues. `claim:pr-2124` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #2035 adds a readiness endpoint. `claim:pr-2035` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #2086 addresses low-level security findings in evmengine, tx types, and valsync keeper. `claim:pr-2086` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #2081 implements safer casts from slices to arrays repo-wide, part of the Spearbit audit. `claim:pr-2081` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #2094 uses a safer package for checking byte lengths. `claim:pr-2094` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #2059 adds OS architecture to build info for safer debugging. `claim:pr-2059` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #2036 enables use of custom RPCs in Docker containers. `claim:pr-2036` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #1996 smooths cosmovisor upgrades by preventing panic or consensus failure logs during upgrades. `claim:pr-1996` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #1956 adds a discv4 debug test. `claim:pr-1956` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #1907 adds support for Cosmos gRPC and REST APIs. `claim:pr-1907` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #1893 refactors caches to avoid mutations of cached values after setting/getting, part of the Spearbit audit. `claim:pr-1893` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #1892 addresses more Spearbit audit issues. `claim:pr-1892` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #1840 adds a create-consensus-key command. `claim:pr-1840` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`
- PR #1818 decreases timeout_propose from 3s to 1s to mitigate slow blocks when a validator is offline. `claim:pr-1818` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf) `source_document_id=srcdoc_b2f58e35cbd623275dfa492dadf7d24f` `source_revision_id=srcrev_c54a3e7ba7def370d6029a135ebdb302` `chunk_id=srcchunk_50bbdc20f24103505b5c7ed7518b128b` `native_locator=https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf` `source_timestamp=2024-10-16T05:12:00Z`

## Sources

- `source_document_id`: `srcdoc_b2f58e35cbd623275dfa492dadf7d24f`
- `source_revision_id`: `srcrev_c54a3e7ba7def370d6029a135ebdb302`
- `source_url`: [Notion source](https://www.notion.so/Omni-Security-Fixes-120051299a5480d7a93feb69230699cf)
