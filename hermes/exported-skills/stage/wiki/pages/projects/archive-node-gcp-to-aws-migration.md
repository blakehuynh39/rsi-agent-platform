---
title: "Archive Node GCP to AWS Migration"
type: "project"
slug: "projects/archive-node-gcp-to-aws-migration"
freshness: "2026-01-28T02:01:31Z"
tags:
  - "archive-node"
  - "aws"
  - "gcp"
  - "ip-migration"
  - "poseidon"
owners:
  - "U04KL1096F9"
  - "U04KTUN5WFQ"
  - "U07A7AUGL5V"
  - "U07FJRNBBEZ"
  - "U07TNT9N4JC"
  - "U080YAW205V"
  - "U09M2SPUTSL"
  - "U0A3GPWELDP"
source_revision_ids:
  - "srcrev_2a16c3b7dde0ec71467b231d1d6aa496"
  - "srcrev_2b3824164e3659ecced604e47cf1eed3"
  - "srcrev_5c4092d9452e21cf002ad64189d61cb0"
  - "srcrev_6004e56f5d083a8ff6a1124f3185dde0"
  - "srcrev_89d68928cd0fd070507d91b5616f5c92"
  - "srcrev_9eb06240fef156c6161515e0acedff34"
  - "srcrev_eec4e851479bdab1dae471e9a3cd9c46"
conflict_state: "none"
---

# Archive Node GCP to AWS Migration

## Summary

The Story team is migrating archive nodes from GCP to AWS, aiming to deprecate direct IP usage (specifically 34.139.142.168) in favor of domain names across Poseidon side apps.

## Claims

- The Story team is deprecating GCP archive nodes and migrating them to AWS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
- Poseidon side apps should switch from using direct IP 34.139.142.168 to domain names. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_6004e56f5d083a8ff6a1124f3185dde0` `chunk_id=srcchunk_3d9d3f558ddc433e5736655c3fe5b594` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560545.583849` `source_timestamp=2026-01-28T00:35:45Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_9eb06240fef156c6161515e0acedff34` `chunk_id=srcchunk_eda3af3429788adea6b36a8258584fd5` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565445.635889` `source_timestamp=2026-01-28T01:57:25Z`
- Some apps were already using domain names on the new infrastructure. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2a16c3b7dde0ec71467b231d1d6aa496` `chunk_id=srcchunk_62ac65792e88473ba6a924f6d52b21ba` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565577.271279` `source_timestamp=2026-01-28T01:59:37Z`
- Staging uses https://internal-full.aeneid.storyrpc.io/ and production uses https://internal-full.storyrpc.io. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- No direct usage of 34.139.142.168 was found in AWS secrets. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2b3824164e3659ecced604e47cf1eed3` `chunk_id=srcchunk_2a5d1c61e56dca62789b9846e2357229` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565691.322389` `source_timestamp=2026-01-28T02:01:31Z`
- The team may need IP whitelisting for access to the new AWS server. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_5c4092d9452e21cf002ad64189d61cb0` `chunk_id=srcchunk_74a030c936cb75386f4465497f7eaaee` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560890.600279` `source_timestamp=2026-01-28T00:41:30Z`

## Open Questions

- Are there any other hardcoded IPs?
- Have all Poseidon side apps been migrated to use domain names?
- What is the timeline for GCP deprecation?

## Sources

- `source_document_id`: `srcdoc_83db2545a88ec9b2fc8fa27e705a93e3`
- `source_revision_id`: `srcrev_b2a9c49498b4464ea592f4038e5c7021`
