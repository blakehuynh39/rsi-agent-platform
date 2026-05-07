---
title: "TMKMS"
type: "system"
slug: "systems/tmkms"
freshness: "2024-09-26T17:37:00Z"
tags:
  - "remote-signing"
  - "security"
  - "signing"
  - "validator"
owners: []
source_revision_ids:
  - "srcrev_5d1c80397cba6aeaba6a230fd5feaa83"
conflict_state: "none"
---

# TMKMS

## Summary

TMKMS is a remote validator signing service that separates the validator node from the signing process, enhancing security by preventing key compromise on the validator node.

## Claims

- TMKMS is used for remote validator signing, separating the validator node from the process actually performing the signing. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1) `source_document_id=srcdoc_e9fae3552a116603de3ff5edcf9b8a14` `source_revision_id=srcrev_5d1c80397cba6aeaba6a230fd5feaa83` `chunk_id=srcchunk_cfd8c5a668b865ab854858397031c9c0` `native_locator=https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1` `source_timestamp=2024-09-26T17:37:00Z`
- The default story validator node uses plaintext priv_validator_key.json for consensus signing, which risks key compromise if the node is accessed by malicious users. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1) `source_document_id=srcdoc_e9fae3552a116603de3ff5edcf9b8a14` `source_revision_id=srcrev_5d1c80397cba6aeaba6a230fd5feaa83` `chunk_id=srcchunk_cfd8c5a668b865ab854858397031c9c0` `native_locator=https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1` `source_timestamp=2024-09-26T17:37:00Z`
- The softsign provider in TMKMS allows remote signing via a separate machine using software. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1) `source_document_id=srcdoc_e9fae3552a116603de3ff5edcf9b8a14` `source_revision_id=srcrev_5d1c80397cba6aeaba6a230fd5feaa83` `chunk_id=srcchunk_cfd8c5a668b865ab854858397031c9c0` `native_locator=https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1` `source_timestamp=2024-09-26T17:37:00Z`
- The TMKMS machine requires rust, gcc, pkg-config, and libusb to be installed. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1) `source_document_id=srcdoc_e9fae3552a116603de3ff5edcf9b8a14` `source_revision_id=srcrev_5d1c80397cba6aeaba6a230fd5feaa83` `chunk_id=srcchunk_cfd8c5a668b865ab854858397031c9c0` `native_locator=https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1` `source_timestamp=2024-09-26T17:37:00Z`
- Installation involves exporting RUSTFLAGS for x86_64, installing tmkms with cargo, initializing with tmkms init, generating a secret connection key, copying the priv_validator_key.json from the validator node, importing it, and modifying tmkms.toml. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1) `source_document_id=srcdoc_e9fae3552a116603de3ff5edcf9b8a14` `source_revision_id=srcrev_5d1c80397cba6aeaba6a230fd5feaa83` `chunk_id=srcchunk_cfd8c5a668b865ab854858397031c9c0` `native_locator=https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1` `source_timestamp=2024-09-26T17:37:00Z`
- The chain ID configured in tmkms.toml is "iliad-0". `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1) `source_document_id=srcdoc_e9fae3552a116603de3ff5edcf9b8a14` `source_revision_id=srcrev_5d1c80397cba6aeaba6a230fd5feaa83` `chunk_id=srcchunk_cfd8c5a668b865ab854858397031c9c0` `native_locator=https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-1` `source_timestamp=2024-09-26T17:37:00Z`
- TMKMS successfully signs consensus votes; example log shows signing of Prevote and Precommit at height 1424, round 0. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-2) `source_document_id=srcdoc_e9fae3552a116603de3ff5edcf9b8a14` `source_revision_id=srcrev_5d1c80397cba6aeaba6a230fd5feaa83` `chunk_id=srcchunk_9a6da09564b497afaa2662d0e8eeb000` `native_locator=https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1#chunk-2` `source_timestamp=2024-09-26T17:37:00Z`

## Sources

- `source_document_id`: `srcdoc_e9fae3552a116603de3ff5edcf9b8a14`
- `source_revision_id`: `srcrev_5d1c80397cba6aeaba6a230fd5feaa83`
- `source_url`: [Notion source](https://www.notion.so/TMKMS-10c051299a54802a8285f582a50612a1)
