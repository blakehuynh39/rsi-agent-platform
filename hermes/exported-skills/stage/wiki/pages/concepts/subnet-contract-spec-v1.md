---
title: "Subnet Contract Spec V1"
type: "concept"
slug: "concepts/subnet-contract-spec-v1"
freshness: "2025-09-11T02:12:00Z"
tags:
  - "contract"
  - "specification"
  - "subnet"
  - "testing"
  - "worker"
  - "workflow"
owners: []
source_revision_ids:
  - "srcrev_8f44f6d242c9b23843daf4673866d383"
conflict_state: "none"
---

# Subnet Contract Spec V1

## Summary

Detailed functional test cases for the Subnet Contract V1 covering worker registration, workflow management, activity execution, rewards, access control, upgrades, and configuration.

## Claims

- The Subnet Contract Spec V1 defines functional test cases for worker registration and staking, including registration with a minimum stake of 100 PSDN tokens (WF-001) and rejection for insufficient stake (WF-002). `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-1) `source_document_id=srcdoc_b8b6ab1c492ace2285e8f44b6ceb09b0` `source_revision_id=srcrev_8f44f6d242c9b23843daf4673866d383` `chunk_id=srcchunk_c0b5a92e6b18a7add12ee3448105736e` `native_locator=https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-1` `source_timestamp=2025-09-11T02:12:00Z`
- The spec includes test cases for workflow management such as registering a new workflow definition (WF-007), preventing duplicate workflow names (WF-008), and starting a workflow instance (WF-009). `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-1) `source_document_id=srcdoc_b8b6ab1c492ace2285e8f44b6ceb09b0` `source_revision_id=srcrev_8f44f6d242c9b23843daf4673866d383` `chunk_id=srcchunk_c0b5a92e6b18a7add12ee3448105736e` `native_locator=https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-1` `source_timestamp=2025-09-11T02:12:00Z`
- Activity execution test cases cover worker polling for activities (WF-014) and claiming activities, with additional cases for rewards, access control, upgrades, and configuration updates. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-1) `source_document_id=srcdoc_b8b6ab1c492ace2285e8f44b6ceb09b0` `source_revision_id=srcrev_8f44f6d242c9b23843daf4673866d383` `chunk_id=srcchunk_c0b5a92e6b18a7add12ee3448105736e` `native_locator=https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-1` `source_timestamp=2025-09-11T02:12:00Z`
  - citation: [Notion source](https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-2) `source_document_id=srcdoc_b8b6ab1c492ace2285e8f44b6ceb09b0` `source_revision_id=srcrev_8f44f6d242c9b23843daf4673866d383` `chunk_id=srcchunk_3560025cd9c82c0ce894890a9a358c6e` `native_locator=https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-2` `source_timestamp=2025-09-11T02:12:00Z`
- Edge case testing includes handling head-of-line blocking with expired activities (WF-046) and processing empty workflow results gracefully (WF-047). `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-2) `source_document_id=srcdoc_b8b6ab1c492ace2285e8f44b6ceb09b0` `source_revision_id=srcrev_8f44f6d242c9b23843daf4673866d383` `chunk_id=srcchunk_3560025cd9c82c0ce894890a9a358c6e` `native_locator=https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d#chunk-2` `source_timestamp=2025-09-11T02:12:00Z`

## Related Pages

- `concepts/subnet-specifications`

## Sources

- `source_document_id`: `srcdoc_b8b6ab1c492ace2285e8f44b6ceb09b0`
- `source_revision_id`: `srcrev_8f44f6d242c9b23843daf4673866d383`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Contract-Spec-V1-269051299a5480fbb12cd38434a55d7d)
