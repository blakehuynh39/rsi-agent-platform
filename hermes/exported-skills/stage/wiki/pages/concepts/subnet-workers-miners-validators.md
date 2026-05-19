---
title: "Subnet Workers (Miners \u0026 Validators)"
type: "concept"
slug: "concepts/subnet-workers-miners-validators"
freshness: "2025-08-11T19:46:00Z"
tags:
  - "miners"
  - "processing"
  - "subnet"
  - "validation"
  - "validators"
  - "workers"
  - "workflow"
owners: []
source_revision_ids:
  - "srcrev_7d58f76331d9e024a803dd1ccd83ca8a"
conflict_state: "none"
---

# Subnet Workers (Miners & Validators)

## Summary

A subnet worker node can register as either a data miner or a data validator. Miners handle processing jobs (filtering and annotation), while validators handle validation jobs. The two roles are kept separate to prevent self-validation, limit validator count, and allow different staking criteria.

## Claims

- When registering on the subnet as a worker, the node can either become a data miner or a data validator. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1) `source_document_id=srcdoc_d0294ac78e860a8d1c7ffc45158056bf` `source_revision_id=srcrev_7d58f76331d9e024a803dd1ccd83ca8a` `chunk_id=srcchunk_f2c098a240d83068e8d6cc996d1da3ee` `native_locator=https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1` `source_timestamp=2025-08-11T19:46:00Z`
- A data miner does processing jobs, and a data validator does validation jobs. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1) `source_document_id=srcdoc_d0294ac78e860a8d1c7ffc45158056bf` `source_revision_id=srcrev_7d58f76331d9e024a803dd1ccd83ca8a` `chunk_id=srcchunk_f2c098a240d83068e8d6cc996d1da3ee` `native_locator=https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1` `source_timestamp=2025-08-11T19:46:00Z`
- Miners and validators are kept separate for three reasons: to prevent miners from validating their own work, to limit the number of data validators for more accountability, and to allow different criteria for becoming a validator (e.g., more $POS staking and longer bonding period). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1) `source_document_id=srcdoc_d0294ac78e860a8d1c7ffc45158056bf` `source_revision_id=srcrev_7d58f76331d9e024a803dd1ccd83ca8a` `chunk_id=srcchunk_f2c098a240d83068e8d6cc996d1da3ee` `native_locator=https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1` `source_timestamp=2025-08-11T19:46:00Z`
- Task table shows: Filter (subnet of processing) is performed by a miner and includes CPU-only steps (metadata & FPS check, video length, etc.) and with-GPU steps (CPU + GPU-accelerated ffmpeg). Processing is performed by a miner and includes CPU-only steps (sample audio/video frames, upload annotation) and with-GPU steps (CPU + annotate using OSS model or OpenAPI with API cost). Validation is performed by a validator and includes CPU-only steps (check basic validity of annotation per YAML spec) and with-GPU steps (CPU + re-run annotation to generate and compare annotations). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1) `source_document_id=srcdoc_d0294ac78e860a8d1c7ffc45158056bf` `source_revision_id=srcrev_7d58f76331d9e024a803dd1ccd83ca8a` `chunk_id=srcchunk_f2c098a240d83068e8d6cc996d1da3ee` `native_locator=https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1` `source_timestamp=2025-08-11T19:46:00Z`
- A processing job is completed by one data miner, and a validation job is duplicated and completed by many data validators (the exact number referenced in a linked page). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1) `source_document_id=srcdoc_d0294ac78e860a8d1c7ffc45158056bf` `source_revision_id=srcrev_7d58f76331d9e024a803dd1ccd83ca8a` `chunk_id=srcchunk_f2c098a240d83068e8d6cc996d1da3ee` `native_locator=https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1` `source_timestamp=2025-08-11T19:46:00Z`
- The workflow YAML version video/1.0.0 defines a Full Video Processing Pipeline with activities: 'preprocess video' (id: video_preprocess, worker: processing, require_gpu: true); 'validate video' (id: video_validate, worker: validation, require_gpu: false, uses GPT-4 Vision or equivalent to compare annotations and produce confidence_score); and 'human validate video' (id: video_human_validation, worker: validation, require_gpu: false, evaluates annotations and uploads metrics). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1) `source_document_id=srcdoc_d0294ac78e860a8d1c7ffc45158056bf` `source_revision_id=srcrev_7d58f76331d9e024a803dd1ccd83ca8a` `chunk_id=srcchunk_f2c098a240d83068e8d6cc996d1da3ee` `native_locator=https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-1` `source_timestamp=2025-08-11T19:46:00Z`
  - citation: [Notion source](https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-2) `source_document_id=srcdoc_d0294ac78e860a8d1c7ffc45158056bf` `source_revision_id=srcrev_7d58f76331d9e024a803dd1ccd83ca8a` `chunk_id=srcchunk_f6f14be0c165c44066f860e170261153` `native_locator=https://www.notion.so/Subnet-Workers-Miners-Validators-23f051299a5480ca83badf1822ef9165#chunk-2` `source_timestamp=2025-08-11T19:46:00Z`

## Sources

- `source_document_id`: `srcdoc_67ae03c2c6c9aa6eaa1bd5c0f7dd8f3f`
- `source_revision_id`: `srcrev_7fe38cf94e5767c134f47244bcf6f60b`
- `source_url`: [Notion source](https://www.notion.so/Subnet-358051299a5480f49fc2de7420dad342)
