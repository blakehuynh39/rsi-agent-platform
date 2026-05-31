---
title: "Poseidon Subnet Management API"
type: "project"
slug: "projects/poseidon-subnet-management-api"
freshness: "2025-10-22T21:22:00Z"
tags:
  - "api"
  - "go"
  - "poseidon"
  - "subnet-management"
owners: []
source_revision_ids:
  - "srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1"
conflict_state: "none"
---

# Poseidon Subnet Management API

## Summary

Design and implementation considerations for the Poseidon subnet management API, covering environment configuration, secrets management, database setup, and developer experience enhancements.

## Claims

- The project uses explicit .env variants for each lifecycle stage: .env.local, .env.staging, and .env.prod. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- A centralized config loader in internal/config/config.go merges .env files, Vault secrets, and environment variables. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- The Config struct includes fields for Env, Port, DBConn, JWTSecret, VaultAddress, SubnetNetwork, and ChainRPCURL. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Secrets are loaded from .env for local development and from Vault for staging/production. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Local Postgres is provisioned via docker-compose with image postgres:16, database poseidon, user postgres, password postgres, and port 5432. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Cloud database access uses port-forwarding via RDS-proxy; see GCP DB access for instructions. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Swagger documentation is currently running locally and needs deployment. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- A make dev target should spin up the API and database via Docker Compose. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Pre-commit hooks should run go fmt, go vet, and staticcheck. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Prometheus metrics should be integrated on /metrics. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Structured logging should use zerolog. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Version information should be embedded into the binary using -X main.Version=$(git describe --tags). `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- OpenAPI documentation should be added for /api/v1 endpoints using swag init. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- Kubernetes readiness endpoints /healthz, /readyz, and /livez should be implemented. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- The CI/CD pipeline uses GitHub Actions. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- The database setup may need to read onchain data and must separate devnet, testnet, and mainnet configurations. `claim:claim_1_16` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`
- API key management is a required consideration. `claim:claim_1_17` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e) `source_document_id=srcdoc_1d8c04fd4cf68c002c24eb61180db287` `source_revision_id=srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1` `chunk_id=srcchunk_614c77da79abbf6410490f5fb787783f` `native_locator=https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e` `source_timestamp=2025-10-22T21:22:00Z`

## Open Questions

- Observability strategy is undefined (marked as ????)

## Sources

- `source_document_id`: `srcdoc_1d8c04fd4cf68c002c24eb61180db287`
- `source_revision_id`: `srcrev_2bd0fd9ebb20fdf10c39a273eadb22f1`
- `source_url`: [Notion source](https://www.notion.so/Poseidon-subnet-management-API-repo-considerations-28c051299a548050b85aca55b4d47e9e)
