---
title: "Decision: Grant S3 Read Access for Fraud Analysis"
type: "decision"
slug: "decisions/fraud-analysis-s3-access-decision"
freshness: "2026-05-05T23:15:59Z"
tags:
  - "access-control"
  - "aws"
  - "fraud-detection"
  - "s3"
  - "terraform"
owners:
  - "U067QP5PD6J"
  - "U0772SH7BRA"
source_revision_ids:
  - "srcrev_089a2dfade7d0e8973f72a195c1eceb4"
  - "srcrev_401235a323f1fd0fee05139bc5a6b89c"
  - "srcrev_42ba1d42caee56bb509ebb4ec5b5939f"
  - "srcrev_5141512f012f1774f165e15ec96edd61"
  - "srcrev_55fb3ce90f29952b7aed81c3e521464f"
  - "srcrev_5e22831899f2340df39c4c0348d14f91"
  - "srcrev_6088526bdddb8613431d585bef494c6d"
  - "srcrev_63d7c032b238911b635213ddc7e320a6"
  - "srcrev_7ba1bd42f0635b7a7a0c4f78652b0661"
  - "srcrev_8075dd7f5c41e2c3d22b6583c8907b85"
  - "srcrev_c8698f2b9ba021079acef3b34683272e"
  - "srcrev_e5bc2fa5eba65f73d68a5c225c3d2560"
  - "srcrev_f82df6f6c8737a180476cadfe89741e6"
conflict_state: "none"
---

# Decision: Grant S3 Read Access for Fraud Analysis

## Summary

Decision to grant read-only S3 access to the media bucket for fraud analysis, implemented via Terraform PR #71, and requiring STS temporary credentials.

## Claims

- U067QP5PD6J requested access to the media S3 bucket for programmatically reading data to run fraud analysis locally. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_401235a323f1fd0fee05139bc5a6b89c` `chunk_id=srcchunk_f7ea4fae7b11ca2696331a2f91807552` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778011511.186209` `source_timestamp=2026-05-05T20:05:11Z`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_5141512f012f1774f165e15ec96edd61` `chunk_id=srcchunk_72dd311400e94f620d1ee9b1ba83339c` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778011525.508249` `source_timestamp=2026-05-05T20:05:25Z`
- The access requested is read-only (only read actions). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_63d7c032b238911b635213ddc7e320a6` `chunk_id=srcchunk_25462d0b2c39a2d372de6952402f82e3` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778012867.220449` `source_timestamp=2026-05-05T20:27:47Z`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_7ba1bd42f0635b7a7a0c4f78652b0661` `chunk_id=srcchunk_5d619406564f148f3e6659599a18910b` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778013010.632069` `source_timestamp=2026-05-05T20:30:10Z`
- A pull request (storyprotocol/story-infra-aws#71) was created for the access grant. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_f82df6f6c8737a180476cadfe89741e6` `chunk_id=srcchunk_de158e420d2dde6946c853237ec03d0d` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778011352.919529` `source_timestamp=2026-05-05T20:02:32Z`
- The reviewer (U0772SH7BRA) approved the PR as read-only. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_63d7c032b238911b635213ddc7e320a6` `chunk_id=srcchunk_25462d0b2c39a2d372de6952402f82e3` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778012867.220449` `source_timestamp=2026-05-05T20:27:47Z`
- Terraform plan and apply initially did not run, but U0772SH7BRA ran and applied it successfully. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_6088526bdddb8613431d585bef494c6d` `chunk_id=srcchunk_cabc21fb4d5d1abb1ec2e8f93e171dd4` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778013740.209739` `source_timestamp=2026-05-05T20:42:20Z`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_55fb3ce90f29952b7aed81c3e521464f` `chunk_id=srcchunk_030033209e52bb33ffb35c52f3988603` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778014292.812779` `source_timestamp=2026-05-05T20:51:32Z`
- U067QP5PD6J was asked to check access after apply. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_089a2dfade7d0e8973f72a195c1eceb4` `chunk_id=srcchunk_c75016f2af4c00586e3e875783820de7` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778014296.875699` `source_timestamp=2026-05-05T20:51:36Z`
- U067QP5PD6J asked how to create access keys for a Python script. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_8075dd7f5c41e2c3d22b6583c8907b85` `chunk_id=srcchunk_3514b3b43bc55076d1f977e752a570cb` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778022394.631879` `source_timestamp=2026-05-05T23:06:34Z`
- U0772SH7BRA stated that static AWS credentials are not allowed, and recommended using aws-vault to inject temporary STS credentials at runtime. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_5e22831899f2340df39c4c0348d14f91` `chunk_id=srcchunk_531f1af26795d60a10aad542886dc48b` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778022686.791149` `source_timestamp=2026-05-05T23:11:26Z`
- U067QP5PD6J expressed disappointment. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_e5bc2fa5eba65f73d68a5c225c3d2560` `chunk_id=srcchunk_fcfd0abf90f99b3755dadb147d53900c` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778022712.776439` `source_timestamp=2026-05-05T23:11:52Z`
- U067QP5PD6J attempted to use AWS SSO to login to the piplab AWS account but it did not work. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_42ba1d42caee56bb509ebb4ec5b5939f` `chunk_id=srcchunk_1b8cc45cb2284bfbe5422da1b8373fc2` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778022908.805049` `source_timestamp=2026-05-05T23:15:08Z`
- U08332YRB7W offered to help via a quick huddle. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_80eadc32d0e629997819db8d81b98cee` `source_revision_id=srcrev_c8698f2b9ba021079acef3b34683272e` `chunk_id=srcchunk_fe41ec452d233aafbf6d8fafea4a8968` `native_locator=slack:C0AL7EKNHDF:1778011352.919529:1778022959.405489` `source_timestamp=2026-05-05T23:15:59Z`

## Open Questions

- How was the AWS SSO login issue resolved? Awaiting huddle outcome.

## Related Pages

- `aws-credential-management-policy`

## Sources

- `source_document_id`: `srcdoc_80eadc32d0e629997819db8d81b98cee`
- `source_revision_id`: `srcrev_c8698f2b9ba021079acef3b34683272e`
