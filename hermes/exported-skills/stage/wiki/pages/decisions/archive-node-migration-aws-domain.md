---
title: "Archive Node Migration to AWS and Domain Names"
type: "decision"
slug: "decisions/archive-node-migration-aws-domain"
freshness: "2026-01-28T02:01:31Z"
tags:
  - "archive-node"
  - "aws"
  - "domain-name"
  - "gcp-deprecation"
  - "migration"
  - "story-protocol"
owners:
  - "U04KTUN5WFQ"
  - "U07TNT9N4JC"
  - "U080YAW205V"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_173502db147e7c749e821b57c872bcda"
  - "srcrev_2b3824164e3659ecced604e47cf1eed3"
  - "srcrev_8872d38cf73b0c69424977724763295f"
  - "srcrev_89d68928cd0fd070507d91b5616f5c92"
  - "srcrev_9eb06240fef156c6161515e0acedff34"
  - "srcrev_b14b87d7915d702b72e8a377c2ad87ba"
  - "srcrev_eec4e851479bdab1dae471e9a3cd9c46"
conflict_state: "none"
---

# Archive Node Migration to AWS and Domain Names

## Summary

Migration of Story Protocol archive nodes from GCP to AWS, requiring all services to switch from direct IP (34.139.142.168) to domain names (internal-full.storyrpc.io for prod, internal-full.aeneid.storyrpc.io for staging).

## Claims

- The GCP-hosted archive node at IP 34.139.142.168 is being deprecated in favor of AWS-hosted nodes. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
- All services should use domain names instead of direct IPs for archive node access. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_173502db147e7c749e821b57c872bcda` `chunk_id=srcchunk_709a4da024e4b9beee949901957802d1` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565263.872619` `source_timestamp=2026-01-28T01:54:23Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_9eb06240fef156c6161515e0acedff34` `chunk_id=srcchunk_eda3af3429788adea6b36a8258584fd5` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565445.635889` `source_timestamp=2026-01-28T01:57:25Z`
- The production archive node domain is https://internal-full.storyrpc.io. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- The staging archive node domain is https://internal-full.aeneid.storyrpc.io/. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- U080YAW205V is actively working on the migration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_8872d38cf73b0c69424977724763295f` `chunk_id=srcchunk_3f5588523759603df0699a914d8b470b` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560665.172459` `source_timestamp=2026-01-28T00:37:45Z`
- A check of secrets did not find any usage of the old IP 34.139.142.168. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2b3824164e3659ecced604e47cf1eed3` `chunk_id=srcchunk_2a5d1c61e56dca62789b9846e2357229` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565691.322389` `source_timestamp=2026-01-28T02:01:31Z`
- The depin-api charts in the PSDN repository contain archive node endpoint configuration for staging and production. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_b14b87d7915d702b72e8a377c2ad87ba` `chunk_id=srcchunk_c96c6d81cef6b86a7866f4a901253a0e` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565546.498389` `source_timestamp=2026-01-28T01:59:06Z`

## Sources

- `source_document_id`: `srcdoc_83db2545a88ec9b2fc8fa27e705a93e3`
- `source_revision_id`: `srcrev_afa7c60246340ee791496822a2b01c0d`
