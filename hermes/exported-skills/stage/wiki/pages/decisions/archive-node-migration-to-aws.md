---
title: "Archive Node Migration to AWS"
type: "decision"
slug: "decisions/archive-node-migration-to-aws"
freshness: "2026-01-28T02:01:31Z"
tags:
  - "archive-node"
  - "aws"
  - "gcp"
  - "infrastructure"
  - "migration"
owners: []
source_revision_ids:
  - "srcrev_069d1d98b9583041c100720ae1c3ad4e"
  - "srcrev_23d69a418ba088f83409a62660832c85"
  - "srcrev_2b3824164e3659ecced604e47cf1eed3"
  - "srcrev_3c95e7d3c20cc8547b7bc496e31b3b86"
  - "srcrev_6004e56f5d083a8ff6a1124f3185dde0"
  - "srcrev_69b4a23486c05bf6f222cdfd947055c0"
  - "srcrev_8872d38cf73b0c69424977724763295f"
  - "srcrev_89d68928cd0fd070507d91b5616f5c92"
  - "srcrev_9eb06240fef156c6161515e0acedff34"
  - "srcrev_b14b87d7915d702b72e8a377c2ad87ba"
  - "srcrev_eec4e851479bdab1dae471e9a3cd9c46"
conflict_state: "none"
---

# Archive Node Migration to AWS

## Summary

Migrating Story Protocol's archive nodes from GCP to AWS, deprecating direct IP access in favor of domain names.

## Claims

- A request was made to check if any Poseidon side apps are using Story Archive node IP directly, because servers in GCP are planned for removal. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_23d69a418ba088f83409a62660832c85` `chunk_id=srcchunk_813dbcaf6fa29529c7a97ab2221e6527` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560386.047159` `source_timestamp=2026-01-28T00:33:06Z`
- IP registration is still using archive node by design. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_3c95e7d3c20cc8547b7bc496e31b3b86` `chunk_id=srcchunk_87930fff0fa56c263b351f5c66effa56` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560495.293529` `source_timestamp=2026-01-28T00:34:55Z`
- Suggestion to change from IP to domain name. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_6004e56f5d083a8ff6a1124f3185dde0` `chunk_id=srcchunk_3d9d3f558ddc433e5736655c3fe5b594` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560545.583849` `source_timestamp=2026-01-28T00:35:45Z`
- Migration to AWS is in progress, and <@U080YAW205V> is the person to contact. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_8872d38cf73b0c69424977724763295f` `chunk_id=srcchunk_3f5588523759603df0699a914d8b470b` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560665.172459` `source_timestamp=2026-01-28T00:37:45Z`
- The archive nodes are being migrated to AWS; the previous GCP archive node will be deprecated. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
- If apps are already using a domain name, they should be unaffected by the migration. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_69b4a23486c05bf6f222cdfd947055c0` `chunk_id=srcchunk_f2846794e664b7dd850c03dbcc65ce74` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565307.265889` `source_timestamp=2026-01-28T01:55:07Z`
- Confirmation to search for direct usage of IP 34.139.142.168. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_9eb06240fef156c6161515e0acedff34` `chunk_id=srcchunk_eda3af3429788adea6b36a8258584fd5` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565445.635889` `source_timestamp=2026-01-28T01:57:25Z`
- In depin-api, staging uses https://internal-full.aeneid.storyrpc.io/ and prod uses https://internal-full.storyrpc.io. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_b14b87d7915d702b72e8a377c2ad87ba` `chunk_id=srcchunk_c96c6d81cef6b86a7866f4a901253a0e` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565546.498389` `source_timestamp=2026-01-28T01:59:06Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- Bank vault is not used; secrets have been migrated to AWS Secrets Manager. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_069d1d98b9583041c100720ae1c3ad4e` `chunk_id=srcchunk_3bb1fb817821b349afd3ab642dd5d770` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565604.067249` `source_timestamp=2026-01-28T02:00:24Z`
- An AWS secret check found no usage of IP 34.139.142.168. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2b3824164e3659ecced604e47cf1eed3` `chunk_id=srcchunk_2a5d1c61e56dca62789b9846e2357229` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565691.322389` `source_timestamp=2026-01-28T02:01:31Z`

## Open Questions

- Are all apps fully migrated to domain name references?
- Is the GCP archive node fully deprecated?

## Sources

- `source_document_id`: `srcdoc_83db2545a88ec9b2fc8fa27e705a93e3`
- `source_revision_id`: `srcrev_6004e56f5d083a8ff6a1124f3185dde0`
