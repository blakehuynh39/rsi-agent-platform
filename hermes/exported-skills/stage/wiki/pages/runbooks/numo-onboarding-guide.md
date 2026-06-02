---
title: "Numo Onboarding Guide"
type: "runbook"
slug: "runbooks/numo-onboarding-guide"
freshness: "2026-06-02T00:25:00Z"
tags:
  - "development"
  - "onboarding"
  - "setup"
owners:
  - "PIP Labs"
source_revision_ids:
  - "srcrev_f4f0737c3faeb96f40356202cf217726"
conflict_state: "none"
---

# Numo Onboarding Guide

## Summary

Step-by-step guide for new team members to set up the Numo development environment, including repository cloning, environment configuration, and key documentation.

## Claims

- Clone both repos into a shared parent folder: git clone https://github.com/piplabs/numo-monorepo.git and git clone https://github.com/piplabs/depin-backend.git. `claim:claim_4_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4e5ac035a044e71cb3f79255559a5437` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1` `source_timestamp=2026-06-02T00:25:00Z`
- Install dependencies with pnpm install; Node ≥ 23 and pnpm 10 required. `claim:claim_4_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4e5ac035a044e71cb3f79255559a5437` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1` `source_timestamp=2026-06-02T00:25:00Z`
- Set up environment variables by copying .env.example files: cp apps/web/.env.example apps/web/.env.local, and similarly for admin, landing, and backend. `claim:claim_4_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4e5ac035a044e71cb3f79255559a5437` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1` `source_timestamp=2026-06-02T00:25:00Z`
- Start dev servers: pnpm dev runs all apps simultaneously; individual commands: pnpm dev:web (:3000), pnpm dev:admin (:3002), pnpm dev:landing (:3001). `claim:claim_4_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4e5ac035a044e71cb3f79255559a5437` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-1` `source_timestamp=2026-06-02T00:25:00Z`
- Key documentation includes: numo-monorepo/docs/README.md (catalog), references/docs/numo-prd.md (PRD), depin-backend/AGENTS.md (agent instructions), depin-backend/apps/api/config/base.toml (base config). `claim:claim_4_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-5) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4be16a23258eba9342fd4d7ccc5ef03d` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-5` `source_timestamp=2026-06-02T00:25:00Z`

## Related Pages

- `numo-architecture`
- `numo-platform-overview`

## Sources

- `source_document_id`: `srcdoc_5b2f0266ee274de6a7ca742845dfbcd4`
- `source_revision_id`: `srcrev_f4f0737c3faeb96f40356202cf217726`
- `source_url`: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3)
