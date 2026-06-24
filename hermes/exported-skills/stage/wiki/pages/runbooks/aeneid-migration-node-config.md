---
title: "Aeneid Migration Node Configuration"
type: "runbook"
slug: "runbooks/aeneid-migration-node-config"
freshness: "2026-01-29T08:04:21Z"
tags:
  - "aeneid"
  - "archive-node"
  - "migration"
  - "node-config"
  - "pruning"
owners:
  - "U079ZJ48D62"
  - "U07A7AUGL5V"
  - "U08332YRB7W"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_54a62aa7847d01295b613d82136d7300"
  - "srcrev_620e90b01000ebf4daff2d1034e67250"
  - "srcrev_92bc1160d03ad720c6978f693bf3485f"
  - "srcrev_c66e9373281c4bccbfd4bac5dd550873"
  - "srcrev_d5cb001d34b93ead8c0fa946e757c55a"
  - "srcrev_f626e8dc3b27c7993da65c3fb145cca0"
conflict_state: "none"
---

# Aeneid Migration Node Configuration

## Summary

Server configuration for the Aeneid migration, including story and geth setups, and pruning settings for archive nodes.

## Claims

- Aeneid migration documentation was created and shared on Notion at https://www.notion.so/storyprotocol/Story-Aeneid-Migration-2f6051299a5480ff97bbee675f7d067e `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_d5cb001d34b93ead8c0fa946e757c55a` `chunk_id=srcchunk_4d40bcb636fe8f23d04b53f65c8a9b3e` `native_locator=slack:C0547N89JUB:1769648276.485679:1769648276.485679` `source_timestamp=2026-01-29T00:57:56Z`
- The server configuration is similar to mainnet. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_d5cb001d34b93ead8c0fa946e757c55a` `chunk_id=srcchunk_4d40bcb636fe8f23d04b53f65c8a9b3e` `native_locator=slack:C0547N89JUB:1769648276.485679:1769648276.485679` `source_timestamp=2026-01-29T00:57:56Z`
- Node packages were upgraded to v1.5.2 except the archive snapshot which is still being upgraded. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_d5cb001d34b93ead8c0fa946e757c55a` `chunk_id=srcchunk_4d40bcb636fe8f23d04b53f65c8a9b3e` `native_locator=slack:C0547N89JUB:1769648276.485679:1769648276.485679` `source_timestamp=2026-01-29T00:57:56Z`
- Story node runs using cosmovisor with command: /usr/local/bin/cosmovisor run run --api-enable --api-address=0.0.0.0:1317 `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_92bc1160d03ad720c6978f693bf3485f` `chunk_id=srcchunk_1804e397e621b543a21f5b720de0af2b` `native_locator=slack:C0547N89JUB:1769649992.721089` `source_timestamp=2026-01-29T01:26:32Z`
- For archive nodes, pruning in story.toml should be set to 'nothing' to keep all historical state. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_620e90b01000ebf4daff2d1034e67250` `chunk_id=srcchunk_7738d1820e73c6f0470c7a707ad6389c` `native_locator=slack:C0547N89JUB:1769673351.311449` `source_timestamp=2026-01-29T07:55:51Z`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_c66e9373281c4bccbfd4bac5dd550873` `chunk_id=srcchunk_abd6851e2328f5ba9a5dcb5a9b632fc9` `native_locator=slack:C0547N89JUB:1769673861.426279` `source_timestamp=2026-01-29T08:04:21Z`
- The geth configuration looks fine. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_54a62aa7847d01295b613d82136d7300` `chunk_id=srcchunk_b826089f1c1aa2e47dd06d53e97d42a8` `native_locator=slack:C0547N89JUB:1769653719.284439` `source_timestamp=2026-01-29T02:28:39Z`
- If migrating from GCP to AWS using a snapshot at blockheight X, syncing from there is fastest. The pruning configuration choice (default vs nothing) does not require rebuilding the node from scratch; it can be changed and the node will continue syncing. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_f626e8dc3b27c7993da65c3fb145cca0` `chunk_id=srcchunk_6bef21adf9c5678e0b3bf5b3ca02b201` `native_locator=slack:C0547N89JUB:1769673681.622729` `source_timestamp=2026-01-29T08:01:23Z`
  - citation: `source_document_id=srcdoc_998dff1c966caff947043f1e8ac3282b` `source_revision_id=srcrev_c66e9373281c4bccbfd4bac5dd550873` `chunk_id=srcchunk_abd6851e2328f5ba9a5dcb5a9b632fc9` `native_locator=slack:C0547N89JUB:1769673861.426279` `source_timestamp=2026-01-29T08:04:21Z`

## Sources

- `source_document_id`: `srcdoc_998dff1c966caff947043f1e8ac3282b`
- `source_revision_id`: `srcrev_933c4edda895c878bc8cff83f552d040`
