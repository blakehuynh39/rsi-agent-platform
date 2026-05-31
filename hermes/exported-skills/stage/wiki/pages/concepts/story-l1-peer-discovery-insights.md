---
title: "Story L1 Peer Discovery Insights"
type: "concept"
slug: "concepts/story-l1-peer-discovery-insights"
freshness: "2024-10-02T23:46:00Z"
tags:
  - "consensus"
  - "geth"
  - "networking"
  - "peer-discovery"
  - "story-l1"
owners: []
source_revision_ids:
  - "srcrev_d70592364b1e7ce5f36332f30f6ccf1a"
conflict_state: "none"
---

# Story L1 Peer Discovery Insights

## Summary

Observations on how geth and consensus (CometBFT) nodes discover and maintain peers, including the effects of NoDiscovery, BootstrapNodes, seed mode, and PEX settings.

## Claims

- Setting NoDiscovery to true in geth prevents the node from PINGing other nodes for discovery, stopping it from finding peers to connect to. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_e9f4f9219569ffc88a0278f7b2825877` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6` `source_timestamp=2024-10-02T23:46:00Z`
- If a bootnode has NoDiscovery=false and other nodes have NoDiscovery=true, the bootnode gets inbound connections but does not relay peer info; if bootnode has NoDiscovery=true and others false, no peers are found. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_e9f4f9219569ffc88a0278f7b2825877` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6` `source_timestamp=2024-10-02T23:46:00Z`
- Consensus seed nodes disconnect after sending peer addresses; they can be set with seed_mode=true (requires PEX on). `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_e9f4f9219569ffc88a0278f7b2825877` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6` `source_timestamp=2024-10-02T23:46:00Z`
- If PEX is false and no persistent_peers are set, consensus nodes cannot connect to any other clients. `claim:claim_2_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_e9f4f9219569ffc88a0278f7b2825877` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6` `source_timestamp=2024-10-02T23:46:00Z`
- Unlike geth, consensus nodes will eventually reconnect to a bootnode after it comes back, even if launched without it initially. `claim:claim_2_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_e9f4f9219569ffc88a0278f7b2825877` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6` `source_timestamp=2024-10-02T23:46:00Z`
- When a seed node is in seed mode and PEX is true, all nodes connect initially, the seed disconnects, but later all nodes reconnect with the seed after 10-15 seconds. `claim:claim_2_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6) `source_document_id=srcdoc_63839ac5d351533add9aa1535af1a240` `source_revision_id=srcrev_d70592364b1e7ce5f36332f30f6ccf1a` `chunk_id=srcchunk_e9f4f9219569ffc88a0278f7b2825877` `native_locator=https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14#chunk-6` `source_timestamp=2024-10-02T23:46:00Z`

## Related Pages

- `story-l1-two-node-local-setup`

## Sources

- `source_document_id`: `srcdoc_63839ac5d351533add9aa1535af1a240`
- `source_revision_id`: `srcrev_d70592364b1e7ce5f36332f30f6ccf1a`
- `source_url`: [Notion source](https://www.notion.so/Mac-OS-X-Story-L1-Two-Node-Setup-e6f7641845bd40959d1f2b2893c0ba14)
