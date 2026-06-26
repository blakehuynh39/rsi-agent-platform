---
title: "Story-API Incident Log"
type: "runbook"
slug: "runbooks/story-api-incidents"
freshness: "2026-03-08T14:02:47Z"
tags:
  - "incidents"
  - "spike-protection"
  - "story-api"
owners: []
source_revision_ids:
  - "srcrev_0164320d94e22ef2212b2090ccb11839"
  - "srcrev_04981584077c67e4a38f1e930c6385a5"
  - "srcrev_14d523bc692ef25bd507d08aad42bfd4"
  - "srcrev_1545c2fdeccfc3b22f86a4f6c52bf9cb"
  - "srcrev_2ff752e1f4ad1e5e8d39934671f3a715"
  - "srcrev_379b45e6e702a1ae11d1b0440ef615ee"
  - "srcrev_3ec500bd96e008fe3703a41ba04ea937"
  - "srcrev_453fb7b54987d47369dbd44173781667"
  - "srcrev_6132bde7728b1c79df608d63c11d20b0"
  - "srcrev_63c2bedfe39075f7674fd1c75d32d9f4"
  - "srcrev_714aafecd06b4812d0a226a3123451c5"
  - "srcrev_752d8dfbdf5c92a161ade263cba6018a"
  - "srcrev_8b9e54067c6b33d524350ee2066b563f"
  - "srcrev_9393c10a664e8731b6abba51bc5f0bd3"
  - "srcrev_9a7dd4306308669877cf971be5ea2e82"
  - "srcrev_a93b1a67eb49a05c4d36b8a096476030"
  - "srcrev_b99acd3174f2ef1733436d7868687b7a"
  - "srcrev_c514c2240abc90a64b451f32e0dcad87"
  - "srcrev_c891d129f4c0ce4ebb05b1eb7efd7859"
  - "srcrev_df0e1895da131c758b246bd20586870f"
  - "srcrev_ef246ea5725f157389a07fd1fc29bb39"
  - "srcrev_f1c898a96aa68a2c60f5ff1d9325fb58"
conflict_state: "none"
---

# Story-API Incident Log

## Summary

Collection of story-api error events and spike protection activations/deactivations observed in Slack channel.

## Claims

