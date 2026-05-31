---
title: "Web2 Infra Threat Model"
type: "system"
slug: "systems/web2-infra-threat-model"
freshness: "2025-04-24T20:13:00Z"
tags:
  - "infrastructure"
  - "security"
  - "stride"
  - "threat-modeling"
owners: []
source_revision_ids:
  - "srcrev_ea737c4e1dbd0fa12fd03d945c0f631e"
conflict_state: "none"
---

# Web2 Infra Threat Model

## Summary

Threat model for the Web2 infrastructure covering Kafka-based streaming pipeline and AI/ML data services (Iceberg, Milvus, Spark). Uses STRIDE categories to identify spoofing, tampering, repudiation, information disclosure, denial of service, and elevation of privilege threats, with detailed mitigations for each.

## Claims

- The threat model depends heavily on which infrastructure is chosen. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-1) `source_document_id=srcdoc_8d323e78677ddc0edcb502a171920adf` `source_revision_id=srcrev_ea737c4e1dbd0fa12fd03d945c0f631e` `chunk_id=srcchunk_02c171d1e1f24a958ee38d6bc373e7d0` `native_locator=https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-1` `source_timestamp=2025-04-24T20:13:00Z`
- For the Kafka pipeline, spoofing threats include unauthorized actors impersonating producers, services, or consumers. Mitigations include mutual TLS (mTLS) for all Kafka brokers, producers, and consumers; OAuth2, SASL/SCRAM, or mTLS authentication for Kafka client identity validation; JWT validation on Auth Gateway using trusted identity providers; and requiring producer identity embedding in messages with signed claims. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-1) `source_document_id=srcdoc_8d323e78677ddc0edcb502a171920adf` `source_revision_id=srcrev_ea737c4e1dbd0fa12fd03d945c0f631e` `chunk_id=srcchunk_02c171d1e1f24a958ee38d6bc373e7d0` `native_locator=https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-1` `source_timestamp=2025-04-24T20:13:00Z`
- For the Kafka pipeline, tampering threats include payloads modified in transit or within brokers, corrupting stream data. Mitigations include enforcing encryption in transit using TLS on Kafka + Proxy + Gateway; signing messages at the application level (e.g., with HMAC or ECDSA) and verifying at consumer level; using immutable, append-only topics with validation at each stage; and applying schema enforcement via Confluent Schema Registry to reject malformed data. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-1) `source_document_id=srcdoc_8d323e78677ddc0edcb502a171920adf` `source_revision_id=srcrev_ea737c4e1dbd0fa12fd03d945c0f631e` `chunk_id=srcchunk_02c171d1e1f24a958ee38d6bc373e7d0` `native_locator=https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-1` `source_timestamp=2025-04-24T20:13:00Z`
- For the Kafka pipeline, repudiation threats include no traceability of who produced/consumed messages or executed workflows. Mitigations include enabling Kafka audit logs for produce/consume actions. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-1) `source_document_id=srcdoc_8d323e78677ddc0edcb502a171920adf` `source_revision_id=srcrev_ea737c4e1dbd0fa12fd03d945c0f631e` `chunk_id=srcchunk_02c171d1e1f24a958ee38d6bc373e7d0` `native_locator=https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-1` `source_timestamp=2025-04-24T20:13:00Z`
- For the AI/ML data services (Iceberg, Milvus, Spark), spoofing threats include unauthorized services impersonating trusted components such as fake vector data injection or unverified Spark job triggers. Mitigations include using mutual TLS (mTLS) across all services; requiring JWT-based or OAuth2 authentication for API access; implementing client certificates validated against a known CA; and setting up an identity-aware proxy in front of Milvus/Spark endpoints. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-2) `source_document_id=srcdoc_8d323e78677ddc0edcb502a171920adf` `source_revision_id=srcrev_ea737c4e1dbd0fa12fd03d945c0f631e` `chunk_id=srcchunk_38a350fe474b9aa9b039064cb0ccdc8c` `native_locator=https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-2` `source_timestamp=2025-04-24T20:13:00Z`
- For the AI/ML data services, tampering threats include metadata or vector records being altered maliciously such as vector corruption or unauthorized metadata edits. Mitigations include enforcing immutable metadata snapshots using Iceberg's versioning; using Apache Iceberg commit validation hooks to block unauthorized writes; signing and hashing data before ingest and verifying before vectorization; and limiting Spark jobs to read-only execution roles unless explicitly needed. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-2) `source_document_id=srcdoc_8d323e78677ddc0edcb502a171920adf` `source_revision_id=srcrev_ea737c4e1dbd0fa12fd03d945c0f631e` `chunk_id=srcchunk_38a350fe474b9aa9b039064cb0ccdc8c` `native_locator=https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-2` `source_timestamp=2025-04-24T20:13:00Z`
- For the AI/ML data services, repudiation threats include users denying changes made to datasets or vector queries due to lack of audit trails. Mitigations include enabling detailed audit logging for Iceberg table mutations (via Hive Metastore or Iceberg catalog logging); configuring Milvus to log query origin, timestamp, and result summary; and integrating Spark event logging and job lineage tracking using Spark History Server or Apache Atlas. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-2) `source_document_id=srcdoc_8d323e78677ddc0edcb502a171920adf` `source_revision_id=srcrev_ea737c4e1dbd0fa12fd03d945c0f631e` `chunk_id=srcchunk_38a350fe474b9aa9b039064cb0ccdc8c` `native_locator=https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf#chunk-2` `source_timestamp=2025-04-24T20:13:00Z`

## Open Questions

- What are the remaining STRIDE categories (Information Disclosure, Denial of Service, Elevation of Privilege) and their mitigations for both the Kafka pipeline and AI/ML services?
- Which specific infrastructure components will be selected?

## Sources

- `source_document_id`: `srcdoc_8d323e78677ddc0edcb502a171920adf`
- `source_revision_id`: `srcrev_ea737c4e1dbd0fa12fd03d945c0f631e`
- `source_url`: [Notion source](https://www.notion.so/Web2-Infra-threat-modeling-1df051299a548049b423c818733acfdf)
