---
title: "L1 Test Plan"
type: "project"
slug: "projects/l1-test-plan"
freshness: "2024-07-05T17:36:00Z"
tags:
  - "l1"
  - "network"
  - "testing"
owners: []
source_revision_ids:
  - "srcrev_45f62e4367fa112ef356448f12675d22"
conflict_state: "none"
---

# L1 Test Plan

## Summary

Test plan for L1 network covering network health, stability, performance, fault tolerance, and user experience including validator, delegator, and developer experiences. Includes malicious node test scenarios.

## Claims

- Network test includes health checks via API health tests and basic functional tests such as account transfer and smart contract calls. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- Block production is verified by checking if new blocks are produced. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- Staking functions tested include deposit, withdraw, delegation, creating new validator, and reward distribution. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- Network stability and reliability are assessed via baseline transaction traffic and stress tests sending a high volume of transactions to see if the network stabilizes. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- Network performance is measured using a TPS test. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- Network fault tolerance testing includes node crash tests and malicious node tests. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- User experience testing covers validator experience, delegator experience, and staking dashboard. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- Developer experience testing covers RPC node, faucet, explorer, wallet, and smart contract development with Foundry. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- Malicious proposer test scenarios include messing up the withdraw queue ordering, providing wrong withdraw amounts, hiding a withdraw, adding a non-existent withdraw, adding a non-existent deposit, hiding a deposit, missing or extra EVM logs, and adding arbitrary deposit and withdraw transactions into CometBFT. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`
- Malicious voter test scenario includes double voting. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0) `source_document_id=srcdoc_7a971abb060a22e2ead5440e73f5ac58` `source_revision_id=srcrev_45f62e4367fa112ef356448f12675d22` `chunk_id=srcchunk_4d8dbbd39db7db880f51fbb9b3121f1f` `native_locator=https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0` `source_timestamp=2024-07-05T17:36:00Z`

## Sources

- `source_document_id`: `srcdoc_7a971abb060a22e2ead5440e73f5ac58`
- `source_revision_id`: `srcrev_45f62e4367fa112ef356448f12675d22`
- `source_url`: [Notion source](https://www.notion.so/L1-Test-plan-414052aa435d44a286aa48ccfb3637e0)
