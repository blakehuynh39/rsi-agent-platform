---
title: "L1 Internal Devnet Roadmap"
type: "project"
slug: "projects/l1-internal-devnet-roadmap"
freshness: "2024-06-01T04:42:00Z"
tags:
  - "api"
  - "devnet"
  - "l1"
  - "partner-integration"
  - "roadmap"
  - "sdk"
owners:
  - "Andy"
  - "Haodi"
  - "Lutty"
  - "Zerui"
source_revision_ids:
  - "srcrev_260a2ac56e92a9f0b8f5ba3ca2786888"
conflict_state: "none"
---

# L1 Internal Devnet Roadmap

## Summary

Weekly roadmap tracking L1 devnet setup, SDK development, API optimization, and partner integration from May 13 to June 3.

## Claims

- Week of May 13: L1 team set up repos (cl, el, bft), removed custom cross-chain code, set up local test environment running cl and el together, and set up a 5-node devnet environment. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 20: L1 team removed custom cross-chain code, set up 15-node devnet environment, and automated devnet setup. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 27: L1 team enabled EVM end-to-end flow, set up block scout (Andy), performed JSON RPC tests (Haodi & Lutty), set up CI/CD workflow (Andy), used e2e docker local deployment, set up 4+1 devnet (Zerui), and planned consensus/performance testing. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of June 3: L1 team planned monitoring with Prometheus for geth and comet, listing public APIs in Postman, performance testing, end-to-end user testing, onboarding docs, faucet app, and integration support for staking and predeploy/precompile. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 13: SDK team supported SPG function registerWithMetadata, performed bug fixes, and explored Python SDK with task planning. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 20: SDK team completed new permission and delegation functions (BatchPermission and AllPermission), performed bug fixes, and worked on Python SDK royalty and dispute, publishing the first pip package. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 27: SDK team performed bug fixes and worked on Python SDK SPG functions, permissions, and IPAccounts. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of June 3: SDK team planned to publish the first version of the React SDK. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 13: API team optimized list ipAsset calls to reduce latency, promoted staging API to prod, planned API Gateway optimization, and performed bug fixes. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 20: API team deployed new list ipAsset changes to reduce latency, figured out missing params in license term structure, and performed bug fixes. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 27: API team deployed new list ipAsset changes to production and performed bug fixes. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of June 3: API team planned to complete Hub Backend Design. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 13: Partner Integration team scoped Magma, held a Rarible call, and held a Wormhole call for cross-chain API exploration. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`
- Week of May 20 and May 27: Partner Integration team worked on cross-chain API PoC registration flow. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720) `source_document_id=srcdoc_b0be50d4369d435d4f317063e61f8d03` `source_revision_id=srcrev_260a2ac56e92a9f0b8f5ba3ca2786888` `chunk_id=srcchunk_c3ccfe689e5b4186c30abf80ef81121e` `native_locator=https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720` `source_timestamp=2024-06-01T04:42:00Z`

## Sources

- `source_document_id`: `srcdoc_b0be50d4369d435d4f317063e61f8d03`
- `source_revision_id`: `srcrev_260a2ac56e92a9f0b8f5ba3ca2786888`
- `source_url`: [Notion source](https://www.notion.so/L1-Internal-devnet-fb96c079395b4c58939b308c2dc3b720)
