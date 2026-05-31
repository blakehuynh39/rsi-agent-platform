---
title: "Post-Mortem: iliad NFT mint high gas issue"
type: "runbook"
slug: "runbooks/post-mortem-iliad-nft-mint-high-gas-issue"
freshness: "2024-09-10T18:07:00Z"
tags:
  - "gas"
  - "iliad"
  - "nft"
  - "post-mortem"
  - "testnet"
owners: []
source_revision_ids:
  - "srcrev_9cd3dff05d06478b032c5baa59f6a1cc"
conflict_state: "none"
---

# Post-Mortem: iliad NFT mint high gas issue

## Summary

Post-mortem analysis of the iliad testnet NFT mint event that caused a massive gas price spike up to 7000 Gwei on August 27. The incident was triggered by a backup plan that removed Twitter authentication and smart contract gating after the backend hit Twitter OAuth rate limits, leading to uncontrolled minting. The page documents the timeline, impact metrics, and a set of recommendations for future launches.

## Claims

- The NFT mint website was launched as part of the iliad testnet launch to attract users and activate the community. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_86c258ba095c36c048bb8d4dbdf57563` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1` `source_timestamp=2024-09-10T18:07:00Z`
- The original PRD required every visitor to authenticate their X account and retweet the testnet launch tweet before minting, with a signature checking mechanism on the smart contract. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_86c258ba095c36c048bb8d4dbdf57563` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1` `source_timestamp=2024-09-10T18:07:00Z`
- Within a few hours after launch, the backend hit a rate limit issue with the Twitter OAuth API. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_86c258ba095c36c048bb8d4dbdf57563` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1` `source_timestamp=2024-09-10T18:07:00Z`
- At the time of the rate limit, the mempool had about 3-4k pending transactions. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_86c258ba095c36c048bb8d4dbdf57563` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1` `source_timestamp=2024-09-10T18:07:00Z`
- The backup plan was triggered: removed Twitter account checking, disabled the gate on the NFT mint smart contract, and allowed anyone to mint from the frontend or directly from the smart contract. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_86c258ba095c36c048bb8d4dbdf57563` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1` `source_timestamp=2024-09-10T18:07:00Z`
- The IP token distribution at the testnet faucet was increased from 1 to 10. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_86c258ba095c36c048bb8d4dbdf57563` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1` `source_timestamp=2024-09-10T18:07:00Z`
- On the night of August 27, a huge spike in NFT minting events caused the network gas price to reach up to 7000 Gwei. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_86c258ba095c36c048bb8d4dbdf57563` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1` `source_timestamp=2024-09-10T18:07:00Z`
- The high gas price persisted until the minting website was stopped. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_86c258ba095c36c048bb8d4dbdf57563` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-1` `source_timestamp=2024-09-10T18:07:00Z`
- The explorer reached 20k users in 2 days with 56k+ page views and over 200k+ user action events recorded. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-2) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_e8b50ddc9a10d4d55116f25562360f5c` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-2` `source_timestamp=2024-09-10T18:07:00Z`
- Peak average active users reached 300+ per 30 minutes. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-2) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_e8b50ddc9a10d4d55116f25562360f5c` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-2` `source_timestamp=2024-09-10T18:07:00Z`
- For future launches, alternative human verification methods should be considered. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- For the faucet app, switch to an alternative authentication approach with a signature requirement checked on the backend. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- Potentially explore Gitcoin Pass as a verification method. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- Offer a variety of OAuth providers (Google, Discord, World Coin) as options, with more coins for better authentication. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- Send user transaction requests to the backend to keep a record for potential appeasement to users who miss out on mints. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- Our own app should use private RPC endpoints, with backend and RPC in the same VPC and private bandwidth. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- 429 errors on testnet.storyscan.xyz were caused by the Cloudflare default setting of 50 requests/sec; updating to 200 req/sec solved the issue. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- Use DNS round robin to distribute RPC servers. `claim:claim_1_18` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- NFT minting calls SPG which adds extra gas cost; should call the protocol directly. `claim:claim_1_19` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- More verification methods other than Twitter should be offered, such as Google accounts. `claim:claim_1_20` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- Another popular verification method is checking the user's balance on mainnet to have at least 0.01. `claim:claim_1_21` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- To avoid gas wars, users should submit a signature and the backend can mint to their wallet over a longer period to smooth out the gas price peak. `claim:claim_1_22` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`
- Consider splitting the process into two phases: first phase collects whitelist via Twitter sign-up, second phase allows whitelisted users to mint. `claim:claim_1_23` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4) `source_document_id=srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21` `source_revision_id=srcrev_9cd3dff05d06478b032c5baa59f6a1cc` `chunk_id=srcchunk_361aef0f7bee2fd7b023eea22e82c0eb` `native_locator=https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd#chunk-4` `source_timestamp=2024-09-10T18:07:00Z`

## Sources

- `source_document_id`: `srcdoc_6f0ca3d62b8e045f26dd4d6da567cc21`
- `source_revision_id`: `srcrev_9cd3dff05d06478b032c5baa59f6a1cc`
- `source_url`: [Notion source](https://www.notion.so/Post-Mortem-iliad-NFT-mint-high-gas-issue-c20f5cc64589483483a2e0e4b3d910fd)
