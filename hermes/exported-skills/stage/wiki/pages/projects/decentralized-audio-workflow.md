---
title: "Decentralized Audio Workflow"
type: "project"
slug: "projects/decentralized-audio-workflow"
freshness: "2026-01-13T15:25:00Z"
tags:
  - "audio-processing"
  - "poseidon"
  - "proteus-devnet"
  - "smart-contracts"
owners: []
source_revision_ids:
  - "srcrev_9251c7b8a29bcfdceda58e72617571e1"
conflict_state: "none"
---

# Decentralized Audio Workflow

## Summary

Design for a smart-contract-based audio processing workflow to migrate the Poseidon audio processing pipeline to the Proteus devnet.

## Claims

- The document defines requirements for a smart-contract-based audio processing workflow to support migration of the Poseidon audio processing pipeline to the Proteus devnet. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- The centralized workflow consists of two activities: handle_speech_quality and report_speech_quality. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- handle_speech_quality involves preprocessing (downloading and transcoding), validating (retrieving file script and user's seed phrase, generating quality score), and uploading media and score to R2 if validation succeeds. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- handle_speech_quality is enqueued in the gpu_task_queue. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- report_speech_quality involves uploading the score to the DePin app DB and is enqueued in the cpu_task_queue. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- IP registration is handled via a separate cron job, not in the workflow. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- The distinction between gpu_task_queue and cpu_task_queue is not yet significant because validation uses the OpenAI API and does not require a GPU. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- Retry policy: initial_interval 1 second, backoff_coefficient 2.0, maximum_interval 100 seconds, maximum_attempts 3, non_retryable_error_types ["FileNotFoundError", "InvalidFileError"]. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- handle_speech_quality has a 10-minute start-to-close timeout; report_speech_quality has a 1-minute start-to-close timeout. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_991ebd4b8283f6af25d14246e3194c18` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-1` `source_timestamp=2026-01-13T15:25:00Z`
- In the smart-contract based workflow, IP registration is an activity; if it fails, the workflow fails. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-2) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_db948ee79df7ef5b6f83701f9d573054` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-2` `source_timestamp=2026-01-13T15:25:00Z`
- After IP registration, a first validation is scheduled; if it fails, a second validation is scheduled. If the second validation fails, the workflow fails; otherwise, the workflow succeeds. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-2) `source_document_id=srcdoc_dfd5e8466525e8445819c776c6457f75` `source_revision_id=srcrev_9251c7b8a29bcfdceda58e72617571e1` `chunk_id=srcchunk_db948ee79df7ef5b6f83701f9d573054` `native_locator=https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd#chunk-2` `source_timestamp=2026-01-13T15:25:00Z`

## Open Questions

- How does the two-validation approach in the smart-contract workflow relate to the single validation in the centralized workflow?
- The smart-contract based workflow design section is incomplete; full activity definitions and processing details are missing.
- What are the specific smart-contract interactions and on-chain components?

## Related Pages

- `poseidon-audio-processing-pipeline`
- `proteus-devnet`

## Sources

- `source_document_id`: `srcdoc_dfd5e8466525e8445819c776c6457f75`
- `source_revision_id`: `srcrev_9251c7b8a29bcfdceda58e72617571e1`
- `source_url`: [Notion source](https://www.notion.so/Decentralized-Audio-Workflow-Design-Document-272051299a548039a47ae9c6e2cb85fd)
