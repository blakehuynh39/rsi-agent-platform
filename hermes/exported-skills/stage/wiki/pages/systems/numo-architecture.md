---
title: "Numo Architecture"
type: "system"
slug: "systems/numo-architecture"
freshness: "2026-06-02T00:25:00Z"
tags:
  - "architecture"
  - "numo"
  - "system-design"
owners:
  - "PIP Labs"
source_revision_ids:
  - "srcrev_f4f0737c3faeb96f40356202cf217726"
conflict_state: "none"
---

# Numo Architecture

## Summary

High-level architecture diagrams and component descriptions for the Numo platform, including system context, backend services, and deployment topology.

## Claims

- The system context diagram shows contributors (web, mobile, World App mini-app, internal admin) interacting with the Numo platform (landing, web SPA, RN mobile, admin), which uses a PostgreSQL 16 database, IP registration worker, and external services (Dynamic, World ID, Stripe, LinkedIn, R2, Turnstile, Castle, ElevenLabs, Beehiiv, OpenRouter, Firebase, Story Protocol). Observability includes Grafana, Sentry, and GA4. `claim:claim_2_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-5) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_4be16a23258eba9342fd4d7ccc5ef03d` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-5` `source_timestamp=2026-06-02T00:25:00Z`
- The backend service architecture is layered: entrypoints (Public, Admin, NDV APIs) → middleware (auth, Turnstile, Castle, Sentry, metrics) → services (campaigns, submissions, rewards, etc.) → integrations (Dynamic, World, Stripe, etc.) and data layer (PostgreSQL, R2, MinIO/S3). Background jobs include multiplier sweep, idempotency cleanup, safety refresh, and hot path cache. `claim:claim_2_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-6) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_b75c1f36ffbec972a5f4b71a86a8f1bb` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-6` `source_timestamp=2026-06-02T00:25:00Z`
- Deployment topology: Monorepo (GitHub) triggers GitHub Actions to build, push to ECR, and deploy via ArgoCD to Vercel (web, admin, landing) and Kubernetes (API, IP registration worker pods). Infrastructure is managed via Terraform (EKS, VPC, RDS, IAM) and secrets via AWS Secrets Manager. `claim:claim_2_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-7) `source_document_id=srcdoc_5b2f0266ee274de6a7ca742845dfbcd4` `source_revision_id=srcrev_f4f0737c3faeb96f40356202cf217726` `chunk_id=srcchunk_805517320a4a83352c5c160d1124394c` `native_locator=https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3#chunk-7` `source_timestamp=2026-06-02T00:25:00Z`

## Related Pages

- `numo-onboarding-guide`
- `numo-platform-overview`
- `numo-rewards-system`

## Sources

- `source_document_id`: `srcdoc_5b2f0266ee274de6a7ca742845dfbcd4`
- `source_revision_id`: `srcrev_f4f0737c3faeb96f40356202cf217726`
- `source_url`: [Notion source](https://www.notion.so/Numo-Jed-Onboarding-364051299a5480e89387ce3d9cc535d3)
