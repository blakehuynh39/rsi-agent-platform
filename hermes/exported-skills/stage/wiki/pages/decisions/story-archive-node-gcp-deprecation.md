---
title: "GCP Story Archive Node Deprecation"
type: "decision"
slug: "decisions/story-archive-node-gcp-deprecation"
freshness: "2026-01-28T02:01:31Z"
tags:
  - "archive-node"
  - "AWS"
  - "deprecation"
  - "migration"
  - "Poseidon"
owners:
  - "U07A7AUGL5V"
  - "U07FJRNBBEZ"
  - "U07TNT9N4JC"
  - "U080YAW205V"
  - "U09M2SPUTSL"
  - "U0A3GPWELDP"
source_revision_ids:
  - "srcrev_069d1d98b9583041c100720ae1c3ad4e"
  - "srcrev_173502db147e7c749e821b57c872bcda"
  - "srcrev_2a16c3b7dde0ec71467b231d1d6aa496"
  - "srcrev_2b3824164e3659ecced604e47cf1eed3"
  - "srcrev_3c95e7d3c20cc8547b7bc496e31b3b86"
  - "srcrev_6004e56f5d083a8ff6a1124f3185dde0"
  - "srcrev_89d68928cd0fd070507d91b5616f5c92"
  - "srcrev_9eb06240fef156c6161515e0acedff34"
  - "srcrev_b14b87d7915d702b72e8a377c2ad87ba"
  - "srcrev_eec4e851479bdab1dae471e9a3cd9c46"
conflict_state: "none"
---

# GCP Story Archive Node Deprecation

## Summary

Deprecation of the GCP-hosted Story archive node IP 34.139.142.168, migrating to AWS and ensuring services use domain names instead of IP.

## Claims

- The Story archive node at IP 34.139.142.168 is hosted in GCP and is being deprecated. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
- IP registration service likely still uses the archive node by design; need to verify and migrate to domain. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_3c95e7d3c20cc8547b7bc496e31b3b86` `chunk_id=srcchunk_87930fff0fa56c263b351f5c66effa56` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560495.293529` `source_timestamp=2026-01-28T00:34:55Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_6004e56f5d083a8ff6a1124f3185dde0` `chunk_id=srcchunk_3d9d3f558ddc433e5736655c3fe5b594` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560545.583849` `source_timestamp=2026-01-28T00:35:45Z`
- The depin-api service in staging and production already uses domain names: staging uses https://internal-full.aeneid.storyrpc.io/ and prod uses https://internal-full.storyrpc.io. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_b14b87d7915d702b72e8a377c2ad87ba` `chunk_id=srcchunk_c96c6d81cef6b86a7866f4a901253a0e` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565546.498389` `source_timestamp=2026-01-28T01:59:06Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- On the new infrastructure, domain names have been used since inception. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2a16c3b7dde0ec71467b231d1d6aa496` `chunk_id=srcchunk_62ac65792e88473ba6a924f6d52b21ba` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565577.271279` `source_timestamp=2026-01-28T01:59:37Z`
- Bank vault is no longer used; secrets have been migrated to AWS Secrets Manager. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_069d1d98b9583041c100720ae1c3ad4e` `chunk_id=srcchunk_3bb1fb817821b349afd3ab642dd5d770` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565604.067249` `source_timestamp=2026-01-28T02:00:24Z`
- A check of AWS secrets did not find any usage of IP 34.139.142.168. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2b3824164e3659ecced604e47cf1eed3` `chunk_id=srcchunk_2a5d1c61e56dca62789b9846e2357229` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565691.322389` `source_timestamp=2026-01-28T02:01:31Z`
- Continued usage of the IP directly may require changes; services should switch to domain names. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_173502db147e7c749e821b57c872bcda` `chunk_id=srcchunk_709a4da024e4b9beee949901957802d1` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565263.872619` `source_timestamp=2026-01-28T01:54:23Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_9eb06240fef156c6161515e0acedff34` `chunk_id=srcchunk_eda3af3429788adea6b36a8258584fd5` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565445.635889` `source_timestamp=2026-01-28T01:57:25Z`

## Open Questions

- What is the plan for IP registration service if it still uses the archive node directly?

## Sources

- `source_document_id`: `srcdoc_83db2545a88ec9b2fc8fa27e705a93e3`
- `source_revision_id`: `srcrev_ce2b757dea16b9bb58b73697346b0789`
