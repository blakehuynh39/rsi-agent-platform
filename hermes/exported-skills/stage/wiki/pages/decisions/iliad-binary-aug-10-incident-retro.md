---
title: "Aug 10 iliad Binary Issue Retro"
type: "decision"
slug: "decisions/iliad-binary-aug-10-incident-retro"
freshness: "2024-08-12T14:52:00Z"
tags:
  - "client-bug"
  - "iliad"
  - "incident"
  - "retrospective"
owners: []
source_revision_ids:
  - "srcrev_daed8067665995ec2d3bf731b2725c33"
conflict_state: "none"
---

# Aug 10 iliad Binary Issue Retro

## Summary

On August 10, partners reported crashes with the iliad binary. The root cause was a client software bug where the networkId changed to 1723217281, but the client assumed chainId equaled networkId. A fix parameterized chainId with a default of 1513. An additional issue was found where the ABI packed into the iliad binary did not match the latest pushed onto geth. The team identified missing sanity checks, lack of CI tests for validator operations, and dependency management issues as contributing factors.

## Claims

- On August 10, partners reported operations on the iliad binary causing crashes. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2) `source_document_id=srcdoc_ddfc32239ce984f8dec9cc5bf4ab51b9` `source_revision_id=srcrev_daed8067665995ec2d3bf731b2725c33` `chunk_id=srcchunk_f34cb5f8b511b75310495b6f44b70ae3` `native_locator=https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2` `source_timestamp=2024-08-12T14:52:00Z`
- The cause was quickly identified as the networkId switching to a new value of 1723217281 while the client assumed chainId to be the same as networkId. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2) `source_document_id=srcdoc_ddfc32239ce984f8dec9cc5bf4ab51b9` `source_revision_id=srcrev_daed8067665995ec2d3bf731b2725c33` `chunk_id=srcchunk_f34cb5f8b511b75310495b6f44b70ae3` `native_locator=https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2` `source_timestamp=2024-08-12T14:52:00Z`
- The fix was to parameterize chainId in the iliad CLI tool with a default value of 1513. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2) `source_document_id=srcdoc_ddfc32239ce984f8dec9cc5bf4ab51b9` `source_revision_id=srcrev_daed8067665995ec2d3bf731b2725c33` `chunk_id=srcchunk_f34cb5f8b511b75310495b6f44b70ae3` `native_locator=https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2` `source_timestamp=2024-08-12T14:52:00Z`
- Another issue was detected: the ABI packed into the iliad binary did not match the latest pushed onto geth. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2) `source_document_id=srcdoc_ddfc32239ce984f8dec9cc5bf4ab51b9` `source_revision_id=srcrev_daed8067665995ec2d3bf731b2725c33` `chunk_id=srcchunk_f34cb5f8b511b75310495b6f44b70ae3` `native_locator=https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2` `source_timestamp=2024-08-12T14:52:00Z`
- The biggest issue was not performing a full sanity check that all client operations would work prior to relaying the new release details to partners. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2) `source_document_id=srcdoc_ddfc32239ce984f8dec9cc5bf4ab51b9` `source_revision_id=srcrev_daed8067665995ec2d3bf731b2725c33` `chunk_id=srcchunk_f34cb5f8b511b75310495b6f44b70ae3` `native_locator=https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2` `source_timestamp=2024-08-12T14:52:00Z`
- There were no CI tests for the binary checking standard operations included in the partner testnet guide (create validator, stake, unstake). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2) `source_document_id=srcdoc_ddfc32239ce984f8dec9cc5bf4ab51b9` `source_revision_id=srcrev_daed8067665995ec2d3bf731b2725c33` `chunk_id=srcchunk_f34cb5f8b511b75310495b6f44b70ae3` `native_locator=https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2` `source_timestamp=2024-08-12T14:52:00Z`
- iliad requires a dependency on the forked geth when dealing with validator-related changes; the team should start packaging changes via releases. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2) `source_document_id=srcdoc_ddfc32239ce984f8dec9cc5bf4ab51b9` `source_revision_id=srcrev_daed8067665995ec2d3bf731b2725c33` `chunk_id=srcchunk_f34cb5f8b511b75310495b6f44b70ae3` `native_locator=https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2` `source_timestamp=2024-08-12T14:52:00Z`

## Sources

- `source_document_id`: `srcdoc_ddfc32239ce984f8dec9cc5bf4ab51b9`
- `source_revision_id`: `srcrev_daed8067665995ec2d3bf731b2725c33`
- `source_url`: [Notion source](https://www.notion.so/Aug-10-iliad-Binary-Issue-Retro-6f9caca19b5e460d8676c18b331de8d2)
