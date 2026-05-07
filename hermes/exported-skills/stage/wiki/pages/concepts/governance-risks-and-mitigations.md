---
title: "Governance Risks and Mitigations"
type: "concept"
slug: "concepts/governance-risks-and-mitigations"
freshness: "2026-05-05T06:43:25Z"
tags:
  - "dao"
  - "governance"
  - "security"
  - "tokenomics"
owners: []
source_revision_ids:
  - "srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18"
conflict_state: "none"
---

# Governance Risks and Mitigations

## Summary

Overview of governance risks in DAOs and various mitigation strategies, including economic attacks, sub-DAOs, timelocks, security councils, delegation, and conviction voting.

## Claims

- Market mechanisms for token allocation fail to distinguish between users who want to make valuable contributions and attackers who want to disrupt or control the project, as both are behaviorally indistinguishable in a public marketplace. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-1) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_d46eb7297dabe4392f105691f33250a5` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-1` `source_timestamp=2026-05-05T06:43:25Z`
- Economic attacks on DAO governance include vote buying, plutocracy, and sybil attacks, which are tactics a wealthy bad actor can use to unfairly influence voting. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-1) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_d46eb7297dabe4392f105691f33250a5` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-1` `source_timestamp=2026-05-05T06:43:25Z`
- Projects particularly vulnerable to blatant economic attacks have fully on-chain governance, high TVL treasury managed by token voting, low price governance token, and low voter turnout. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-1) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_d46eb7297dabe4392f105691f33250a5` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-1` `source_timestamp=2026-05-05T06:43:25Z`
- MakerDAO, one of the first DAOs dedicated to maintaining the stablecoin DAI, has highly decentralized and process-heavy governance, which is slowing it down; as part of the 'Maker Endgame', smaller and more focused sub-DAOs will split from the main one. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_fb4671746d98214723148c9d361b4c42` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3` `source_timestamp=2026-05-05T06:43:25Z`
- Timelocks define a delay between when a proposal passes and when it can be executed, serving as a security measure for the community to cancel erroneous or malicious decisions and to allow members to exit or change their positions. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_fb4671746d98214723148c9d361b4c42` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3` `source_timestamp=2026-05-05T06:43:25Z`
- A security council or security multisig, ideally controlled by technical and security-aware signers appointed by governance, holds rights to cancel a vote, pause the protocol, and exercise other emergency powers; Arbitrum's security council is a notable example. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_fb4671746d98214723148c9d361b4c42` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3` `source_timestamp=2026-05-05T06:43:25Z`
- Story Protocol already supports security council functionality in its smart contracts via OpenZeppelin AccessManager Guardian roles. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_fb4671746d98214723148c9d361b4c42` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3` `source_timestamp=2026-05-05T06:43:25Z`
- There are already professional delegator services for DAOs, such as StableLab, and some voices call for incentivizing voters or delegators to boost participation, though skepticism remains due to cases where highly paid delegates did not vote due to legal liability concerns. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_fb4671746d98214723148c9d361b4c42` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3` `source_timestamp=2026-05-05T06:43:25Z`
- Conviction voting, proposed by BlockScience, mitigates last-minute vote swings by having a holder's vote gain governance power multiplier over time; if a whale switches close to the deadline, their multiplier resets while honest voters' multipliers keep growing. A con is that it is complicated for the average voter, which can deter participation. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3) `source_document_id=srcdoc_0a0e152c0d9864d60144b06e817ec706` `source_revision_id=srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18` `chunk_id=srcchunk_fb4671746d98214723148c9d361b4c42` `native_locator=https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680#chunk-3` `source_timestamp=2026-05-05T06:43:25Z`

## Sources

- `source_document_id`: `srcdoc_0a0e152c0d9864d60144b06e817ec706`
- `source_revision_id`: `srcrev_7f64dc489e2f9a64b6bd2fd14d57bb18`
- `source_url`: [Notion source](https://www.notion.so/Governance-Risks-and-Mitigations-f8ac46d7af714a72a2b5b4d6faefc680)
