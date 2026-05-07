---
title: "Subnet 2.0 Validation Scheme"
type: "decision"
slug: "decisions/subnet-2-0-validation-scheme"
freshness: "2025-12-09T00:52:00Z"
tags:
  - "auditor"
  - "optimistic-validation"
  - "subnet-2-0"
  - "validation"
owners: []
source_revision_ids:
  - "srcrev_7c80a282f63331ad550359da468dd210"
conflict_state: "none"
---

# Subnet 2.0 Validation Scheme

## Summary

Explores the validation scheme for Subnet 2.0, weighing optimistic validation against the need for a final auditor, likely the subnet owner, to ensure data quality.

## Claims

- Optimistic validation is hard to make a generalized ZK proof for because pipeline logic is arbitrary, especially if it involves running local models. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8) `source_document_id=srcdoc_7e8213264783d5e2024d11c6a19d5ee0` `source_revision_id=srcrev_7c80a282f63331ad550359da468dd210` `chunk_id=srcchunk_0a2af40e93f5ed649bce18ace454dae3` `native_locator=https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8` `source_timestamp=2025-12-09T00:52:00Z`
- Challengers in an optimistic validation scheme need access to the data to challenge, so they have to be whitelisted. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8) `source_document_id=srcdoc_7e8213264783d5e2024d11c6a19d5ee0` `source_revision_id=srcrev_7c80a282f63331ad550359da468dd210` `chunk_id=srcchunk_0a2af40e93f5ed649bce18ace454dae3` `native_locator=https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8` `source_timestamp=2025-12-09T00:52:00Z`
- The challenge time window may not satisfy a DePIN app's requirement, e.g., a 1-day SLA for validation results versus a 7-day challenge window. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8) `source_document_id=srcdoc_7e8213264783d5e2024d11c6a19d5ee0` `source_revision_id=srcrev_7c80a282f63331ad550359da468dd210` `chunk_id=srcchunk_0a2af40e93f5ed649bce18ace454dae3` `native_locator=https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8` `source_timestamp=2025-12-09T00:52:00Z`
- Validators have an incentive to use the cheapest models, which could lead to subpar deepfake models or blindly passing validations. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8) `source_document_id=srcdoc_7e8213264783d5e2024d11c6a19d5ee0` `source_revision_id=srcrev_7c80a282f63331ad550359da468dd210` `chunk_id=srcchunk_0a2af40e93f5ed649bce18ace454dae3` `native_locator=https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8` `source_timestamp=2025-12-09T00:52:00Z`
- The final auditor should be the subnet owner because they are ultimately responsible for data quality and sellability. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8) `source_document_id=srcdoc_7e8213264783d5e2024d11c6a19d5ee0` `source_revision_id=srcrev_7c80a282f63331ad550359da468dd210` `chunk_id=srcchunk_0a2af40e93f5ed649bce18ace454dae3` `native_locator=https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8` `source_timestamp=2025-12-09T00:52:00Z`
- Subnet owners must run a minimal standard validator to validate challenged data, but should not receive rewards for punishing workers. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8) `source_document_id=srcdoc_7e8213264783d5e2024d11c6a19d5ee0` `source_revision_id=srcrev_7c80a282f63331ad550359da468dd210` `chunk_id=srcchunk_0a2af40e93f5ed649bce18ace454dae3` `native_locator=https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8` `source_timestamp=2025-12-09T00:52:00Z`

## Open Questions

- Should we use optimistic validation?
- Who is the final auditor?

## Sources

- `source_document_id`: `srcdoc_7e8213264783d5e2024d11c6a19d5ee0`
- `source_revision_id`: `srcrev_7c80a282f63331ad550359da468dd210`
- `source_url`: [Notion source](https://www.notion.so/Subnet-2-0-Design-Thoughts-2c0051299a5480cb8cd3f1b1b87115c8)
