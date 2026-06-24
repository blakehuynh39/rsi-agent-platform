---
title: "Archive Node Deprecation and Domain Migration"
type: "project"
slug: "projects/archive-node-deprecation"
freshness: "2026-01-28T02:01:31Z"
tags:
  - "archive-node"
  - "domain-name"
  - "migration"
  - "poseidon"
  - "story-protocol"
owners:
  - "U080YAW205V"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_069d1d98b9583041c100720ae1c3ad4e"
  - "srcrev_23d69a418ba088f83409a62660832c85"
  - "srcrev_2b3824164e3659ecced604e47cf1eed3"
  - "srcrev_3c95e7d3c20cc8547b7bc496e31b3b86"
  - "srcrev_5c4092d9452e21cf002ad64189d61cb0"
  - "srcrev_6004e56f5d083a8ff6a1124f3185dde0"
  - "srcrev_6f048534389c4ddf7c3045edce4737dd"
  - "srcrev_89d68928cd0fd070507d91b5616f5c92"
  - "srcrev_9eb06240fef156c6161515e0acedff34"
  - "srcrev_b14b87d7915d702b72e8a377c2ad87ba"
  - "srcrev_eec4e851479bdab1dae471e9a3cd9c46"
conflict_state: "none"
---

# Archive Node Deprecation and Domain Migration

## Summary

The Story Protocol infra team is migrating archive nodes from GCP to AWS and requiring all dependent services to use domain names instead of hardcoded IPs. Poseidon's depin-api already uses domains, but other services like IP registration may need updating. Team is deprecating the old GCP server and will remove it once all dependencies are resolved.

## Claims

- The team plans to remove the GCP archive node servers and is verifying that no app uses the node IP (34.139.142.168) directly. `claim:archive-node-migration-plan` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_23d69a418ba088f83409a62660832c85` `chunk_id=srcchunk_813dbcaf6fa29529c7a97ab2221e6527` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560386.047159` `source_timestamp=2026-01-28T00:33:06Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
- IP registration service currently uses the archive node directly by design and needs to be updated to use a domain. `claim:ip-registration-uses-archive-node` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_3c95e7d3c20cc8547b7bc496e31b3b86` `chunk_id=srcchunk_87930fff0fa56c263b351f5c66effa56` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560495.293529` `source_timestamp=2026-01-28T00:34:55Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_6004e56f5d083a8ff6a1124f3185dde0` `chunk_id=srcchunk_3d9d3f558ddc433e5736655c3fe5b594` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560545.583849` `source_timestamp=2026-01-28T00:35:45Z`
- Poseidon's depin-api service already uses domain names for staging and production, with staging using https://internal-full.aeneid.storyrpc.io/ and production using https://internal-full.storyrpc.io. `claim:depin-api-uses-domains` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_b14b87d7915d702b72e8a377c2ad87ba` `chunk_id=srcchunk_c96c6d81cef6b86a7866f4a901253a0e` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565546.498389` `source_timestamp=2026-01-28T01:59:06Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- Secrets have been migrated from Bank Vault to AWS Secrets Manager, and no usage of the old IP was found in AWS secrets. `claim:secrets-migrated-to-aws` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_069d1d98b9583041c100720ae1c3ad4e` `chunk_id=srcchunk_3bb1fb817821b349afd3ab642dd5d770` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565604.067249` `source_timestamp=2026-01-28T02:00:24Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2b3824164e3659ecced604e47cf1eed3` `chunk_id=srcchunk_2a5d1c61e56dca62789b9846e2357229` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565691.322389` `source_timestamp=2026-01-28T02:01:31Z`
- All archive nodes are being migrated to AWS and the old GCP nodes will be deprecated. `claim:archive-node-migration-aws` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
- Access to the new AWS servers may require IP whitelisting. `claim:aws-server-access-whitelisting` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_5c4092d9452e21cf002ad64189d61cb0` `chunk_id=srcchunk_74a030c936cb75386f4465497f7eaaee` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560890.600279` `source_timestamp=2026-01-28T00:41:30Z`
- The team confirmed that using domain names instead of IP addresses should ensure no impact from the migration. `claim:domain-usage-confirmation` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_6f048534389c4ddf7c3045edce4737dd` `chunk_id=srcchunk_06f0a88f8d234976b4cbaecbea3b52c2` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565367.504729` `source_timestamp=2026-01-28T01:56:07Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_9eb06240fef156c6161515e0acedff34` `chunk_id=srcchunk_eda3af3429788adea6b36a8258584fd5` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565445.635889` `source_timestamp=2026-01-28T01:57:25Z`

## Open Questions

- Are there any other services beyond IP registration that directly use the old archive node IP?
- When will the old GCP server be fully decommissioned?

## Sources

- `source_document_id`: `srcdoc_83db2545a88ec9b2fc8fa27e705a93e3`
- `source_revision_id`: `srcrev_b14b87d7915d702b72e8a377c2ad87ba`
