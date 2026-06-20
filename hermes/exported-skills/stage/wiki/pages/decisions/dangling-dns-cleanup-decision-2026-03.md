---
title: "Decision: Remove Dangling DNS Domains (March 2026)"
type: "decision"
slug: "decisions/dangling-dns-cleanup-decision-2026-03"
freshness: "2026-03-30T17:02:33Z"
tags:
  - "cleanup"
  - "decommissioning"
  - "dns"
  - "security"
owners:
  - "Woojin"
source_revision_ids:
  - "srcrev_5a5f28ffdbb0fabbe02b97b21304443e"
  - "srcrev_744d06dfd0c5cb186845ac70c9b95ab6"
  - "srcrev_78d5044ef36bcefe8c11d6a5b5817fc3"
  - "srcrev_9fe1f438bdf731b8358704a7c2247793"
  - "srcrev_a75db093bbaee392424c4fbc29b0580f"
  - "srcrev_c5bb86bcfe8e8c9e5d5285865cc8fc41"
conflict_state: "none"
---

# Decision: Remove Dangling DNS Domains (March 2026)

## Summary

The team reviewed a list of 23 dangling DNS domains under storyapis.com and storyprotocol.net and decided to remove them, with explicit caution noted for domains that might be temporarily scaled down or planned for future deployment.

## Claims

- Woojin proposed removing 23 DNS domains that appeared dangling, listing each with its status (502, 403, timeout, etc.) and reason (e.g., 'GCP service removed', 'Odyssey network deprecated'). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aaf5451ab2fd1c105f1b10cfd4740f1` `source_revision_id=srcrev_c5bb86bcfe8e8c9e5d5285865cc8fc41` `chunk_id=srcchunk_08b2647dcc0a847eaeb73d85fb2e8d38` `native_locator=slack:C0547N89JUB:1774423285.762959:1774423285.762959` `source_timestamp=2026-03-25T07:21:25Z`
- SecBot advised caution before deletion, flagging stg-faucet.storyapis.com ('scaled down' might mean paused, not removed), cloudbeaver-poseidon.storyapis.com ('0 replicas' could be intentional for future scaling), and hey-api/hey-demo domains ('not deployed' could indicate planned deployment). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aaf5451ab2fd1c105f1b10cfd4740f1` `source_revision_id=srcrev_78d5044ef36bcefe8c11d6a5b5817fc3` `chunk_id=srcchunk_5523b6e332ce6c23e91c6ff06c118504` `native_locator=slack:C0547N89JUB:1774423285.762959:1774423302.720809` `source_timestamp=2026-03-25T07:21:42Z`
- Woojin stated he would proceed with the removal the next day if no objections were raised. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aaf5451ab2fd1c105f1b10cfd4740f1` `source_revision_id=srcrev_9fe1f438bdf731b8358704a7c2247793` `chunk_id=srcchunk_b8d2a5051cf9d2f067e875c9f021de7c` `native_locator=slack:C0547N89JUB:1774423285.762959:1774849769.354599` `source_timestamp=2026-03-30T05:49:29Z`
- Multiple team members confirmed the removal plan, with Andy (U05A515NBFC), Boris (U04L0DD6B6F), and another (U0772SH7BRA) all indicating it looked good. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_8aaf5451ab2fd1c105f1b10cfd4740f1` `source_revision_id=srcrev_a75db093bbaee392424c4fbc29b0580f` `chunk_id=srcchunk_25d8096097513f02e7463102096be3d0` `native_locator=slack:C0547N89JUB:1774423285.762959:1774851286.105559` `source_timestamp=2026-03-30T06:14:46Z`
  - citation: `source_document_id=srcdoc_8aaf5451ab2fd1c105f1b10cfd4740f1` `source_revision_id=srcrev_5a5f28ffdbb0fabbe02b97b21304443e` `chunk_id=srcchunk_d5d6554dfbb8610f90845ee2aac29679` `native_locator=slack:C0547N89JUB:1774423285.762959:1774851444.540589` `source_timestamp=2026-03-30T06:17:24Z`
  - citation: `source_document_id=srcdoc_8aaf5451ab2fd1c105f1b10cfd4740f1` `source_revision_id=srcrev_744d06dfd0c5cb186845ac70c9b95ab6` `chunk_id=srcchunk_627ff672fd509a1f67c63c280e8e044f` `native_locator=slack:C0547N89JUB:1774423285.762959:1774890153.345529` `source_timestamp=2026-03-30T17:02:33Z`

## Sources

- `source_document_id`: `srcdoc_8aaf5451ab2fd1c105f1b10cfd4740f1`
- `source_revision_id`: `srcrev_744d06dfd0c5cb186845ac70c9b95ab6`
