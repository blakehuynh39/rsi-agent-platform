---
title: "AWS Credential Management Policy"
type: "policy"
slug: "policies/aws-credential-management-policy"
freshness: "2026-05-05T23:11:26Z"
tags:
  - "aws"
  - "credentials"
  - "security"
  - "sts"
owners:
  - "U0772SH7BRA"
source_revision_ids:
  - "srcrev_5e22831899f2340df39c4c0348d14f91"
conflict_state: "none"
---

# AWS Credential Management Policy

## Summary

Policy prohibiting static AWS credentials for programmatic access; requires use of temporary STS credentials via aws-vault.

## Claims

- Creating static AWS credentials for programmatic access is not allowed. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_5e22831899f2340df39c4c0348d14f91` `chunk_id=srcchunk_531f1af26795d60a10aad542886dc48b` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778022686.791149` `source_timestamp=2026-05-05T23:11:26Z`
- For temporary credentials, aws-vault (https://github.com/99designs/aws-vault) can be used to inject AWS STS credentials at runtime. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_5e22831899f2340df39c4c0348d14f91` `chunk_id=srcchunk_531f1af26795d60a10aad542886dc48b` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778022686.791149` `source_timestamp=2026-05-05T23:11:26Z`

## Related Pages

- `fraud-analysis-s3-access-decision`

## Sources

- `source_document_id`: `srcdoc_80eadc32d0e629997819db8d81b98cee`
- `source_revision_id`: `srcrev_c8698f2b9ba021079acef3b34683272e`
