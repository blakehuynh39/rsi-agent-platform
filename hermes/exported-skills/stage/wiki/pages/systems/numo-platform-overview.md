---
title: "Numo Platform Overview"
type: "system"
slug: "systems/numo-platform-overview"
freshness: "2026-06-02T00:25:00Z"
tags:
  - "contributor"
  - "numo"
  - "platform"
owners:
  - "PIP Labs"
source_revision_ids:
  - "srcrev_f4f0737c3faeb96f40356202cf217726"
conflict_state: "none"
---

# Numo Platform Overview

## Summary

Numo is a contributor platform for collecting training data (audio, multimodal) with rewards. It consists of a monorepo with four apps and a separate backend.

## Claims

- Numo spans 4 apps in a monorepo: web SPA (apps/web), admin dashboard (apps/admin), React Native mobile (apps/react-native), landing page (apps/landing). `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4e5ac035a044e71cb3f79255559a5437` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1` `source_timestamp=2026-06-02T00:25:00Z`
- The backend is a Rust (Axum) + PostgreSQL 16 API in the depin-backend repo, deployed to Kubernetes namespace story. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4e5ac035a044e71cb3f79255559a5437` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1` `source_timestamp=2026-06-02T00:25:00Z`
- Repository URLs: monorepo https://github.com/piplabs/numo-monorepo, backend https://github.com/piplabs/depin-backend. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4e5ac035a044e71cb3f79255559a5437` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1` `source_timestamp=2026-06-02T00:25:00Z`

## Related Pages

- `numo-architecture`
- `numo-onboarding-guide`

## Sources

- `source_document_id`: `srcdoc_5b2f0266ee274de6a7ca742845dfbcd4`
- `source_revision_id`: `srcrev_f4f0737c3faeb96f40356202cf217726`
- `source_url`: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3)
