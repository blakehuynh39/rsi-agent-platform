---
title: "Story Archive Node: GCP Deprecation and AWS Migration"
type: "decision"
slug: "decisions/story-archive-node-gcp-deprecation-aws-migration"
freshness: "2026-01-28T02:01:31Z"
tags:
  - "archive-node"
  - "aws"
  - "domain-name"
  - "gcp"
  - "ip-address"
  - "migration"
  - "poseidon"
  - "story"
owners:
  - "U04KL1096F9"
  - "U04KTUN5WFQ"
  - "U07A7AUGL5V"
  - "U07FJRNBBEZ"
  - "U07TNT9N4JC"
  - "U080YAW205V"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_173502db147e7c749e821b57c872bcda"
  - "srcrev_2a16c3b7dde0ec71467b231d1d6aa496"
  - "srcrev_2b3824164e3659ecced604e47cf1eed3"
  - "srcrev_3c95e7d3c20cc8547b7bc496e31b3b86"
  - "srcrev_6004e56f5d083a8ff6a1124f3185dde0"
  - "srcrev_69b4a23486c05bf6f222cdfd947055c0"
  - "srcrev_8872d38cf73b0c69424977724763295f"
  - "srcrev_89d68928cd0fd070507d91b5616f5c92"
  - "srcrev_9eb06240fef156c6161515e0acedff34"
  - "srcrev_eec4e851479bdab1dae471e9a3cd9c46"
conflict_state: "none"
---

# Story Archive Node: GCP Deprecation and AWS Migration

## Summary

Decision to migrate Story archive nodes from GCP to AWS, deprecating the GCP archive node at IP 34.139.142.168. Poseidon apps must use domain names (staging: https://internal-full.aeneid.storyrpc.io/, production: https://internal-full.storyrpc.io) instead of hardcoded IPs. After checking, no hardcoded IP usage was found, so no immediate changes are needed.

## Claims

- The Story archive node in GCP (IP 34.139.142.168) is being deprecated and migrated to AWS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_9eb06240fef156c6161515e0acedff34` `chunk_id=srcchunk_eda3af3429788adea6b36a8258584fd5` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565445.635889` `source_timestamp=2026-01-28T01:57:25Z`
- Poseidon apps should use domain names to connect to archive nodes, not hardcoded IPs. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_6004e56f5d083a8ff6a1124f3185dde0` `chunk_id=srcchunk_3d9d3f558ddc433e5736655c3fe5b594` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560545.583849` `source_timestamp=2026-01-28T00:35:45Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_173502db147e7c749e821b57c872bcda` `chunk_id=srcchunk_709a4da024e4b9beee949901957802d1` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565263.872619` `source_timestamp=2026-01-28T01:54:23Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_69b4a23486c05bf6f222cdfd947055c0` `chunk_id=srcchunk_f2846794e664b7dd850c03dbcc65ce74` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565307.265889` `source_timestamp=2026-01-28T01:55:07Z`
- The domain names for archive node access are: staging: https://internal-full.aeneid.storyrpc.io/, production: https://internal-full.storyrpc.io. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- A check of Poseidon AWS secrets and app configurations found no direct usage of the IP 34.139.142.168, and domain names are already in use. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2b3824164e3659ecced604e47cf1eed3` `chunk_id=srcchunk_2a5d1c61e56dca62789b9846e2357229` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565691.322389` `source_timestamp=2026-01-28T02:01:31Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2a16c3b7dde0ec71467b231d1d6aa496` `chunk_id=srcchunk_62ac65792e88473ba6a924f6d52b21ba` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565577.271279` `source_timestamp=2026-01-28T01:59:37Z`
- The IP registration app was previously using the archive node directly, but migration to domain names is in progress. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_3c95e7d3c20cc8547b7bc496e31b3b86` `chunk_id=srcchunk_87930fff0fa56c263b351f5c66effa56` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560495.293529` `source_timestamp=2026-01-28T00:34:55Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_8872d38cf73b0c69424977724763295f` `chunk_id=srcchunk_3f5588523759603df0699a914d8b470b` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560665.172459` `source_timestamp=2026-01-28T00:37:45Z`

## Open Questions

- Are all Poseidon apps fully migrated to use domain names for archive node access?

## Sources

- `source_document_id`: `srcdoc_83db2545a88ec9b2fc8fa27e705a93e3`
- `source_revision_id`: `srcrev_5c4092d9452e21cf002ad64189d61cb0`
