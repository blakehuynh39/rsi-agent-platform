---
title: "Documentation Portal Structure"
type: "concept"
slug: "concepts/documentation-portal-structure"
freshness: "2024-09-11T03:58:00Z"
tags:
  - "documentation"
  - "explanation"
  - "how-to"
  - "reference"
  - "structure"
  - "tutorials"
owners: []
source_revision_ids:
  - "srcrev_1f0b482596763ffe094dd66aae64a25c"
conflict_state: "none"
---

# Documentation Portal Structure

## Summary

Defines the overall framework for the documentation portal, outlining the goals and minimum requirements for Tutorials, How-to Guides, Explanations, and Reference sections, with a priority on tutorials for EthDenver.

## Claims

- The documentation portal should follow a framework divided into Tutorials, How-to Guides, Explanations, and Reference sections. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-1) `source_document_id=srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4` `source_revision_id=srcrev_1f0b482596763ffe094dd66aae64a25c` `chunk_id=srcchunk_595ec4e3092f5b4b3e4b8dce22d8a6c3` `native_locator=https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-1` `source_timestamp=2024-09-11T03:58:00Z`
- Tutorials should be learning-oriented and educational, spending more time explaining things than how-to guides, and are the priority for EthDenver. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-1) `source_document_id=srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4` `source_revision_id=srcrev_1f0b482596763ffe094dd66aae64a25c` `chunk_id=srcchunk_595ec4e3092f5b4b3e4b8dce22d8a6c3` `native_locator=https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-1` `source_timestamp=2024-09-11T03:58:00Z`
- The goal of tutorials is to help developers grasp core protocol concepts practically, with copy-pasteable and forkable code acting as a quickstart. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-1) `source_document_id=srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4` `source_revision_id=srcrev_1f0b482596763ffe094dd66aae64a25c` `chunk_id=srcchunk_595ec4e3092f5b4b3e4b8dce22d8a6c3` `native_locator=https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-1` `source_timestamp=2024-09-11T03:58:00Z`
- A minimum required tutorial is a simple end-to-end SDK tutorial showcasing basic piping for building a dapp on SP, potentially as a Next.js app with a simplified admin panel and a glossary of actions. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-1) `source_document_id=srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4` `source_revision_id=srcrev_1f0b482596763ffe094dd66aae64a25c` `chunk_id=srcchunk_595ec4e3092f5b4b3e4b8dce22d8a6c3` `native_locator=https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-1` `source_timestamp=2024-09-11T03:58:00Z`
- Explanations should go beyond protocol explanations to outline how SP fits in the business domain, including problems it solves and how companies can use it, with case studies interlinked. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-2) `source_document_id=srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4` `source_revision_id=srcrev_1f0b482596763ffe094dd66aae64a25c` `chunk_id=srcchunk_5a53d97f6a5276d437279ecff0a6cfc7` `native_locator=https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-2` `source_timestamp=2024-09-11T03:58:00Z`
- Reference documentation should cover every SDK function, data type, and contract function/argument, with simple usage examples, similar to Postgres documentation. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-2) `source_document_id=srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4` `source_revision_id=srcrev_1f0b482596763ffe094dd66aae64a25c` `chunk_id=srcchunk_5a53d97f6a5276d437279ecff0a6cfc7` `native_locator=https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-2` `source_timestamp=2024-09-11T03:58:00Z`
- Good documentation is defined by ease of finding exactly what is needed, not by visual prettiness, and should document every common question or issue. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-2) `source_document_id=srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4` `source_revision_id=srcrev_1f0b482596763ffe094dd66aae64a25c` `chunk_id=srcchunk_5a53d97f6a5276d437279ecff0a6cfc7` `native_locator=https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-2` `source_timestamp=2024-09-11T03:58:00Z`
- Documentation should be a living document with dedicated time every sprint for enhancements, and every PR change should be used as an opportunity to review and improve documentation. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-2) `source_document_id=srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4` `source_revision_id=srcrev_1f0b482596763ffe094dd66aae64a25c` `chunk_id=srcchunk_5a53d97f6a5276d437279ecff0a6cfc7` `native_locator=https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca#chunk-2` `source_timestamp=2024-09-11T03:58:00Z`

## Sources

- `source_document_id`: `srcdoc_1caa72f6ee9a6a3e9321a3db3e85fbf4`
- `source_revision_id`: `srcrev_1f0b482596763ffe094dd66aae64a25c`
- `source_url`: [Notion source](https://www.notion.so/Documentation-Structure-759428204de9469d9bea22daf1d636ca)