- Spike protection for spans was activated on project story-api at 2025-12-22 10:50:56 UTC. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_df0e1895da131c758b246bd20586870f` `chunk_id=srcchunk_9cbdbd286653a844530f2a99d53a318c` `native_locator=slack:C07K3J4JTH6:1766400657.101659:1766400657.101659` `source_timestamp=2025-12-22T10:50:57Z`
- Spike protection for spans was deactivated on project story-api at 2025-12-22 11:40:07 UTC. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_9393c10a664e8731b6abba51bc5f0bd3` `chunk_id=srcchunk_9b997b4abcac93ac516006cbd8c575bc` `native_locator=slack:C07K3J4JTH6:1766403609.592829:1766403609.592829` `source_timestamp=2025-12-22T11:40:09Z`
- Spike protection for spans was activated again on 2025-12-23 11:51:02 UTC. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_c514c2240abc90a64b451f32e0dcad87` `chunk_id=srcchunk_e23397d214bd1df15d34a06e6a64ad3d` `native_locator=slack:C07K3J4JTH6:1766490664.608659:1766490664.608659` `source_timestamp=2025-12-23T11:51:04Z`
- Spike protection for spans was deactivated again on 2025-12-23 12:30:08 UTC. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_9a7dd4306308669877cf971be5ea2e82` `chunk_id=srcchunk_3cdc2b3b9f093f679ad4d13604db7cc3` `native_locator=slack:C07K3J4JTH6:1766493009.928139:1766493009.928139` `source_timestamp=2025-12-23T12:30:09Z`
- On an earlier occasion, spike protection was activated on 2026-01-28 12:58:51 UTC. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_379b45e6e702a1ae11d1b0440ef615ee` `chunk_id=srcchunk_da8f383daf60ed5a2b457ee748b0b657` `native_locator=slack:C07K3J4JTH6:1769605132.596599:1769605132.596599` `source_timestamp=2026-01-28T12:58:52Z`
- Spike protection was deactivated on 2026-01-28 13:30:11 UTC. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_a93b1a67eb49a05c4d36b8a096476030` `chunk_id=srcchunk_118a8103e203972532ae1cb2c4b3299e` `native_locator=slack:C07K3J4JTH6:1769607016.591659:1769607016.591659` `source_timestamp=2026-01-28T13:30:16Z`
- POST /hub/users/notifications failed with HTTP 500: Failed to list notifications. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_6132bde7728b1c79df608d63c11d20b0` `chunk_id=srcchunk_7861e6307d1b3cbc24afbf020574a2e0` `native_locator=slack:C07K3J4JTH6:1768318821.917419:1768318821.917419` `source_timestamp=2026-01-13T15:40:21Z`
- POST /api/v3/*any failed with HTTP 500: Internal Server Error. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_2ff752e1f4ad1e5e8d39934671f3a715` `chunk_id=srcchunk_44b483fd230d18c7820483d75bb046c4` `native_locator=slack:C07K3J4JTH6:1768319520.673739:1768319520.673739` `source_timestamp=2026-01-13T15:52:00Z`
- POST /api/v4/search failed with HTTP 500: Internal Server Error. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_04981584077c67e4a38f1e930c6385a5` `chunk_id=srcchunk_8266cf74f8f2f60810e9ead303d72d55` `native_locator=slack:C07K3J4JTH6:1768329761.070009:1768329761.070009` `source_timestamp=2026-01-13T18:42:41Z`
- POST /api/v4/assets/edges failed with HTTP 500: Internal Server Error. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_14d523bc692ef25bd507d08aad42bfd4` `chunk_id=srcchunk_b799ee9bd55e6550ab0fae1d459dab51` `native_locator=slack:C07K3J4JTH6:1771978937.140859:1771978937.140859` `source_timestamp=2026-02-25T00:22:17Z`
- POST /api/v4/licenses/tokens failed with HTTP 500: Internal Server Error. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_0164320d94e22ef2212b2090ccb11839` `chunk_id=srcchunk_4653a2251d7ee5d37d06ba7df10ac1d1` `native_locator=slack:C07K3J4JTH6:1771978941.115389:1771978941.115389` `source_timestamp=2026-02-25T00:22:21Z`
- POST /api/v4/transactions failed with HTTP 500: Internal Server Error. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_453fb7b54987d47369dbd44173781667` `chunk_id=srcchunk_8b02a6d37e7b070dee608c31fb02806f` `native_locator=slack:C07K3J4JTH6:1771978941.335999:1771978941.335999` `source_timestamp=2026-02-25T00:22:21Z`
- POST /api/v4/disputes failed with HTTP 500: Internal Server Error. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_3ec500bd96e008fe3703a41ba04ea937` `chunk_id=srcchunk_abe087262d3518f1daf17902e69e50f2` `native_locator=slack:C07K3J4JTH6:1771978942.800849:1771978942.800849` `source_timestamp=2026-02-25T00:22:22Z`
- GET /api/v4/disputes/:disputeId failed with HTTP 500: Internal Server Error. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_714aafecd06b4812d0a226a3123451c5` `chunk_id=srcchunk_8acbed96f5a5b896b3a34bf8917f589c` `native_locator=slack:C07K3J4JTH6:1771979411.809849:1771979411.809849` `source_timestamp=2026-02-25T00:30:11Z`
- Search error: `errors.withStack: failed to perform search: all retries failed: failed to unmarshal response: unexpected end of JSON input`. `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_b99acd3174f2ef1733436d7868687b7a` `chunk_id=srcchunk_dc2858c5a7224ca6fc20e75ba5a998df` `native_locator=slack:C07K3J4JTH6:1771982460.283569:1771982460.283569` `source_timestamp=2026-02-25T01:21:00Z`
- Search error: `errors.withStack: failed to perform search: all retries failed: received status code 500`. `claim:claim_1_16` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_f1c898a96aa68a2c60f5ff1d9325fb58` `chunk_id=srcchunk_b12d4abc40197deb174c105d9fc1293b` `native_locator=slack:C07K3J4JTH6:1771983427.572969:1771983427.572969` `source_timestamp=2026-02-25T01:37:07Z`
- Fatal database authentication error: `FATAL #28P01 password authentication failed for user "postgres"`. `claim:claim_1_17` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_752d8dfbdf5c92a161ade263cba6018a` `chunk_id=srcchunk_74861cda23622408408de2df2376026a` `native_locator=slack:C07K3J4JTH6:1772068619.578939:1772068619.578939` `source_timestamp=2026-02-26T01:16:59Z`
- IP asset creation failed with not-null constraint violation on column `blacklisted` in table `ip_assets`. `claim:claim_1_18` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_ef246ea5725f157389a07fd1fc29bb39` `chunk_id=srcchunk_933b2cc07efe96a6262bb90e0eb8dba0` `native_locator=slack:C07K3J4JTH6:1772155654.063019:1772155654.063019` `source_timestamp=2026-02-27T01:27:34Z`
- Entity creation failed with null value in column `id` of relation `users`. `claim:claim_1_19` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_1545c2fdeccfc3b22f86a4f6c52bf9cb` `chunk_id=srcchunk_9d6dca803e954efac329e4417ce9911c` `native_locator=slack:C07K3J4JTH6:1772168340.760199:1772168340.760199` `source_timestamp=2026-02-27T04:59:00Z`
- POST /api/v4/collections failed with HTTP 500: Internal Server Error. `claim:claim_1_20` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_63c2bedfe39075f7674fd1c75d32d9f4` `chunk_id=srcchunk_3b1fffc54ab96944d7d2107701f8dacd` `native_locator=slack:C07K3J4JTH6:1772502680.253939:1772502680.253939` `source_timestamp=2026-03-03T01:51:20Z`
- Parse error: `strconv.ParseInt: parsing "99999999999999999999999": value out of range`. `claim:claim_1_21` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_c891d129f4c0ce4ebb05b1eb7efd7859` `chunk_id=srcchunk_89d7fd90eb662ad70cb3dce802444e39` `native_locator=slack:C07K3J4JTH6:1772978546.338639:1772978546.338639` `source_timestamp=2026-03-08T14:02:26Z`
- Another GET /api/v4/disputes/:disputeId failed with HTTP 500: Internal server error. `claim:claim_1_22` `confidence:1.00`
  - citation: `source_document_id=srcdoc_5d3459afcedd0f5ec46a635b1efb5eea` `source_revision_id=srcrev_8b9e54067c6b33d524350ee2066b563f` `chunk_id=srcchunk_a378e5f08e3849c1c102363f6dc16b15` `native_locator=slack:C07K3J4JTH6:1772978567.054859:1772978567.054859` `source_timestamp=2026-03-08T14:02:47Z`

## Open Questions

- Are the 500 errors correlated with the spike protection events or independent?
- What causes the repeated spike protection activations for the story-api spans?
- Why does password authentication fail for the postgres user? Is it a transient issue?

## Sources

- `source_document_id`: `srcdoc_5d3459afcedd0f5ec46a635b1efb5eea`
- `source_revision_id`: `srcrev_9393c10a664e8731b6abba51bc5f0bd3`
