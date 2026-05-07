---
title: "Privacy-Preserving Inference Techniques"
type: "concept"
slug: "concepts/privacy-preserving-inference-techniques"
freshness: "2025-03-27T17:52:00Z"
tags:
  - "inference"
  - "machine-learning"
  - "privacy"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_7a1a240805a7ae66f47a2fc0ccb92d16"
conflict_state: "none"
---

# Privacy-Preserving Inference Techniques

## Summary

Overview of methods to protect input data and model weights during inference, including homomorphic encryption, split learning, federated learning, zero-knowledge proofs, consensus-based verification, and trusted execution environments.

## Claims

- Model inversion attacks allow adversaries to reconstruct original training data (e.g., images or text) from model outputs. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6) `source_document_id=srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0` `source_revision_id=srcrev_7a1a240805a7ae66f47a2fc0ccb92d16` `chunk_id=srcchunk_fefec5f9a7accffc9593cea8c2ac36d5` `native_locator=https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6` `source_timestamp=2025-03-27T17:52:00Z`
- Membership inference attacks determine if a specific person’s data was used to train the model, critical for GDPR compliance. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6) `source_document_id=srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0` `source_revision_id=srcrev_7a1a240805a7ae66f47a2fc0ccb92d16` `chunk_id=srcchunk_fefec5f9a7accffc9593cea8c2ac36d5` `native_locator=https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6` `source_timestamp=2025-03-27T17:52:00Z`
- Homomorphic encryption enables inference on encrypted data, but is very slow for large models. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6) `source_document_id=srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0` `source_revision_id=srcrev_7a1a240805a7ae66f47a2fc0ccb92d16` `chunk_id=srcchunk_fefec5f9a7accffc9593cea8c2ac36d5` `native_locator=https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6` `source_timestamp=2025-03-27T17:52:00Z`
- Split learning splits the model so that raw input data never leaves the user’s device; PriMed applied this to medical imaging. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6) `source_document_id=srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0` `source_revision_id=srcrev_7a1a240805a7ae66f47a2fc0ccb92d16` `chunk_id=srcchunk_fefec5f9a7accffc9593cea8c2ac36d5` `native_locator=https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6` `source_timestamp=2025-03-27T17:52:00Z`
- Zero-knowledge proofs (zkDPS) can verify that a model was executed correctly without revealing the model itself, adapted for LLMs. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6) `source_document_id=srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0` `source_revision_id=srcrev_7a1a240805a7ae66f47a2fc0ccb92d16` `chunk_id=srcchunk_fefec5f9a7accffc9593cea8c2ac36d5` `native_locator=https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6` `source_timestamp=2025-03-27T17:52:00Z`
- Trusted Execution Environments (TEEs) like Intel SGX run model and data in an isolated memory region, protecting them even if the OS is compromised. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6) `source_document_id=srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0` `source_revision_id=srcrev_7a1a240805a7ae66f47a2fc0ccb92d16` `chunk_id=srcchunk_fefec5f9a7accffc9593cea8c2ac36d5` `native_locator=https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6` `source_timestamp=2025-03-27T17:52:00Z`
- Consensus-Based Verification (CBV) is faster but less secure than zkDPS, suitable for less sensitive inference like product recommendations. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6) `source_document_id=srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0` `source_revision_id=srcrev_7a1a240805a7ae66f47a2fc0ccb92d16` `chunk_id=srcchunk_fefec5f9a7accffc9593cea8c2ac36d5` `native_locator=https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6` `source_timestamp=2025-03-27T17:52:00Z`
- Federated learning is used alongside split learning for inference, with examples like Google GBoard autocomplete. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6) `source_document_id=srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0` `source_revision_id=srcrev_7a1a240805a7ae66f47a2fc0ccb92d16` `chunk_id=srcchunk_fefec5f9a7accffc9593cea8c2ac36d5` `native_locator=https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6` `source_timestamp=2025-03-27T17:52:00Z`

## Open Questions

- What are the best practices from literature for privacy-preserving inference? (The source document was cut off before listing them.)

## Sources

- `source_document_id`: `srcdoc_3b1ebfd793e2b9ca8826bebe49b3dfe0`
- `source_revision_id`: `srcrev_7a1a240805a7ae66f47a2fc0ccb92d16`
- `source_url`: [Notion source](https://www.notion.so/Findings-1c3051299a548093b7aff8524d8864e6)
