---
title: "Data Broadcaster"
type: "system"
slug: "systems/data-broadcaster"
freshness: "2026-05-27T22:32:00Z"
tags:
  - "blockchain"
  - "data-broadcaster"
  - "dynamodb"
  - "postgres"
  - "story-l1"
owners:
  - "Data Engineering"
source_revision_ids:
  - "srcrev_41419f35f81d8994463b51129d16e317"
conflict_state: "none"
---

# Data Broadcaster

## Summary

A backend service that reads audit events from Story DynamoDB, broadcasts them onchain via `registerBatch` transactions, and persists transaction status in a broadcaster-owned Postgres database.

## Claims

- The Data Broadcaster reads audit events from Story DynamoDB trace items (META#<seq>) and broadcasts them onchain. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-1) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_78bba44d65215ca857e0bd4af4caa5dc` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-1` `source_timestamp=2026-05-27T22:32:00Z`
- The service uses a locked throughput model with durable broadcaster jobs and immutable submission batches. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-1) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_78bba44d65215ca857e0bd4af4caa5dc` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-1` `source_timestamp=2026-05-27T22:32:00Z`
- The Broadcasters components include Stream Consumer, Bootstrap Scanner, Submitter, and Confirmer. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-1) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_78bba44d65215ca857e0bd4af4caa5dc` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-1` `source_timestamp=2026-05-27T22:32:00Z`
- Materialization creates broadcaster_jobs from DynamoDB Stream INSERT records with SK starting with META#. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-2) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_d80542ac197a1a366f535f18da92c8e3` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-2` `source_timestamp=2026-05-27T22:32:00Z`
- If a record with the same event_id and identical fields is seen, materialization treats it as a replay and no-ops. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-2) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_d80542ac197a1a366f535f18da92c8e3` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-2` `source_timestamp=2026-05-27T22:32:00Z`
- The submitter batches jobs, encodes registerBatch(Registration[]) calldata, reserves wallets and nonces, and broadcasts transactions. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-3) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_d93ead98f38be7421b0fdecaa6dade2e` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-3` `source_timestamp=2026-05-27T22:32:00Z`
- The confirmer polls for receipt using eth_getTransactionReceipt and drives batches and jobs to terminal states (CONFIRMED, failed, gas_exhausted). `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-4) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_5349fc2fbd4bbc42cb10e1d2108f38a2` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-4` `source_timestamp=2026-05-27T22:32:00Z`
- V1 does not write onchain status back to DynamoDB; Broadcaster DB is the authoritative source of onchain lifecycle. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-1) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_78bba44d65215ca857e0bd4af4caa5dc` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-1` `source_timestamp=2026-05-27T22:32:00Z`
- The broadcaster guardrails include per-wallet in-flight cap, attempt limits with backoff, and gas limit multipliers. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-7) `source_document_id=srcdoc_49ee8c02d32f092f540db43787bf8841` `source_revision_id=srcrev_41419f35f81d8994463b51129d16e317` `chunk_id=srcchunk_138572ca04e352613677daaa30a1b370` `native_locator=https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc#chunk-7` `source_timestamp=2026-05-27T22:32:00Z`

## Sources

- `source_document_id`: `srcdoc_49ee8c02d32f092f540db43787bf8841`
- `source_revision_id`: `srcrev_41419f35f81d8994463b51129d16e317`
- `source_url`: [Notion source](https://www.notion.so/Data-Broadcaster-Design-36d051299a54809d8deee22c8388a2dc)
