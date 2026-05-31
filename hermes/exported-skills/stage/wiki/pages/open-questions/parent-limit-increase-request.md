---
title: "Parent Limit Increase Request"
type: "open_question"
slug: "open-questions/parent-limit-increase-request"
freshness: "2025-04-03T20:14:00Z"
tags:
  - "feedback"
  - "protocol-design"
  - "UX"
owners: []
source_revision_ids:
  - "srcrev_df199e95cd72ac31d5bfee7c1c8d3ba9"
conflict_state: "none"
---

# Parent Limit Increase Request

## Summary

Mahojin requests increasing the limit on number of parents and ancestors, currently 2 parents and 14 ancestors, to at least 16 direct parents, to accommodate common use cases like using multiple AI models.

## Claims

- Currently there is a limit of 2 parents and 14 ancestors. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Feedback-from-Partners-34d9f06d973c4b19b686223f13144287) `source_document_id=srcdoc_6cc68314e0c9a8a9692cb3301ed17c5c` `source_revision_id=srcrev_df199e95cd72ac31d5bfee7c1c8d3ba9` `chunk_id=srcchunk_69b4b9d3a42322ba477aa667a470e269` `native_locator=https://www.notion.so/Protocol-Feedback-from-Partners-34d9f06d973c4b19b686223f13144287` `source_timestamp=2025-04-03T20:14:00Z`
- Mahojin expects at least 16 direct parents. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Feedback-from-Partners-34d9f06d973c4b19b686223f13144287) `source_document_id=srcdoc_6cc68314e0c9a8a9692cb3301ed17c5c` `source_revision_id=srcrev_df199e95cd72ac31d5bfee7c1c8d3ba9` `chunk_id=srcchunk_69b4b9d3a42322ba477aa667a470e269` `native_locator=https://www.notion.so/Protocol-Feedback-from-Partners-34d9f06d973c4b19b686223f13144287` `source_timestamp=2025-04-03T20:14:00Z`
- The limited parent/ancestor count makes UX unnatural, e.g., when using 4 AI models to create an image, users must exclude 2 models to fit the limit, and models with many ancestors become hard to register as parents. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Protocol-Feedback-from-Partners-34d9f06d973c4b19b686223f13144287) `source_document_id=srcdoc_6cc68314e0c9a8a9692cb3301ed17c5c` `source_revision_id=srcrev_df199e95cd72ac31d5bfee7c1c8d3ba9` `chunk_id=srcchunk_69b4b9d3a42322ba477aa667a470e269` `native_locator=https://www.notion.so/Protocol-Feedback-from-Partners-34d9f06d973c4b19b686223f13144287` `source_timestamp=2025-04-03T20:14:00Z`

## Open Questions

- What are the implications on gas costs or contract complexity?
- What is the technical feasibility of increasing parent/ancestor limits?

## Sources

- `source_document_id`: `srcdoc_6cc68314e0c9a8a9692cb3301ed17c5c`
- `source_revision_id`: `srcrev_df199e95cd72ac31d5bfee7c1c8d3ba9`
- `source_url`: [Notion source](https://www.notion.so/Protocol-Feedback-from-Partners-34d9f06d973c4b19b686223f13144287)
