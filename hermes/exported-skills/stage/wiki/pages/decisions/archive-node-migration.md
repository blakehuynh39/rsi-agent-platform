---
title: "Archive Node Migration to AWS"
type: "decision"
slug: "decisions/archive-node-migration"
freshness: "2026-01-28T02:01:31Z"
tags:
  - "archive-node"
  - "aws"
  - "gcp"
  - "infrastructure"
  - "migration"
  - "story"
owners:
  - "U07TNT9N4JC"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_069d1d98b9583041c100720ae1c3ad4e"
  - "srcrev_2a16c3b7dde0ec71467b231d1d6aa496"
  - "srcrev_2b3824164e3659ecced604e47cf1eed3"
  - "srcrev_3c95e7d3c20cc8547b7bc496e31b3b86"
  - "srcrev_89d68928cd0fd070507d91b5616f5c92"
  - "srcrev_9eb06240fef156c6161515e0acedff34"
  - "srcrev_eec4e851479bdab1dae471e9a3cd9c46"
conflict_state: "none"
---

# Archive Node Migration to AWS

## Summary

Migration of Story archive nodes from GCP to AWS, deprecating the GCP server at IP 34.139.142.168, and ensuring dependent applications use domain names instead of hardcoded IPs.

## Claims

- Story plans to deprecate the GCP archive node and migrate to AWS. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
- The IP address 34.139.142.168 was previously used as an archive node; apps should avoid hardcoding it and use domain names instead. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_9eb06240fef156c6161515e0acedff34` `chunk_id=srcchunk_eda3af3429788adea6b36a8258584fd5` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565445.635889` `source_timestamp=2026-01-28T01:57:25Z`
- For staging, the archive node is accessed via domain internal-full.aeneid.storyrpc.io; for production, internal-full.storyrpc.io. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- IP registration application is believed to still use the archive node by design. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_3c95e7d3c20cc8547b7bc496e31b3b86` `chunk_id=srcchunk_87930fff0fa56c263b351f5c66effa56` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560495.293529` `source_timestamp=2026-01-28T00:34:55Z`
- Bank vault is not currently used; secrets have been migrated to AWS Secrets Manager. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_069d1d98b9583041c100720ae1c3ad4e` `chunk_id=srcchunk_3bb1fb817821b349afd3ab642dd5d770` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565604.067249` `source_timestamp=2026-01-28T02:00:24Z`
- A check of awe secrets found no usage of IP 34.139.142.168. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2b3824164e3659ecced604e47cf1eed3` `chunk_id=srcchunk_2a5d1c61e56dca62789b9846e2357229` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565691.322389` `source_timestamp=2026-01-28T02:01:31Z`
- If apps are already using domain names, they should be unaffected by the IP deprecation. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2a16c3b7dde0ec71467b231d1d6aa496` `chunk_id=srcchunk_62ac65792e88473ba6a924f6d52b21ba` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565577.271279` `source_timestamp=2026-01-28T01:59:37Z`

## Open Questions

- Are all dependent applications (e.g., IP registration, depin-api, etc.) using domain names or still referencing the IP?

## Sources

- `source_document_id`: `srcdoc_83db2545a88ec9b2fc8fa27e705a93e3`
- `source_revision_id`: `srcrev_6f048534389c4ddf7c3045edce4737dd`
