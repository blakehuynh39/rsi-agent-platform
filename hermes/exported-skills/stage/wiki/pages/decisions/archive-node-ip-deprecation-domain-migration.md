---
title: "Archive Node IP Deprecation and Domain Migration"
type: "decision"
slug: "decisions/archive-node-ip-deprecation-domain-migration"
freshness: "2026-01-28T02:01:31Z"
tags:
  - "archive-node"
  - "aws"
  - "depin-api"
  - "domain"
  - "gcp"
  - "ip-deprecation"
  - "migration"
  - "poseidon"
  - "storyrpc"
owners: []
source_revision_ids:
  - "srcrev_069d1d98b9583041c100720ae1c3ad4e"
  - "srcrev_173502db147e7c749e821b57c872bcda"
  - "srcrev_2b3824164e3659ecced604e47cf1eed3"
  - "srcrev_3c95e7d3c20cc8547b7bc496e31b3b86"
  - "srcrev_6004e56f5d083a8ff6a1124f3185dde0"
  - "srcrev_8872d38cf73b0c69424977724763295f"
  - "srcrev_89d68928cd0fd070507d91b5616f5c92"
  - "srcrev_9eb06240fef156c6161515e0acedff34"
  - "srcrev_eec4e851479bdab1dae471e9a3cd9c46"
conflict_state: "none"
---

# Archive Node IP Deprecation and Domain Migration

## Summary

Story is migrating archive nodes from GCP to AWS, deprecating the hardcoded IP 34.139.142.168. All dependent services must switch to using domain names. Existing services like depin-api already use domains; IP registration still uses the archive node by design and needs updating. Bank vault secrets are already migrated to AWS Secrets Manager.

## Claims

- All archive nodes are being migrated from GCP to AWS, and the old GCP IP 34.139.142.168 will be deprecated. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_89d68928cd0fd070507d91b5616f5c92` `chunk_id=srcchunk_04a9813c6f26b3e12c7a78201b02f926` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565190.490979` `source_timestamp=2026-01-28T01:53:10Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_9eb06240fef156c6161515e0acedff34` `chunk_id=srcchunk_eda3af3429788adea6b36a8258584fd5` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565445.635889` `source_timestamp=2026-01-28T01:57:25Z`
- Services should use domain names instead of hardcoded IP addresses to avoid disruption during migration. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_6004e56f5d083a8ff6a1124f3185dde0` `chunk_id=srcchunk_3d9d3f558ddc433e5736655c3fe5b594` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560545.583849` `source_timestamp=2026-01-28T00:35:45Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_173502db147e7c749e821b57c872bcda` `chunk_id=srcchunk_709a4da024e4b9beee949901957802d1` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565263.872619` `source_timestamp=2026-01-28T01:54:23Z`
- depin-api staging uses internal-full.aeneid.storyrpc.io and production uses internal-full.storyrpc.io, already on domains. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_eec4e851479bdab1dae471e9a3cd9c46` `chunk_id=srcchunk_535c204ee2138e749c6b1d43733df408` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565655.905229` `source_timestamp=2026-01-28T02:00:55Z`
- Bank vault is no longer used; its secrets have been migrated to AWS Secrets Manager. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_069d1d98b9583041c100720ae1c3ad4e` `chunk_id=srcchunk_3bb1fb817821b349afd3ab642dd5d770` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565604.067249` `source_timestamp=2026-01-28T02:00:24Z`
- IP registration still uses the archive node by design; it needs to be updated to use domain, with @U080YAW205V assigned to work on this. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_3c95e7d3c20cc8547b7bc496e31b3b86` `chunk_id=srcchunk_87930fff0fa56c263b351f5c66effa56` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560495.293529` `source_timestamp=2026-01-28T00:34:55Z`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_8872d38cf73b0c69424977724763295f` `chunk_id=srcchunk_3f5588523759603df0699a914d8b470b` `native_locator=slack:C0547N89JUB:1769560386.047159:1769560665.172459` `source_timestamp=2026-01-28T00:37:45Z`
- No usage of 34.139.142.168 was found in AWS secrets. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_83db2545a88ec9b2fc8fa27e705a93e3` `source_revision_id=srcrev_2b3824164e3659ecced604e47cf1eed3` `chunk_id=srcchunk_2a5d1c61e56dca62789b9846e2357229` `native_locator=slack:C0547N89JUB:1769560386.047159:1769565691.322389` `source_timestamp=2026-01-28T02:01:31Z`

## Open Questions

- Are there any other services or configurations still using the old archive node IP 34.139.142.168?
- Has IP registration been updated to use a domain name instead of the archive node IP?

## Related Pages

- `bank-vault`
- `depin-api`
- `ip-registration-service`
- `poseidon-side-apps`
- `story-archive-nodes`

## Sources

- `source_document_id`: `srcdoc_83db2545a88ec9b2fc8fa27e705a93e3`
- `source_revision_id`: `srcrev_23d69a418ba088f83409a62660832c85`
