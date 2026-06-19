---
title: "Disperse.app Integration for Batch Token Transfers"
type: "project"
slug: "projects/disperse-app-batch-token-transfers"
freshness: "2026-05-07T00:13:19Z"
tags:
  - "batch-transfers"
  - "disperse-app"
  - "Story"
  - "token-distribution"
  - "USDC.e"
owners:
  - "U0871SH0FNZ"
  - "U09QGMMUDPC"
source_revision_ids:
  - "srcrev_530f61d71c5cec2bdd3104df6574052b"
  - "srcrev_681b27c5b8f5c301cc38dfdd1585d28c"
  - "srcrev_c3e1488cc54de99453001618c5b74e51"
  - "srcrev_e7066a8a46eab618a7a15c25abe6852f"
conflict_state: "none"
---

# Disperse.app Integration for Batch Token Transfers

## Summary

Investigation into using Disperse.app for cost-effective batch token transfers on Story, with potential for programmatic automation to handle crypto payouts.

## Claims

- Disperse.app enables batch sending of tokens to many addresses cheaply via a disperse function. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_321bc574a452e50c125b933cca808550` `source_revision_id=srcrev_e7066a8a46eab618a7a15c25abe6852f` `chunk_id=srcchunk_3793cd2185a1b69b88caa3693499aca8` `native_locator=slack:C0AL7EKNHDF:1778101166.857099:1778101166.857099` `source_timestamp=2026-05-06T20:59:26Z`
- The service works on Story (USDC.e) and other chains (USDC/T). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_321bc574a452e50c125b933cca808550` `source_revision_id=srcrev_e7066a8a46eab618a7a15c25abe6852f` `chunk_id=srcchunk_3793cd2185a1b69b88caa3693499aca8` `native_locator=slack:C0AL7EKNHDF:1778101166.857099:1778101166.857099` `source_timestamp=2026-05-06T20:59:26Z`
- The UI allows manual import of recipients and amounts via CSV. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_321bc574a452e50c125b933cca808550` `source_revision_id=srcrev_530f61d71c5cec2bdd3104df6574052b` `chunk_id=srcchunk_62e1a6dfc0b012e1647c1a20ffd1d528` `native_locator=slack:C0AL7EKNHDF:1778101166.857099:1778103825.625399` `source_timestamp=2026-05-06T21:43:45Z`
- There is approximately a 600-transfers per transaction limit due to gas constraints. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_321bc574a452e50c125b933cca808550` `source_revision_id=srcrev_530f61d71c5cec2bdd3104df6574052b` `chunk_id=srcchunk_62e1a6dfc0b012e1647c1a20ffd1d528` `native_locator=slack:C0AL7EKNHDF:1778101166.857099:1778103825.625399` `source_timestamp=2026-05-06T21:43:45Z`
- There is no hard limit on total quantity of recipient addresses; the tool has been used for airdrops involving thousands of addresses. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_321bc574a452e50c125b933cca808550` `source_revision_id=srcrev_681b27c5b8f5c301cc38dfdd1585d28c` `chunk_id=srcchunk_af1bf973ba207aeb86eefd6acd43e66c` `native_locator=slack:C0AL7EKNHDF:1778101166.857099:1778112755.465929` `source_timestamp=2026-05-07T00:12:35Z`
- A programmatic approach is desired to automate the process and synchronize status on both sides. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_321bc574a452e50c125b933cca808550` `source_revision_id=srcrev_530f61d71c5cec2bdd3104df6574052b` `chunk_id=srcchunk_62e1a6dfc0b012e1647c1a20ffd1d528` `native_locator=slack:C0AL7EKNHDF:1778101166.857099:1778103825.625399` `source_timestamp=2026-05-06T21:43:45Z`
  - citation: `source_document_id=srcdoc_321bc574a452e50c125b933cca808550` `source_revision_id=srcrev_c3e1488cc54de99453001618c5b74e51` `chunk_id=srcchunk_b5eb89399544d3fa0c30011a1e145f3e` `native_locator=slack:C0AL7EKNHDF:1778101166.857099:1778112799.028069` `source_timestamp=2026-05-07T00:13:19Z`

## Open Questions

- How can payment status be tracked and updated automatically alongside the batch transfer?
- Is there a programmatic API for Disperse.app, or can the smart contract be called directly?
- What are the exact gas limits and optimization strategies for transferring large numbers of recipients on Story?

## Sources

- `source_document_id`: `srcdoc_321bc574a452e50c125b933cca808550`
- `source_revision_id`: `srcrev_c3e1488cc54de99453001618c5b74e51`
