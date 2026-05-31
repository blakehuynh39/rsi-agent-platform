---
title: "SDK V1.3 Design Considerations"
type: "decision"
slug: "decisions/sdk-v1-3-design-considerations"
freshness: "2024-09-16T23:17:00Z"
tags:
  - "abstraction"
  - "design"
  - "gas"
  - "sdk"
  - "viem"
owners: []
source_revision_ids:
  - "srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf"
conflict_state: "none"
---

# SDK V1.3 Design Considerations

## Summary

Design feedback on the SDK V1.3 regarding generated contract functions, gas configuration, and transaction timeout handling, advocating for more user-facing abstraction levels.

## Claims

- In SDK V1, most contract interactions used generated calls to simulated contracts, which limited the ability to configure gas for smart contract calls. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296) `source_document_id=srcdoc_a3e7702755f60f2924a99a3f32ba7649` `source_revision_id=srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf` `chunk_id=srcchunk_9041686b9bcbf40c95d668d1e089f445` `native_locator=https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296` `source_timestamp=2024-09-16T23:17:00Z`
- The use of generated functions makes it difficult to provide users with multiple levels of abstraction for integrating complex behaviors. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296) `source_document_id=srcdoc_a3e7702755f60f2924a99a3f32ba7649` `source_revision_id=srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf` `chunk_id=srcchunk_9041686b9bcbf40c95d668d1e089f445` `native_locator=https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296` `source_timestamp=2024-09-16T23:17:00Z`
- Users should be able to intuitively set gas-related parameters and timeouts on transactions in the SDK. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296) `source_document_id=srcdoc_a3e7702755f60f2924a99a3f32ba7649` `source_revision_id=srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf` `chunk_id=srcchunk_9041686b9bcbf40c95d668d1e089f445` `native_locator=https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296` `source_timestamp=2024-09-16T23:17:00Z`
- A function mapping readme was written and is available at a GitHub Gist URL. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296) `source_document_id=srcdoc_a3e7702755f60f2924a99a3f32ba7649` `source_revision_id=srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf` `chunk_id=srcchunk_9041686b9bcbf40c95d668d1e089f445` `native_locator=https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296` `source_timestamp=2024-09-16T23:17:00Z`

## Related Pages

- `sdk-v1-3-contract-structure`

## Sources

- `source_document_id`: `srcdoc_a3e7702755f60f2924a99a3f32ba7649`
- `source_revision_id`: `srcrev_efd0bd3fbf02f7d5eddc0a5bda977caf`
- `source_url`: [Notion source](https://www.notion.so/SDK-V1-3-Design-Doc-103051299a5480d8acd0c982cecc8296)
