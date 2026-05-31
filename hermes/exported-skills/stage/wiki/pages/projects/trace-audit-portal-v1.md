---
title: "Trace V1: AI Data Audit Portal"
type: "project"
slug: "projects/trace-audit-portal-v1"
freshness: "2026-05-13T19:36:00Z"
tags:
  - "ai-data-provenance"
  - "audit"
  - "trace"
  - "v1"
  - "web3"
owners:
  - "Allen"
  - "Andrea Muttoni"
  - "Avi"
  - "Avneet, Julie"
  - "Blake"
  - "Jacob"
  - "Raul + Lion team"
  - "Romain"
  - "Susan"
source_revision_ids:
  - "srcrev_fb031702b3c20983279d8c892be7ca0f"
conflict_state: "none"
---

# Trace V1: AI Data Audit Portal

## Summary

Trace is the public audit layer for AI training data registered on the Protocol. It provides an immutable receipt registry for data contributions, enabling labs to verify data provenance before licensing, contributors to confirm consent terms, and regulators to audit AI data. V1 ships a central portal and whitelabel embeds by June 15, 2026.

## Claims

- Trace is the public audit layer for data registered on the Protocol. It generates an immutable receipt for every contribution, surfacing app-level, user-level, and asset-level views without hosting the underlying data. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_fb031702b3c20983279d8c892be7ca0f` `chunk_id=srcchunk_9b3f8d655282927b8825e10ee218a6f3` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-13T19:36:00Z`
- V1 mainnet rollout and rebrand launch is targeted for June 15, 2026. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_fb031702b3c20983279d8c892be7ca0f` `chunk_id=srcchunk_9b3f8d655282927b8825e10ee218a6f3` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-13T19:36:00Z`
- Trace uses SHA-256 with a multihash prefix for content addressing. The API form is `sha256:<64-hex-chars>`; the onchain binary form is multihash `0x1220<32-byte-digest>`. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-3) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_fb031702b3c20983279d8c892be7ca0f` `chunk_id=srcchunk_9d78f850ac054ab66fb1cb5b0b103f69` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-3` `source_timestamp=2026-05-13T19:36:00Z`
- Trace stores whether sensitive metadata signals exist (e.g., `exif: present`), but never the actual values. Public, auditable metadata like KYC status, TOS version, and payout status is exposed in full. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_fb031702b3c20983279d8c892be7ca0f` `chunk_id=srcchunk_fec87b8175f75a1270f9c727b637929b` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2` `source_timestamp=2026-05-13T19:36:00Z`
- V1 includes a central portal at `trace.thedatafoundation.ai` and a whitelabel embed for contributing apps under their own subdomains (e.g., `audit.kled.ai`). Both provide app-level, user-level, and asset-level views. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_fb031702b3c20983279d8c892be7ca0f` `chunk_id=srcchunk_fec87b8175f75a1270f9c727b637929b` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-2` `source_timestamp=2026-05-13T19:36:00Z`
- V1 does not include fraud detection (Poseidon workstream), consumer file lookups, a CLI tool, or rendering TOS/PP diffs. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1) `source_document_id=srcdoc_f34407b922c741d72df23780d7864b93` `source_revision_id=srcrev_fb031702b3c20983279d8c892be7ca0f` `chunk_id=srcchunk_9b3f8d655282927b8825e10ee218a6f3` `native_locator=https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8#chunk-1` `source_timestamp=2026-05-13T19:36:00Z`

## Open Questions

- Governance for the compliance frameworks enum (e.g., `eu-ai-act`) is TBD. Proposal: Avneet + Julie (legal) + Avi (data) decide.
- It is unclear whether the full consent signature should be stored onchain or just the signature hash. Flagged for legal review.
- The max retry ceiling before manual escalation is proposed at 24h, pending confirmation from Blake.
- The ~1 min gas batching window is a placeholder. Romain needs to size based on Kled volume and actual costs.

## Related Pages

- `kled-partner-integration`
- `system-onchain-registry`
- `trace-data-model`
- `trace-webhook-api`

## Sources

- `source_document_id`: `srcdoc_f34407b922c741d72df23780d7864b93`
- `source_revision_id`: `srcrev_fb031702b3c20983279d8c892be7ca0f`
- `source_url`: [Notion source](https://www.notion.so/Trace-AI-Data-Audit-Portal-358051299a54806eabbbdfa3ce6181d8)
